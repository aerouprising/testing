package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/joho/godotenv"
	"golang.org/x/text/unicode/norm"
)

var captainsMap map[string]Captain

type ListingCache struct {
	mutex    *sync.Mutex
	listings map[string]Listing
}

var listings = ListingCache{
	mutex:    &sync.Mutex{},
	listings: make(map[string]Listing),
}

type ItemCache struct {
	mutex *sync.Mutex
	items map[string]Item
	names map[string][]string
}

var itemCache = ItemCache{
	mutex: &sync.Mutex{},
	items: make(map[string]Item),
	names: make(map[string][]string),
}

func RemoveAccents(input string) string {
	t := norm.NFD.String(input)
	result := make([]rune, 0, len(t))
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) { // Mn: nonspacing marks
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

func loadListings() {
	listingsAPIURL := os.Getenv("MLB_LISTINGS_URL")

	// Temporary data structure to hold new listings
	tempListings := make(map[string]Listing)

	// Load all listings
	p := 1
	for {
		result := ListingResponse{}
		func() {
			log.Println("Starting to load listings")
			defer func() {
				if err := recover(); err != nil {
					log.Println("Recovering from error:", err)
				}
			}()
			req, _ := http.NewRequest("GET", listingsAPIURL, nil)
			q := req.URL.Query()
			q.Add("page", strconv.Itoa(p))
			req.URL.RawQuery = q.Encode()

			log.Println("URL:", req.URL.String())

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("Error getting listings:", err)
				return
			}
			defer resp.Body.Close()
			log.Println("Got listings page:", p)

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				log.Println("Error decoding response:", err)
				return
			}
			log.Println("Result:", result.Page, result.TotalPages)
			for _, l := range result.Listings {
				tempListings[l.Item.UUID] = l
			}
		}()
		p = result.Page + 1
		if p > result.TotalPages {
			break
		}
	}

	// Replace the old listings with the new ones
	listings.mutex.Lock()
	listings.listings = tempListings
	listings.mutex.Unlock()
}

func loadItems() {
	itemsAPIURL := "https://mlb24.theshow.com/apis/items.json"

	// Temporary data structures to hold new items
	tempItems := make(map[string]Item)
	tempNames := make(map[string][]string)

	// Load all items
	itemPage := 1
	for {
		itemResult := map[string]interface{}{}
		func() {
			// Log only essential information
			log.Println("Starting to load items")
			defer func() {
				if err := recover(); err != nil {
					log.Println("Recovering from error:", err)
				}
			}()
			req, _ := http.NewRequest("GET", itemsAPIURL, nil)
			q := req.URL.Query()
			q.Add("page", strconv.Itoa(itemPage))
			req.URL.RawQuery = q.Encode()

			// Log the URL for debugging purposes
			log.Println("URL:", req.URL.String())

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("Error getting items:", err)
				return
			}
			defer resp.Body.Close()
			log.Println("Got items page:", itemPage)

			if err := json.NewDecoder(resp.Body).Decode(&itemResult); err != nil {
				log.Println("Error decoding item response:", err)
				return
			}
			log.Println("Processed items page:", itemPage)

			items, itemsExist := itemResult["items"].([]interface{})
			if !itemsExist || len(items) == 0 {
				log.Println("No items found in API response")
				return
			}

			itemCache.mutex.Lock()
			defer itemCache.mutex.Unlock()
			for _, i := range items {
				item, itemOk := i.(map[string]interface{})
				if !itemOk {
					continue
				}

				uuid, uuidOk := item["uuid"].(string)
				name, nameOk := item["name"].(string)
				ovr, ovrOk := item["ovr"].(float64)
				contactLeft, _ := item["contact_left"].(float64)
				contactRight, _ := item["contact_right"].(float64)
				hitsPer9, _ := item["hits_per_bf"].(float64)
				walksPer9, _ := item["bb_per_bf"].(float64)
				powerLeft, _ := item["power_left"].(float64)
				powerRight, _ := item["power_right"].(float64)
				speed, _ := item["speed"].(float64)
				vision, _ := item["plate_vision"].(float64)
				clutch, _ := item["batting_clutch"].(float64)
				pitchingClutch, _ := item["pitching_clutch"].(float64)
				fielding, _ := item["fielding_ability"].(float64)
				arm, _ := item["arm_strength"].(float64)
				armAccuracy, _ := item["arm_accuracy"].(float64)
				reaction, _ := item["reaction_time"].(float64)
				strikeoutsPer9, _ := item["k_per_bf"].(float64)
				if uuidOk && nameOk && ovrOk {
					itemDetails := Item{
						UUID:           uuid,
						Name:           name,
						Ovr:            int(ovr),
						ContactLeft:    int(contactLeft),
						ContactRight:   int(contactRight),
						HitsPer9:       int(hitsPer9),
						WalksPer9:      int(walksPer9),
						PowerLeft:      int(powerLeft),
						PowerRight:     int(powerRight),
						Speed:          int(speed),
						Vision:         int(vision),
						Clutch:         int(clutch),
						PitchingClutch: int(pitchingClutch),
						Fielding:       int(fielding),
						Arm:            int(arm),
						ArmAccuracy:    int(armAccuracy),
						Reaction:       int(reaction),
						StrikeoutsPer9: int(strikeoutsPer9),
					}
					tempItems[uuid] = itemDetails
					normalizedName := strings.ToLower(RemoveAccents(name))
					tempNames[normalizedName] = append(tempNames[normalizedName], uuid)
				}
			}
		}()
		itemPage = int(itemResult["page"].(float64)) + 1
		if itemPage > int(itemResult["total_pages"].(float64)) {
			break
		}
	}

	// Replace the old items with the new ones
	itemCache.mutex.Lock()
	itemCache.items = tempItems
	itemCache.names = tempNames
	itemCache.mutex.Unlock()
	log.Println("Item data loaded successfully")
}

func main() {
	http.DefaultClient.Timeout = time.Second * 15

	godotenv.Load()
	// Start the initial load and set up tickers for refreshing data
	go func() {
		loadListings() // Initial load for listings
		listingsTicker := time.NewTicker(60 * time.Second)
		defer listingsTicker.Stop()
		for range listingsTicker.C {
			loadListings()
		}
	}()

	go func() {
		loadItems() // Initial load for items
		itemsTicker := time.NewTicker(30 * time.Minute)
		defer itemsTicker.Stop()
		for range itemsTicker.C {
			loadItems()
		}
	}()

	c := twitch.NewClient("joshq00", os.Getenv("TWITCH_OAUTH_TOKEN"))
	c.Join(strings.Split(os.Getenv("TWITCH_CHANNELS"), ",")...)
	c.OnPrivateMessage(func(m twitch.PrivateMessage) {
		log.Println(m.Channel, m.User.DisplayName, m.Message)
		if strings.HasPrefix(m.Message, "!price ") {
			playerName := strings.TrimPrefix(m.Message, "!price ")
			cards := findCard(playerName)

			if strings.Contains(playerName, " ") {
				for _, item := range cards {
					c.Say(m.Channel,
						fmt.Sprintf("[PRICE] %s (%v) | %s %v | Buy now: %v | Sell now: %v\n", item.Item.Name, item.Item.Ovr, item.Item.Team, item.Item.Rarity, item.BestSellPrice, item.BestBuyPrice),
					)
				}

				if len(cards) == 0 {
					c.Say(m.Channel,
						"Player Card not found")
				}
			}
		} else if strings.HasPrefix(m.Message, "!contact ") {
			playerName := strings.TrimPrefix(m.Message, "!contact ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[CONTACT] %s (%v) | Contact vs Left: %v | Contact vs Right: %v", item.Name, item.Ovr, item.ContactLeft, item.ContactRight))
				}
			} else {
				c.Say(m.Channel, "Contact information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!hitsper9 ") {
			playerName := strings.TrimPrefix(m.Message, "!hitsper9 ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[HITS PER 9] %s (%v) | Hits per 9: %v", item.Name, item.Ovr, item.HitsPer9))
				}
			} else {
				c.Say(m.Channel, "Hits per 9 information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!walksper9 ") {
			playerName := strings.TrimPrefix(m.Message, "!walksper9 ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[WALKS PER 9] %s (%v) | Walks per 9: %v", item.Name, item.Ovr, item.WalksPer9))
				}
			} else {
				c.Say(m.Channel, "Walks per 9 information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!power ") {
			playerName := strings.TrimPrefix(m.Message, "!power ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[POWER] %s (%v) | Power vs Left: %v | Power vs Right: %v", item.Name, item.Ovr, item.PowerLeft, item.PowerRight))
				}
			} else {
				c.Say(m.Channel, "Power information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!speed ") {
			playerName := strings.TrimPrefix(m.Message, "!speed ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[SPEED] %s (%v) | Speed: %v", item.Name, item.Ovr, item.Speed))
				}
			} else {
				c.Say(m.Channel, "Speed information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!vision ") {
			playerName := strings.TrimPrefix(m.Message, "!vision ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[VISION] %s (%v) | Vision: %v", item.Name, item.Ovr, item.Vision))
				}
			} else {
				c.Say(m.Channel, "Vision information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!hittingclutch ") {
			playerName := strings.TrimPrefix(m.Message, "!clutch ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[CLUTCH] %s (%v) | Clutch: %v", item.Name, item.Ovr, item.Clutch))
				}
			} else {
				c.Say(m.Channel, "Clutch information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!pitchingclutch ") {
			playerName := strings.TrimPrefix(m.Message, "!pitchingclutch ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[PITCHING CLUTCH] %s (%v) | Pitching Clutch: %v", item.Name, item.Ovr, item.PitchingClutch))
				}
			} else {
				c.Say(m.Channel, "Pitching Clutch information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!fielding ") {
			playerName := strings.TrimPrefix(m.Message, "!fielding ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[FIELDING] %s (%v) | Fielding: %v", item.Name, item.Ovr, item.Fielding))
				}
			} else {
				c.Say(m.Channel, "Fielding information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!arm ") {
			playerName := strings.TrimPrefix(m.Message, "!arm ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[ARM] %s (%v) | Arm: %v", item.Name, item.Ovr, item.Arm))
				}
			} else {
				c.Say(m.Channel, "Arm information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!armaccuracy ") {
			playerName := strings.TrimPrefix(m.Message, "!armaccuracy ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[ARM ACCURACY] %s (%v) | Arm Accuracy: %v", item.Name, item.Ovr, item.ArmAccuracy))
				}
			} else {
				c.Say(m.Channel, "Arm Accuracy information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!reaction ") {
			playerName := strings.TrimPrefix(m.Message, "!reaction ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[REACTION] %s (%v) | Reaction: %v", item.Name, item.Ovr, item.Reaction))
				}
			} else {
				c.Say(m.Channel, "Reaction information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!strikeoutsper9 ") {
			playerName := strings.TrimPrefix(m.Message, "!strikeoutsper9 ")
			contactItems := getPlayerContacts(playerName)
			if contactItems != nil {
				for _, item := range contactItems {
					c.Say(m.Channel,
						fmt.Sprintf("[STRIKEOUTS PER 9] %s (%v) | Strikeouts per 9: %v", item.Name, item.Ovr, item.StrikeoutsPer9))
				}
			} else {
				c.Say(m.Channel, "Strikeouts per 9 information not available.")
			}
		} else if strings.HasPrefix(m.Message, "!theme ") {
			playerName := strings.TrimPrefix(m.Message, "!theme ")
			eligibleCaptains := findEligibleCaptains(playerName)
			if len(eligibleCaptains) > 0 {
				for _, captain := range eligibleCaptains {
					c.Say(m.Channel,
						fmt.Sprintf("[THEME] %s is eligible for Captain %s: %s", playerName, captain.Name, captain.AbilityDesc))
				}
			} else {
				c.Say(m.Channel, "No matching themes found for the player.")
			}
		}
	})
	c.Connect()
}

func getPlayerContacts(name string) []Item {
	itemCache.mutex.Lock()
	defer itemCache.mutex.Unlock()

	normalizedName := strings.ToLower(RemoveAccents(name))
	uuids, exists := itemCache.names[normalizedName]
	if !exists {
		return nil
	}

	var items []Item
	for _, uuid := range uuids {
		item, exists := itemCache.items[uuid]
		if exists {
			items = append(items, item)
		}
	}
	return items
}

func findCard(playerName string) []Listing {
	listings.mutex.Lock()
	defer listings.mutex.Unlock()

	var result []Listing
	normalizedPlayerName := strings.ToLower(RemoveAccents(playerName))
	for _, listing := range listings.listings { // Access the map of listings
		if strings.EqualFold(strings.ToLower(RemoveAccents(listing.Item.Name)), normalizedPlayerName) {
			result = append(result, listing)
		}
	}
	return result
}

func findEligibleCaptains(playerName string) []Captain {
	itemCache.mutex.Lock()
	defer itemCache.mutex.Unlock()

	normalizedName := strings.ToLower(RemoveAccents(playerName))
	uuids, exists := itemCache.names[normalizedName]
	if !exists {
		return nil
	}

	var eligibleCaptains []Captain
	for _, uuid := range uuids {
		item, exists := itemCache.items[uuid]
		if exists {
			for _, captain := range captainsMap {
				if isPlayerEligibleForCaptain(item, captain) {
					eligibleCaptains = append(eligibleCaptains, captain)
				}
			}
		}
	}
	return eligibleCaptains
}

func isPlayerEligibleForCaptain(item Item, captain Captain) bool {
	description := strings.ToLower(captain.AbilityDesc)

	if strings.Contains(description, "switch hitter") && item.BatHand == "S" {
		return true
	}
	if strings.Contains(description, "new york yankees") && item.Team == "New York Yankees" {
		return true
	}
	// will be adding more later once this works

	return false
}

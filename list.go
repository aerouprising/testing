package main

type ListingResponse struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
	Listings   []Listing `json:"listings"`
}

type Item struct {
	UUID              string `json:"uuid"`
	Type              string `json:"type"`
	Img               string `json:"img"`
	BakedImg          string `json:"baked_img"`
	ScBakedImg        any    `json:"sc_baked_img"`
	Name              string `json:"name"`
	Rarity            string `json:"rarity"`
	Team              string `json:"team"`
	TeamShortName     string `json:"team_short_name"`
	Ovr               int    `json:"ovr"`
	Series            string `json:"series"`
	SeriesTextureName string `json:"series_texture_name"`
	SeriesYear        int    `json:"series_year"`
	DisplayPosition   string `json:"display_position"`
	HasAugment        bool   `json:"has_augment"`
	AugmentText       any    `json:"augment_text"`
	AugmentEndDate    any    `json:"augment_end_date"`
	HasMatchup        bool   `json:"has_matchup"`
	Stars             any    `json:"stars"`
	Trend             any    `json:"trend"`
	NewRank           int    `json:"new_rank"`
	HasRankChange     bool   `json:"has_rank_change"`
	Event             bool   `json:"event"`
	SetName           string `json:"set_name"`
	IsLiveSet         bool   `json:"is_live_set"`
	UIAnimIndex       int    `json:"ui_anim_index"`
	ContactLeft       int    `json:"contact_left"`
	ContactRight      int    `json:"contact_right"`
	HitsPer9          int    `json:"hits_per_bf"`
	WalksPer9         int    `json:"bb_per_bf"`
	PowerLeft         int    `json:"power_left"`
	PowerRight        int    `json:"power_right"`
	Speed             int    `json:"speed"`
	Vision            int    `json:"plate_vision"`
	Clutch            int    `json:"batting_clutch"`
	PitchingClutch    int    `json:"pitching_clutch"`
	Fielding          int    `json:"fielding_ability"`
	Arm               int    `json:"arm_strength"`
	ArmAccuracy       int    `json:"arm_accuracy"`
	Reaction          int    `json:"reaction_time"`
	StrikeoutsPer9    int    `json:"k_per_bf"`
	BatHand           string `json:"bat_hand"`
	ThrowHand         string `json:"throw_hand"`
	IsHitter          string `json:"is_hitter"`
}

type Listing struct {
	ListingName   string `json:"listing_name"`
	BestSellPrice int    `json:"best_sell_price"`
	BestBuyPrice  int    `json:"best_buy_price"`
	Item          Item   `json:"item"`
}
type Captain struct {
	UUID            string         `json:"uuid"`
	Name            string         `json:"name"`
	DisplayPosition string         `json:"display_position"`
	Team            string         `json:"team"`
	Ovr             int            `json:"ovr"`
	AbilityName     string         `json:"ability_name"`
	AbilityDesc     string         `json:"ability_desc"`
	Boosts          []CaptainBoost `json:"boosts"`
}

type CaptainBoost struct {
	Tier        string           `json:"tier"`
	Description string           `json:"description"`
	Attributes  []BoostAttribute `json:"attributes"`
}

type BoostAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Player struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name"`
	DisplayPosition string `json:"display_position"`
	Team            string `json:"team"`
	BatHand         string `json:"bat_hand"`
	ThrowHand       string `json:"throw_hand"`
	Country         string `json:"country"`
	IsHitter        string `json:"is_hitter"`
	PowerVsL        int    `json:"power_left"`
	PowerVsR        int    `json:"power_right"`
	Speed           int    `json:"speed"`
	Vision          int    `json:"plate_vision"`
}
type CaptainResponse struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
	Captains   []Captain `json:"captains"`
}

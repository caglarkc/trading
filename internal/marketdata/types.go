package marketdata

import "time"

type Asset struct {
	ID       string            `json:"id"`
	Symbol   string            `json:"symbol"`
	Name     string            `json:"name"`
	Category string            `json:"category"`
	Exchange string            `json:"exchange"`
	Aliases  map[string]string `json:"aliases"`
}

type Observation struct {
	MinuteVolume     *int64     `json:"minute_volume_shares,omitempty"`
	VolumeMinute     *time.Time `json:"volume_minute,omitempty"`
	VolumeScope      string     `json:"volume_scope,omitempty"`
	ID               string     `json:"id"`
	AssetID          string     `json:"asset_id"`
	Symbol           string     `json:"symbol"`
	SourceID         string     `json:"source_id"`
	SourceURL        string     `json:"source_url"`
	Method           string     `json:"method"`
	IdentityVerified bool       `json:"identity_verified"`
	Currency         string     `json:"currency,omitempty"`
	PriceType        string     `json:"price_type"`
	Session          string     `json:"session"`
	EventTime        *time.Time `json:"event_time"`
	EventTimeText    string     `json:"event_time_text,omitempty"`
	ObservedAt       time.Time  `json:"observed_at"`
	Price            *string    `json:"price"`
	Bid              *string    `json:"bid"`
	Ask              *string    `json:"ask"`
	BidSize          *int64     `json:"bid_size"`
	AskSize          *int64     `json:"ask_size"`
	BBOTime          *time.Time `json:"bbo_time"`
	SizeUnit         string     `json:"size_unit,omitempty"`
	VenueScope       string     `json:"venue_scope,omitempty"`
	Delayed          bool       `json:"delayed"`
	ParserVersion    string     `json:"parser_version"`
	Error            string     `json:"error,omitempty"`
}

type Validation struct {
	Verified     bool      `json:"verified"`
	Reason       string    `json:"reason"`
	Median       *string   `json:"median"`
	RangePercent *string   `json:"range_percent"`
	Observations []string  `json:"observations"`
	AsOf         time.Time `json:"as_of"`
}

type Snapshot struct {
	Asset             Asset         `json:"asset"`
	Observations      []Observation `json:"observations"`
	Validation        Validation    `json:"validation"`
	LiquidityVerified bool          `json:"liquidity_verified"`
}

type Page struct {
	PublishedAt          *time.Time
	Text, Heading, Title string
	RegularPrice         *string
	Method               string
}

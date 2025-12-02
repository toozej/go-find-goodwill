package api

import (
	"time"
)

// ItemResponse represents an item response
type ItemResponse struct {
	ID              int        `json:"id"`
	GoodwillID      string     `json:"goodwill_id"`
	Title           string     `json:"title"`
	Seller          string     `json:"seller"`
	CurrentPrice    float64    `json:"current_price"`
	BuyNowPrice     *float64   `json:"buy_now_price"`
	URL             string     `json:"url"`
	ImageURL        string     `json:"image_url"`
	EndsAt          *time.Time `json:"ends_at"`
	Status          string     `json:"status"`
	Category        string     `json:"category"`
	Subcategory     string     `json:"subcategory"`
	Condition       string     `json:"condition"`
	ShippingCost    *float64   `json:"shipping_cost"`
	ShippingMethod  string     `json:"shipping_method"`
	Description     string     `json:"description"`
	Location        string     `json:"location"`
	PickupAvailable bool       `json:"pickup_available"`
	ReturnsAccepted bool       `json:"returns_accepted"`
	WatchCount      int        `json:"watch_count"`
	BidCount        int        `json:"bid_count"`
	ViewCount       int        `json:"view_count"`
	FirstSeen       time.Time  `json:"first_seen"`
	LastSeen        time.Time  `json:"last_seen"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ItemListResponse represents a list of items with pagination
type ItemListResponse struct {
	Items  []ItemResponse `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// PriceHistoryResponse represents price history for an item
type PriceHistoryResponse struct {
	Price      float64   `json:"price"`
	PriceType  string    `json:"price_type"`
	RecordedAt time.Time `json:"recorded_at"`
}

// BidHistoryResponse represents bid history for an item
type BidHistoryResponse struct {
	BidAmount  float64   `json:"bid_amount"`
	Bidder     string    `json:"bidder"`
	BidderID   string    `json:"bidder_id"`
	RecordedAt time.Time `json:"recorded_at"`
}

// ItemHistoryResponse represents the complete history for an item
type ItemHistoryResponse struct {
	PriceHistory []PriceHistoryResponse `json:"price_history"`
	BidHistory   []BidHistoryResponse   `json:"bid_history"`
}

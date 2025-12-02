package api

import (
	"time"
)

// SearchRequest represents a search creation/update request
type SearchRequest struct {
	Name                      string   `json:"name"`
	Query                     string   `json:"query"`
	RegexPattern              string   `json:"regex_pattern"`
	Enabled                   bool     `json:"enabled"`
	NotificationThresholdDays int      `json:"notification_threshold_days"`
	MinPrice                  *float64 `json:"min_price"`
	MaxPrice                  *float64 `json:"max_price"`
	CategoryFilter            string   `json:"category_filter"`
	SellerFilter              string   `json:"seller_filter"`
	ShippingFilter            string   `json:"shipping_filter"`
	ConditionFilter           string   `json:"condition_filter"`
	SortBy                    string   `json:"sort_by"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	ID                        int        `json:"id"`
	Name                      string     `json:"name"`
	Query                     string     `json:"query"`
	RegexPattern              string     `json:"regex_pattern"`
	Enabled                   bool       `json:"enabled"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
	LastChecked               *time.Time `json:"last_checked"`
	NotificationThresholdDays int        `json:"notification_threshold_days"`
	MinPrice                  *float64   `json:"min_price"`
	MaxPrice                  *float64   `json:"max_price"`
	CategoryFilter            string     `json:"category_filter"`
	SellerFilter              string     `json:"seller_filter"`
	ShippingFilter            string     `json:"shipping_filter"`
	ConditionFilter           string     `json:"condition_filter"`
	SortBy                    string     `json:"sort_by"`
}

// SearchListResponse represents a list of searches with pagination
type SearchListResponse struct {
	Searches []SearchResponse `json:"searches"`
	Total    int              `json:"total"`
	Limit    int              `json:"limit"`
	Offset   int              `json:"offset"`
}

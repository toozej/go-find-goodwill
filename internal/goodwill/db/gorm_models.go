package db

import (
	"time"
)

// GormSearch represents a search query with GORM struct tags
type GormSearch struct {
	ID                        int    `gorm:"primaryKey"`
	Name                      string `gorm:"size:255;not null"`
	Query                     string `gorm:"size:500;not null"`
	RegexPattern              string `gorm:"size:500"`
	Enabled                   bool   `gorm:"default:true"`
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	LastChecked               *time.Time
	NotificationThresholdDays int `gorm:"default:7"`
	MinPrice                  *float64
	MaxPrice                  *float64
	CategoryFilter            string `gorm:"size:100"`
	SellerFilter              string `gorm:"size:100"`
	ShippingFilter            string `gorm:"size:100"`
	ConditionFilter           string `gorm:"size:100"`
	SortBy                    string `gorm:"size:50"`

	// Relationships
	SearchExecutions   []GormSearchExecution   `gorm:"foreignKey:SearchID"`
	SearchItemMappings []GormSearchItemMapping `gorm:"foreignKey:SearchID"`
}

// GormSearchExecution represents search execution history
type GormSearchExecution struct {
	ID            int `gorm:"primaryKey"`
	SearchID      int `gorm:"not null"`
	ExecutedAt    time.Time
	Status        string `gorm:"size:50;not null"`
	ItemsFound    int
	NewItemsFound int
	ErrorMessage  string `gorm:"size:1000"`
	DurationMS    int

	// Relationships
	Search GormSearch `gorm:"foreignKey:SearchID"`
}

// GormItem represents an item from Goodwill listings
type GormItem struct {
	ID              int     `gorm:"primaryKey"`
	GoodwillID      string  `gorm:"size:100;unique;not null"`
	Title           string  `gorm:"size:500;not null"`
	Seller          string  `gorm:"size:100"`
	CurrentPrice    float64 `gorm:"not null;index"`
	BuyNowPrice     *float64
	URL             string `gorm:"size:500;not null"`
	ImageURL        string `gorm:"size:500"`
	EndsAt          *time.Time
	CreatedAt       time.Time `gorm:"index"`
	UpdatedAt       time.Time
	FirstSeen       time.Time
	LastSeen        time.Time `gorm:"index"`
	Status          string    `gorm:"size:50;default:'active';index"`
	Category        string    `gorm:"size:100;index"`
	Subcategory     string    `gorm:"size:100"`
	Condition       string    `gorm:"size:100"`
	ShippingCost    *float64
	ShippingMethod  string `gorm:"size:100"`
	Description     string `gorm:"type:text"`
	Location        string `gorm:"size:200"`
	PickupAvailable bool   `gorm:"default:false"`
	ReturnsAccepted bool   `gorm:"default:false"`
	WatchCount      int    `gorm:"default:0"`
	BidCount        int    `gorm:"default:0"`
	ViewCount       int    `gorm:"default:0"`

	// Merged from GormItemDetails
	Dimensions string `gorm:"size:100"`
	Weight     string `gorm:"size:50"`
	Material   string `gorm:"size:100"`
	Color      string `gorm:"size:50"`
	Brand      string `gorm:"size:100"`
	Model      string `gorm:"size:100"`
	Year       *int

	// Relationships
	PriceHistories     []GormPriceHistory      `gorm:"foreignKey:ItemID"`
	BidHistories       []GormBidHistory        `gorm:"foreignKey:ItemID"`
	Notifications      []GormNotification      `gorm:"foreignKey:ItemID"`
	SearchItemMappings []GormSearchItemMapping `gorm:"foreignKey:ItemID"`
}

// GormPriceHistory represents price history for an item
type GormPriceHistory struct {
	ID         int     `gorm:"primaryKey"`
	ItemID     int     `gorm:"not null"`
	Price      float64 `gorm:"not null"`
	PriceType  string  `gorm:"size:50;not null"`
	RecordedAt time.Time

	// Relationship
	Item GormItem `gorm:"foreignKey:ItemID"`
}

// GormBidHistory represents bid history for an item
type GormBidHistory struct {
	ID         int     `gorm:"primaryKey"`
	ItemID     int     `gorm:"not null"`
	BidAmount  float64 `gorm:"not null"`
	Bidder     string  `gorm:"size:100"`
	BidderID   string  `gorm:"size:100"`
	RecordedAt time.Time

	// Relationship
	Item GormItem `gorm:"foreignKey:ItemID"`
}

// GormNotification represents a notification to be sent
type GormNotification struct {
	ID               int    `gorm:"primaryKey"`
	ItemID           int    `gorm:"not null"`
	SearchID         int    `gorm:"not null"`
	NotificationType string `gorm:"size:100;not null"`
	Status           string `gorm:"size:50;default:'queued';index"`
	SentAt           *time.Time
	DeliveredAt      *time.Time
	ErrorMessage     string `gorm:"size:1000"`
	RetryCount       int    `gorm:"default:0"`
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Relationships
	Item   GormItem   `gorm:"foreignKey:ItemID"`
	Search GormSearch `gorm:"foreignKey:SearchID"`
}

// GormUserAgent represents a user agent for anti-bot measures
type GormUserAgent struct {
	ID         int    `gorm:"primaryKey"`
	UserAgent  string `gorm:"size:500;not null"`
	LastUsed   *time.Time
	UsageCount int  `gorm:"default:0"`
	IsActive   bool `gorm:"default:true"`
}

// GormSystemLog represents system log entries
type GormSystemLog struct {
	ID         int `gorm:"primaryKey"`
	Timestamp  time.Time
	Level      string `gorm:"size:20;not null"`
	Component  string `gorm:"size:100;not null"`
	Message    string `gorm:"size:1000;not null"`
	Details    string `gorm:"type:text"`
	StackTrace string `gorm:"type:text"`
}

// GormSearchItemMapping represents the many-to-many relationship between searches and items
type GormSearchItemMapping struct {
	ID       int `gorm:"primaryKey"`
	SearchID int `gorm:"not null;index"`
	ItemID   int `gorm:"not null;index"`
	FoundAt  time.Time

	// Relationships
	Search GormSearch `gorm:"foreignKey:SearchID"`
	Item   GormItem   `gorm:"foreignKey:ItemID"`
}

// GetAllModels returns all GORM models for AutoMigrate
func GetAllModels() []interface{} {
	return []interface{}{
		&GormSearch{},
		&GormSearchExecution{},
		&GormItem{},
		&GormPriceHistory{},
		&GormBidHistory{},
		&GormNotification{},
		&GormUserAgent{},
		&GormSystemLog{},
		&GormSearchItemMapping{},
		&GormMigration{},
	}
}

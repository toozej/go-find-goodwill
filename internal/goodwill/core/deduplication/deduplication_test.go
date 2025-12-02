package deduplication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
)

func TestGenerateItemFingerprint(t *testing.T) {
	// Create test item
	item := &db.GormItem{
		GoodwillID:   "test-123",
		Title:        "Vintage Camera",
		Seller:       "camera_seller",
		CurrentPrice: 99.99,
		Description:  "A vintage camera in excellent condition",
	}

	// Create deduplication service with nil repository (for testing purposes)
	svc := &DeduplicationService{}

	// Test fingerprint generation
	fingerprint, err := svc.GenerateItemFingerprint(item)
	assert.NoError(t, err)
	assert.NotNil(t, fingerprint)
	assert.NotEmpty(t, fingerprint.TitleFingerprint)
	assert.NotEmpty(t, fingerprint.SellerFingerprint)
	assert.NotEmpty(t, fingerprint.PriceFingerprint)
	assert.NotEmpty(t, fingerprint.DescriptionFingerprint)
	assert.NotEmpty(t, fingerprint.CombinedFingerprint)
}

func TestCalculateSimilarity(t *testing.T) {
	// Create test items
	item1 := &db.GormItem{
		Title:        "Vintage Camera",
		Seller:       "camera_seller",
		CurrentPrice: 99.99,
		Description:  "A vintage camera in excellent condition",
	}

	item2 := &db.GormItem{
		Title:        "Vintage Camera",
		Seller:       "camera_seller",
		CurrentPrice: 99.99,
		Description:  "A vintage camera in excellent condition",
	}

	item3 := &db.GormItem{
		Title:        "Modern Digital Camera",
		Seller:       "electronics_store",
		CurrentPrice: 499.99,
		Description:  "A brand new digital camera",
	}

	// Create deduplication service
	svc := &DeduplicationService{}

	// Test identical items
	similarity, err := svc.CalculateSimilarity(item1, item2)
	assert.NoError(t, err)
	assert.True(t, similarity > 0.9) // Should be very similar

	// Test different items
	similarity, err = svc.CalculateSimilarity(item1, item3)
	assert.NoError(t, err)
	assert.True(t, similarity < 0.5) // Should be quite different
}

func TestDeduplicationStats(t *testing.T) {
	// Create deduplication service
	svc := &DeduplicationService{}

	// Test stats calculation
	stats, err := svc.GetDeduplicationStats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "total_items")
	assert.Contains(t, stats, "active_items")
	assert.Contains(t, stats, "merged_items")
	assert.Contains(t, stats, "duplicate_rate")
}

func TestPriceSimilarity(t *testing.T) {
	// Create test items with different prices
	item1 := &db.GormItem{
		CurrentPrice: 100.00,
	}

	item2 := &db.GormItem{
		CurrentPrice: 100.00, // Same price
	}

	item3 := &db.GormItem{
		CurrentPrice: 110.00, // 10% higher
	}

	item4 := &db.GormItem{
		CurrentPrice: 150.00, // 50% higher
	}

	// Create deduplication service with config
	svc := &DeduplicationService{
		config: &DeduplicationConfig{
			PriceSimilarityThreshold: 0.2, // 20% threshold
		},
	}

	// Test exact same price
	similarity := svc.calculatePriceSimilarity(item1.CurrentPrice, item2.CurrentPrice)
	assert.Equal(t, 1.0, similarity)

	// Test within threshold
	similarity = svc.calculatePriceSimilarity(item1.CurrentPrice, item3.CurrentPrice)
	assert.True(t, similarity > 0.8) // Should be similar but not identical

	// Test outside threshold
	similarity = svc.calculatePriceSimilarity(item1.CurrentPrice, item4.CurrentPrice)
	assert.True(t, similarity < 0.5) // Should be quite different
}

func TestTextNormalization(t *testing.T) {
	svc := &DeduplicationService{}

	// Test text normalization
	testCases := []struct {
		input    string
		expected string
	}{
		{"Vintage Camera", "vintage camera"},
		{"  Vintage  Camera  ", "vintage camera"},
		{"Vintage Camera!", "vintage camera"},
		{"Vintage-Camera_123", "vintagecamera123"},
	}

	for _, tc := range testCases {
		result := svc.normalizeText(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}

func TestFuzzyTextSimilarity(t *testing.T) {
	svc := &DeduplicationService{}

	// Test fuzzy text similarity
	testCases := []struct {
		fp1      string
		fp2      string
		expected float64
	}{
		{"abc123", "abc123", 1.0},    // Exact match
		{"abc123", "abc456", 0.0},    // Partial match (prefix "abc", suffix "3" vs "6") - below threshold
		{"abc123", "xyz789", 0.0},    // No match
		{"abc123", "abc12", 0.45},    // Partial match (prefix "abc12", suffix "3" vs "2")
		{"123abc", "456abc", 0.0},    // Partial match (prefix "1" vs "4", suffix "abc") - below threshold
		{"abc123", "abc123def", 0.4}, // Partial match with longer string
	}

	for _, tc := range testCases {
		result := svc.calculateTextSimilarity(tc.fp1, tc.fp2)
		// Allow some tolerance for floating point comparison
		assert.InDelta(t, tc.expected, result, 0.1, "Expected %v, got %v for %s vs %s", tc.expected, result, tc.fp1, tc.fp2)
	}
}

func TestConfigurableWeights(t *testing.T) {
	// Test with custom weights
	config := &DeduplicationConfig{
		TitleWeight:       0.5, // 50% title
		SellerWeight:      0.1, // 10% seller
		PriceWeight:       0.2, // 20% price
		DescriptionWeight: 0.2, // 20% description
	}

	svc := &DeduplicationService{
		config: config,
	}

	// Verify config is applied
	assert.Equal(t, 0.5, svc.config.TitleWeight)
	assert.Equal(t, 0.1, svc.config.SellerWeight)
	assert.Equal(t, 0.2, svc.config.PriceWeight)
	assert.Equal(t, 0.2, svc.config.DescriptionWeight)
}

func TestDuplicateRateCalculation(t *testing.T) {
	// Test duplicate rate calculation
	rate := calculateDuplicateRate(100, 10)
	assert.Equal(t, 0.1, rate)

	rate = calculateDuplicateRate(50, 0)
	assert.Equal(t, 0.0, rate)

	rate = calculateDuplicateRate(0, 0)
	assert.Equal(t, 0.0, rate)
}

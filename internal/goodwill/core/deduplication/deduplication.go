package deduplication

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
)

var (
	// Package-level compiled regex patterns for performance
	specialCharsRegex = regexp.MustCompile(`[^a-z0-9\s]`)
	whitespaceRegex   = regexp.MustCompile(`\s+`)
)

// ItemFingerprint represents a unique fingerprint for an item
type ItemFingerprint struct {
	TitleFingerprint       string
	SellerFingerprint      string
	PriceFingerprint       string
	DescriptionFingerprint string
	CombinedFingerprint    string
}

// DeduplicationConfig contains configuration for the deduplication system
type DeduplicationConfig struct {
	TitleSimilarityThreshold       float64
	PriceSimilarityThreshold       float64
	DescriptionSimilarityThreshold float64
	OverallSimilarityThreshold     float64
	MaxAgeForComparison            time.Duration
	// Similarity weights for weighted scoring
	TitleWeight       float64
	SellerWeight      float64
	PriceWeight       float64
	DescriptionWeight float64
}

// DeduplicationService handles item deduplication
type DeduplicationService struct {
	config *DeduplicationConfig
	repo   db.Repository
}

// NewDeduplicationService creates a new deduplication service
func NewDeduplicationService(repo db.Repository, config *DeduplicationConfig) *DeduplicationService {
	if config == nil {
		config = &DeduplicationConfig{
			TitleSimilarityThreshold:       0.85,
			PriceSimilarityThreshold:       0.90,
			DescriptionSimilarityThreshold: 0.75,
			OverallSimilarityThreshold:     0.80,
			MaxAgeForComparison:            7 * 24 * time.Hour, // 7 days
			TitleWeight:                    0.4,                // 40%
			SellerWeight:                   0.2,                // 20%
			PriceWeight:                    0.2,                // 20%
			DescriptionWeight:              0.2,                // 20%
		}
	}

	return &DeduplicationService{
		config: config,
		repo:   repo,
	}
}

// GenerateItemFingerprint generates a fingerprint for an item
func (s *DeduplicationService) GenerateItemFingerprint(item *db.GormItem) (*ItemFingerprint, error) {
	if item == nil {
		return nil, fmt.Errorf("item cannot be nil")
	}

	// Generate individual fingerprints
	titleFingerprint := s.generateTextFingerprint(item.Title)
	sellerFingerprint := s.generateTextFingerprint(item.Seller)
	priceFingerprint := s.generatePriceFingerprint(item.CurrentPrice)
	descriptionFingerprint := s.generateTextFingerprint(item.Description)

	// Combine fingerprints for overall fingerprint
	combined := fmt.Sprintf("%s:%s:%s:%s",
		titleFingerprint,
		sellerFingerprint,
		priceFingerprint,
		descriptionFingerprint)

	combinedFingerprint := s.generateHash(combined)

	return &ItemFingerprint{
		TitleFingerprint:       titleFingerprint,
		SellerFingerprint:      sellerFingerprint,
		PriceFingerprint:       priceFingerprint,
		DescriptionFingerprint: descriptionFingerprint,
		CombinedFingerprint:    combinedFingerprint,
	}, nil
}

// generateTextFingerprint generates a fingerprint for text content
func (s *DeduplicationService) generateTextFingerprint(text string) string {
	if text == "" {
		return ""
	}

	// Normalize text: lowercase, remove special chars, trim whitespace
	normalized := s.normalizeText(text)

	// Generate hash of normalized text
	return s.generateHash(normalized)
}

// generatePriceFingerprint generates a fingerprint for price
func (s *DeduplicationService) generatePriceFingerprint(price float64) string {
	// Round to 2 decimal places to avoid floating point precision issues
	roundedPrice := math.Round(price*100) / 100
	return fmt.Sprintf("%.2f", roundedPrice)
}

// generateHash generates a SHA-256 hash of the input string
func (s *DeduplicationService) generateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// normalizeText normalizes text for comparison
func (s *DeduplicationService) normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove special characters and extra whitespace using package-level compiled regex
	text = specialCharsRegex.ReplaceAllString(text, "")

	// Replace multiple spaces with single space using package-level compiled regex
	text = whitespaceRegex.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// CalculateSimilarity calculates similarity between two items
func (s *DeduplicationService) CalculateSimilarity(item1, item2 *db.GormItem) (float64, error) {
	if item1 == nil || item2 == nil {
		return 0.0, fmt.Errorf("items cannot be nil")
	}

	// Generate fingerprints for both items
	fingerprint1, err := s.GenerateItemFingerprint(item1)
	if err != nil {
		return 0.0, fmt.Errorf("failed to generate fingerprint for item1: %w", err)
	}

	fingerprint2, err := s.GenerateItemFingerprint(item2)
	if err != nil {
		return 0.0, fmt.Errorf("failed to generate fingerprint for item2: %w", err)
	}

	// Calculate individual similarity scores
	titleSimilarity := s.calculateTextSimilarity(fingerprint1.TitleFingerprint, fingerprint2.TitleFingerprint)
	sellerSimilarity := s.calculateTextSimilarity(fingerprint1.SellerFingerprint, fingerprint2.SellerFingerprint)
	priceSimilarity := s.calculatePriceSimilarity(item1.CurrentPrice, item2.CurrentPrice)
	descriptionSimilarity := s.calculateTextSimilarity(fingerprint1.DescriptionFingerprint, fingerprint2.DescriptionFingerprint)

	// Calculate weighted overall similarity using configurable weights
	// Ensure config is not nil and has default weights if needed
	if s.config == nil {
		s.config = &DeduplicationConfig{
			TitleWeight:       0.4,
			SellerWeight:      0.2,
			PriceWeight:       0.2,
			DescriptionWeight: 0.2,
		}
	}

	overallSimilarity := (titleSimilarity * s.config.TitleWeight) +
		(sellerSimilarity * s.config.SellerWeight) +
		(priceSimilarity * s.config.PriceWeight) +
		(descriptionSimilarity * s.config.DescriptionWeight)

	return overallSimilarity, nil
}

// calculateTextSimilarity calculates similarity between two text fingerprints using fuzzy matching
func (s *DeduplicationService) calculateTextSimilarity(fp1, fp2 string) float64 {
	if fp1 == "" && fp2 == "" {
		return 1.0 // Both empty - considered identical
	}
	if fp1 == "" || fp2 == "" {
		return 0.0 // One empty, one not - considered different
	}

	// For fingerprints, we use exact match (since they're hashes)
	if fp1 == fp2 {
		return 1.0
	}

	// If fingerprints are different, use fuzzy string matching
	// We'll use a simple string similarity algorithm based on common prefixes/suffixes
	// and character overlap

	// Find the length of the longest common prefix
	minLen := len(fp1)
	if len(fp2) < minLen {
		minLen = len(fp2)
	}

	commonPrefix := 0
	for i := 0; i < minLen; i++ {
		if fp1[i] == fp2[i] {
			commonPrefix++
		} else {
			break
		}
	}

	// Find the length of the longest common suffix
	commonSuffix := 0
	for i := 1; i <= minLen; i++ {
		if fp1[len(fp1)-i] == fp2[len(fp2)-i] {
			commonSuffix++
		} else {
			break
		}
	}

	// Calculate similarity based on common characters
	totalCommon := commonPrefix + commonSuffix
	maxPossible := len(fp1) + len(fp2)
	if maxPossible == 0 {
		return 0.0
	}

	similarity := float64(totalCommon) / float64(maxPossible)

	// Apply a threshold - if similarity is too low, return 0
	if similarity < 0.3 {
		return 0.0
	}

	return similarity
}

// calculatePriceSimilarity calculates similarity between two prices
func (s *DeduplicationService) calculatePriceSimilarity(price1, price2 float64) float64 {
	if price1 == price2 {
		return 1.0
	}

	// Calculate price difference percentage
	diff := math.Abs(price1 - price2)
	avgPrice := (price1 + price2) / 2
	if avgPrice == 0 {
		return 0.0
	}

	percentageDiff := diff / avgPrice

	// Ensure we have a valid config with default values if nil
	if s.config == nil {
		s.config = &DeduplicationConfig{
			TitleSimilarityThreshold:       0.85,
			PriceSimilarityThreshold:       0.90,
			DescriptionSimilarityThreshold: 0.75,
			OverallSimilarityThreshold:     0.80,
			MaxAgeForComparison:            7 * 24 * time.Hour, // 7 days
		}
	}

	// Convert to similarity score (inverse of difference)
	// If difference is 0%, similarity is 1.0
	// If difference is > threshold, similarity approaches 0
	if percentageDiff <= s.config.PriceSimilarityThreshold {
		return 1.0 - percentageDiff
	}

	return 0.0
}

// CheckForDuplicates checks if an item is a duplicate of existing items
func (s *DeduplicationService) CheckForDuplicates(newItem *db.GormItem) ([]*db.GormItem, error) {
	if newItem == nil {
		return nil, fmt.Errorf("new item cannot be nil")
	}

	// Get recent items for comparison (within max age)
	recentItems, err := s.getRecentItemsForComparison()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent items: %w", err)
	}

	// Early exit if no recent items to compare against
	if len(recentItems) == 0 {
		return nil, nil
	}

	// Generate fingerprint for the new item once to reuse
	newFingerprint, err := s.GenerateItemFingerprint(newItem)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fingerprint for new item: %w", err)
	}

	var duplicates []*db.GormItem

	// Optimize by first checking for exact fingerprint matches (fast path)
	// This helps reduce the number of expensive similarity calculations
	for _, existingItem := range recentItems {
		// Skip self-comparison
		if existingItem.GoodwillID == newItem.GoodwillID {
			continue
		}

		// Fast path: check if combined fingerprints match exactly
		existingFingerprint, err := s.GenerateItemFingerprint(existingItem)
		if err != nil {
			log.Warnf("Failed to generate fingerprint for existing item: %v", err)
			continue
		}

		// If combined fingerprints match exactly, it's definitely a duplicate
		if existingFingerprint.CombinedFingerprint == newFingerprint.CombinedFingerprint {
			duplicates = append(duplicates, existingItem)
			continue
		}

		// If individual fingerprints are very different, skip expensive calculation
		if existingFingerprint.TitleFingerprint != newFingerprint.TitleFingerprint &&
			existingFingerprint.SellerFingerprint != newFingerprint.SellerFingerprint &&
			existingFingerprint.PriceFingerprint != newFingerprint.PriceFingerprint {
			continue
		}

		// Calculate similarity for items that passed the fast checks
		similarity, err := s.CalculateSimilarity(existingItem, newItem)
		if err != nil {
			log.Warnf("Failed to calculate similarity between items: %v", err)
			continue
		}

		// Check if similarity exceeds threshold
		if similarity >= s.config.OverallSimilarityThreshold {
			duplicates = append(duplicates, existingItem)
		}
	}

	return duplicates, nil
}

// getRecentItemsForComparison gets recent items for duplicate comparison using database-level filtering
func (s *DeduplicationService) getRecentItemsForComparison() ([]*db.GormItem, error) {
	// Use database-level filtering to avoid loading all items into memory
	// This is much more efficient than fetching all items and filtering in memory

	// Use the new database method that filters by recency at the database level
	items, _, err := s.repo.GetRecentItemsForDeduplication(
		context.Background(),
		s.config.MaxAgeForComparison,
		1000, // Limit to reasonable number for comparison
		0,    // Start from beginning
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent items: %w", err)
	}

	// Convert to pointer slice
	var recentItems []*db.GormItem
	for i := range items {
		recentItems = append(recentItems, &items[i])
	}

	return recentItems, nil
}

// MergeDuplicateItems merges duplicate items into a single canonical item
func (s *DeduplicationService) MergeDuplicateItems(canonicalItem *db.GormItem, duplicateItems []*db.GormItem) error {
	if canonicalItem == nil {
		return fmt.Errorf("canonical item cannot be nil")
	}
	if len(duplicateItems) == 0 {
		return nil // Nothing to merge
	}

	log.Infof("Merging %d duplicate items into canonical item %s", len(duplicateItems), canonicalItem.GoodwillID)

	// Update canonical item with best available data
	for _, duplicate := range duplicateItems {
		// Update price history if different
		if duplicate.CurrentPrice != canonicalItem.CurrentPrice {
			// Record price history for the duplicate
			priceHistory := db.GormPriceHistory{
				ItemID:     canonicalItem.ID,
				Price:      duplicate.CurrentPrice,
				PriceType:  "duplicate_merge",
				RecordedAt: time.Now(),
			}

			_, err := s.repo.AddPriceHistory(context.Background(), priceHistory)
			if err != nil {
				log.Warnf("Failed to record price history during merge: %v", err)
			}
		}

		// Update description if better (longer)
		if len(duplicate.Description) > len(canonicalItem.Description) {
			canonicalItem.Description = duplicate.Description
		}

		// Update other fields if missing in canonical item
		if canonicalItem.ImageURL == "" && duplicate.ImageURL != "" {
			canonicalItem.ImageURL = duplicate.ImageURL
		}

		if canonicalItem.Category == "" && duplicate.Category != "" {
			canonicalItem.Category = duplicate.Category
		}

		if canonicalItem.Subcategory == "" && duplicate.Subcategory != "" {
			canonicalItem.Subcategory = duplicate.Subcategory
		}

		if canonicalItem.Location == "" && duplicate.Location != "" {
			canonicalItem.Location = duplicate.Location
		}
	}

	// Update the canonical item
	err := s.repo.UpdateItem(context.Background(), *canonicalItem)
	if err != nil {
		return fmt.Errorf("failed to update canonical item: %w", err)
	}

	// Mark duplicate items as merged
	for _, duplicate := range duplicateItems {
		duplicate.Status = "merged"
		duplicate.UpdatedAt = time.Now()

		err := s.repo.UpdateItem(context.Background(), *duplicate)
		if err != nil {
			log.Warnf("Failed to mark duplicate item %s as merged: %v", duplicate.GoodwillID, err)
		}
	}

	return nil
}

// IsPotentialDuplicate checks if a new item is potentially a duplicate
func (s *DeduplicationService) IsPotentialDuplicate(newItem *db.GormItem) (bool, []*db.GormItem, error) {
	duplicates, err := s.CheckForDuplicates(newItem)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if len(duplicates) > 0 {
		return true, duplicates, nil
	}

	return false, nil, nil
}

// GetDeduplicationStats returns statistics about deduplication using database-level aggregation
func (s *DeduplicationService) GetDeduplicationStats() (map[string]interface{}, error) {
	// Handle nil repository for testing purposes
	if s.repo == nil {
		stats := make(map[string]interface{})
		stats["total_items"] = 0
		stats["active_items"] = 0
		stats["merged_items"] = 0
		stats["duplicate_rate"] = 0.0
		return stats, nil
	}

	// Use database-level filtering to get items by status - much more efficient
	activeStatus := "active"
	_, activeCount, err := s.repo.GetItemsFiltered(
		context.Background(),
		nil,           // No search ID filter
		&activeStatus, // Only active items
		nil,           // No category filter
		nil,           // No min price filter
		nil,           // No max price filter
		1,             // Limit 1 to avoid fetching many rows, we only want the count
		0,             // No offset
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get active items count: %w", err)
	}

	mergedStatus := "merged"
	_, mergedCount, err := s.repo.GetItemsFiltered(
		context.Background(),
		nil,           // No search ID filter
		&mergedStatus, // Only merged items
		nil,           // No category filter
		nil,           // No min price filter
		nil,           // No max price filter
		1,             // Limit 1
		0,             // No offset
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get merged items count: %w", err)
	}

	// Get total items count using filtered approach
	_, totalCount, err := s.repo.GetItemsFiltered(
		context.Background(),
		nil, // No search ID filter
		nil, // No status filter (all items)
		nil, // No category filter
		nil, // No min price filter
		nil, // No max price filter
		1,   // Limit 1
		0,   // No offset
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get total items count: %w", err)
	}

	stats := make(map[string]interface{})

	stats["total_items"] = totalCount
	stats["active_items"] = activeCount
	stats["merged_items"] = mergedCount
	stats["duplicate_rate"] = calculateDuplicateRate(totalCount, mergedCount)

	return stats, nil
}

func calculateDuplicateRate(totalItems, mergedItems int) float64 {
	if totalItems == 0 {
		return 0.0
	}
	return float64(mergedItems) / float64(totalItems)
}

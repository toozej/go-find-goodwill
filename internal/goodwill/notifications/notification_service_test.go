package notifications

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

func TestNewNotificationManager(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Gotify: config.GotifyConfig{
				Enabled:  true,
				URL:      "http://localhost:8080",
				Token:    "test-token",
				Priority: 5,
			},
			Slack: config.SlackConfig{
				Enabled:   true,
				Token:     "slack-token",
				ChannelID: "C123456",
			},
			Telegram: config.TelegramConfig{
				Enabled: false, // Disable Telegram for now due to initialization issues
				Token:   "telegram-token",
				ChatID:  "123456789",
			},
			Discord: config.DiscordConfig{
				Enabled:   true,
				Token:     "discord-token",
				ChannelID: "D123456",
			},
			Pushover: config.PushoverConfig{
				Enabled:     true,
				Token:       "pushover-token",
				RecipientID: "user123",
			},
			Pushbullet: config.PushbulletConfig{
				Enabled:        true,
				Token:          "pushbullet-token",
				DeviceNickname: "my-device",
			},
		},
	}

	// Create notification manager
	manager, err := NewNotificationManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, true, manager.condense)
	assert.Equal(t, 2, len(manager.notifiers)) // Gotify + Nikoksr
}

func TestNotificationManagerNotifyFound(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Gotify: config.GotifyConfig{
				Enabled:  true,
				URL:      "http://localhost:8080",
				Token:    "test-token",
				Priority: 5,
			},
		},
	}

	// Create notification manager
	manager, err := NewNotificationManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test data
	item := &db.GormItem{
		ID:           1,
		GoodwillID:   "test-123",
		Title:        "Vintage Camera",
		CurrentPrice: 99.99,
		URL:          "http://example.com/item/123",
		Category:     "Cameras",
		EndsAt:       timePtr(time.Now().Add(24 * time.Hour)),
	}

	search := &db.GormSearch{
		ID:    1,
		Name:  "Test Search",
		Query: "vintage camera",
	}

	// Test notification (this will fail to send but should not panic)
	err = manager.NotifyFound(context.Background(), item, search)
	assert.Error(t, err) // Should return error when sending fails
}

func TestNotificationManagerNotifyFoundItems(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Gotify: config.GotifyConfig{
				Enabled:  true,
				URL:      "http://localhost:8080",
				Token:    "test-token",
				Priority: 5,
			},
		},
	}

	// Create notification manager
	manager, err := NewNotificationManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test data
	items := []*db.GormItem{
		{
			ID:           1,
			GoodwillID:   "test-123",
			Title:        "Vintage Camera",
			CurrentPrice: 99.99,
			URL:          "http://example.com/item/123",
			Category:     "Cameras",
			EndsAt:       timePtr(time.Now().Add(24 * time.Hour)),
		},
		{
			ID:           2,
			GoodwillID:   "test-456",
			Title:        "Antique Clock",
			CurrentPrice: 49.99,
			URL:          "http://example.com/item/456",
			Category:     "Clocks",
			EndsAt:       timePtr(time.Now().Add(48 * time.Hour)),
		},
	}

	search := &db.GormSearch{
		ID:    1,
		Name:  "Test Search",
		Query: "vintage items",
	}

	// Test notification (this will fail to send but should not panic)
	err = manager.NotifyFoundItems(context.Background(), items, search)
	assert.Error(t, err) // Should return error when sending fails
}

func TestNotificationManagerNotifyHeartbeat(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Notification: config.NotificationConfig{
			Gotify: config.GotifyConfig{
				Enabled:  true,
				URL:      "http://localhost:8080",
				Token:    "test-token",
				Priority: 5,
			},
		},
	}

	// Create notification manager
	manager, err := NewNotificationManager(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test heartbeat notification (this will fail to send but should not panic)
	err = manager.NotifyHeartbeat(context.Background())
	assert.Error(t, err) // Should return error when sending fails
}

func timePtr(t time.Time) *time.Time {
	return &t
}

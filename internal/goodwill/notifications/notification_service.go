package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/discord"
	"github.com/nikoksr/notify/service/pushbullet"
	"github.com/nikoksr/notify/service/pushover"
	"github.com/nikoksr/notify/service/slack"
	"github.com/nikoksr/notify/service/telegram"
	log "github.com/sirupsen/logrus"

	"github.com/toozej/go-find-goodwill/internal/goodwill/db"
	"github.com/toozej/go-find-goodwill/pkg/config"
)

// Notifier is an interface for sending notifications
type Notifier interface {
	Notify(ctx context.Context, subject, message string) error
}

// GotifyNotifier implements direct Gotify API integration
type GotifyNotifier struct {
	endpoint string
	token    string
	client   *http.Client
}

// NewGotifyNotifier creates a new Gotify notifier
func NewGotifyNotifier(endpoint, token string) *GotifyNotifier {
	return &GotifyNotifier{
		endpoint: strings.TrimSuffix(endpoint, "/"),
		token:    token,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a notification to Gotify
func (g *GotifyNotifier) Notify(ctx context.Context, subject, message string) error {
	url := fmt.Sprintf("%s/message?token=%s", g.endpoint, g.token)

	payload := map[string]interface{}{
		"title":    subject,
		"message":  message,
		"priority": 5,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("gotify returned status code %d", resp.StatusCode)
	}

	return nil
}

// NikoksrNotifier uses the nikoksr/notify library for other notification services
type NikoksrNotifier struct {
	notifier *notify.Notify
}

// NewNikoksrNotifier creates a new notifier using nikoksr/notify
func NewNikoksrNotifier() *NikoksrNotifier {
	return &NikoksrNotifier{
		notifier: notify.New(),
	}
}

// AddSlack adds Slack notification service
func (n *NikoksrNotifier) AddSlack(token string, channelID string) {
	service := slack.New(token)
	service.AddReceivers(channelID)
	n.notifier.UseServices(service)
}

// AddTelegram adds Telegram notification service
func (n *NikoksrNotifier) AddTelegram(token string, chatID int64) {
	service, _ := telegram.New(token)
	service.AddReceivers(chatID)
	n.notifier.UseServices(service)
}

// AddDiscord adds Discord notification service
func (n *NikoksrNotifier) AddDiscord(token string, channelID string) {
	service := discord.New()
	_ = service.AuthenticateWithBotToken(token)
	service.AddReceivers(channelID)
	n.notifier.UseServices(service)
}

// AddPushover adds Pushover notification service
func (n *NikoksrNotifier) AddPushover(token string, recipientID string) {
	service := pushover.New(token)
	service.AddReceivers(recipientID)
	n.notifier.UseServices(service)
}

// AddPushbullet adds Pushbullet notification service
func (n *NikoksrNotifier) AddPushbullet(token string, deviceNickname string) {
	service := pushbullet.New(token)
	service.AddReceivers(deviceNickname)
	n.notifier.UseServices(service)
}

// Notify sends a notification using nikoksr/notify
func (n *NikoksrNotifier) Notify(ctx context.Context, subject, message string) error {
	return n.notifier.Send(ctx, subject, message)
}

// NotificationManager manages multiple notification providers
type NotificationManager struct {
	notifiers []Notifier
	condense  bool
}

// NewNotificationManager creates a notification manager from config
func NewNotificationManager(cfg *config.Config) (*NotificationManager, error) {
	manager := &NotificationManager{}

	// Determine condense setting - should be independent of Gotify
	// For now, we'll use a reasonable default or make it configurable
	// This fixes the arbitrary coupling to Gotify being enabled
	manager.condense = true // Default to true for now, can be made configurable later

	// Add nicoksr notify for handling multiple services
	nikoksrNotifier := NewNikoksrNotifier()
	nikoksrAdded := false

	// Add Gotify if enabled
	if cfg.Notification.Gotify.Enabled {
		gotify := NewGotifyNotifier(cfg.Notification.Gotify.URL, cfg.Notification.Gotify.Token)
		manager.notifiers = append(manager.notifiers, gotify)
	}

	// Add Slack if configured
	if cfg.Notification.Slack.Enabled {
		nikoksrNotifier.AddSlack(cfg.Notification.Slack.Token, cfg.Notification.Slack.ChannelID)
		nikoksrAdded = true
	}

	// Add Telegram if configured
	if cfg.Notification.Telegram.Enabled {
		chatID, err := strconv.ParseInt(cfg.Notification.Telegram.ChatID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid Telegram chat ID: %w", err)
		}
		nikoksrNotifier.AddTelegram(cfg.Notification.Telegram.Token, chatID)
		nikoksrAdded = true
	}

	// Add Discord if configured
	if cfg.Notification.Discord.Enabled {
		nikoksrNotifier.AddDiscord(cfg.Notification.Discord.Token, cfg.Notification.Discord.ChannelID)
		nikoksrAdded = true
	}

	// Add Pushover if configured
	if cfg.Notification.Pushover.Enabled {
		nikoksrNotifier.AddPushover(cfg.Notification.Pushover.Token, cfg.Notification.Pushover.RecipientID)
		nikoksrAdded = true
	}

	// Add Pushbullet if configured
	if cfg.Notification.Pushbullet.Enabled {
		nikoksrNotifier.AddPushbullet(cfg.Notification.Pushbullet.Token, cfg.Notification.Pushbullet.DeviceNickname)
		nikoksrAdded = true
	}

	// Add nikoksr notifier if any services were added to it
	if nikoksrAdded {
		manager.notifiers = append(manager.notifiers, nikoksrNotifier)
	}

	return manager, nil
}

// NotifyFound sends notifications for found goodwill items
func (m *NotificationManager) NotifyFound(ctx context.Context, item *db.GormItem, search *db.GormSearch) error {
	subject := fmt.Sprintf("GFG - Found %s!", item.Title)
	var endsAtStr string
	if item.EndsAt != nil {
		endsAtStr = item.EndsAt.Format("2006-01-02 15:04:05")
	} else {
		endsAtStr = "Unknown"
	}
	message := fmt.Sprintf("Found %s at %s for $%.2f\nURL: %s\nCategory: %s\nEnds: %s",
		item.Title,
		item.Seller,
		item.CurrentPrice,
		item.URL,
		item.Category,
		endsAtStr,
	)

	log.Info(message)

	var errors []error
	for _, notifier := range m.notifiers {
		if err := notifier.Notify(ctx, subject, message); err != nil {
			log.Errorf("Failed to send notification: %v", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send %d notifications: %v", len(errors), errors)
	}
	return nil
}

package threat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"open-website-defender/internal/infrastructure/logging"
	"open-website-defender/internal/usecase/system"
	"time"

	"github.com/spf13/viper"
)

type webhookPayload struct {
	EventType string `json:"event_type"`
	ClientIP  string `json:"client_ip"`
	Reason    string `json:"reason"`
	BannedFor string `json:"banned_for"`
	Timestamp string `json:"timestamp"`
}

func getWebhookURL() string {
	settings, err := system.GetSystemService().GetSettings()
	if err == nil && settings != nil && settings.WebhookURL != "" {
		return settings.WebhookURL
	}
	return viper.GetString("webhook.url")
}

func sendWebhookNotification(eventType, clientIP, reason string, duration time.Duration) {
	webhookURL := getWebhookURL()
	if webhookURL == "" {
		return
	}

	// Check if this event type is in the configured events list
	configuredEvents := viper.GetStringSlice("webhook.events")
	if len(configuredEvents) > 0 {
		found := false
		for _, e := range configuredEvents {
			if e == eventType {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	payload := webhookPayload{
		EventType: eventType,
		ClientIP:  clientIP,
		Reason:    reason,
		BannedFor: duration.String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logging.Sugar.Errorf("Failed to marshal webhook payload: %v", err)
		return
	}

	timeout := viper.GetInt("webhook.timeout")
	if timeout <= 0 {
		timeout = 5
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		logging.Sugar.Errorf("Webhook notification failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logging.Sugar.Warnf("Webhook returned status %d", resp.StatusCode)
	} else {
		logging.Sugar.Debugf("Webhook notification sent for %s event (IP: %s)", eventType, clientIP)
	}
}

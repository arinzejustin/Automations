package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-faker/faker/v4"
)

// Subscriber represents a mock subscriber payload.
type Subscriber struct {
	Email string `faker:"email" json:"email"`
}

// LogPayload defines the structure for sending logs.
type LogPayload struct {
	Email     string `json:"email"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// sendJSON performs an HTTP POST with JSON payload and custom headers.
func sendJSON(url string, data any, headers map[string]string) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// main is the entry point for the program.
func main() {
	apiURL := os.Getenv("API_URL")
	logURL := os.Getenv("LOG_URL")
	origin := os.Getenv("ORIGIN")
	
	fmt.Println(apiURL, logURL, origin)

	if apiURL == "" {
		fmt.Println("❌ API_URL not set in environment")
		return
	}

	// Generate a fake subscriber
	subscriber := Subscriber{
		Email: faker.Email(),
	}

	headers := map[string]string{
		"Origin": origin,
	}

	status := "FAILED"
	message := "Subscription failed"

	// Attempt subscription
	subscribeURL := fmt.Sprintf("%s/subscribe", strings.TrimRight(apiURL, "/"))
	data := map[string]string{"email": subscriber.Email}

	subscribeRes, subscribeErr := sendJSON(subscribeURL, data, headers)

	if subscribeErr != nil {
		message = fmt.Sprintf("Error sending subscribe request: %v", subscribeErr)
	} else {
		defer subscribeRes.Body.Close()
		message = subscribeRes.Status

		if subscribeRes.StatusCode >= 200 && subscribeRes.StatusCode < 300 {
			status = "SUCCESS"

			// Attempt to unsubscribe after success
			unsubscribeURL := fmt.Sprintf("%s/unsubscribe", strings.TrimRight(apiURL, "/"))
			data := map[string]string{"email": subscriber.Email}

			unsubscribeRes, unsubscribeErr := sendJSON(unsubscribeURL, data, headers)
			if unsubscribeErr != nil {
				message = fmt.Sprintf("Error sending unsubscribe request: %v", unsubscribeErr)
			} else {
				defer unsubscribeRes.Body.Close()
				message = fmt.Sprintf("Unsubscribe request: %s", unsubscribeRes.Status)
			}
		}
	}

	// Prepare and send log payload
	logPayload := LogPayload{
		Email:     subscriber.Email,
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    status,
		Message:   message,
	}

	if logURL != "" {
		if _, err := sendJSON(logURL, logPayload, nil); err != nil {
			fmt.Printf("⚠️  Failed to send log: %v\n", err)
		}
	}

	// Console summary
	fmt.Printf("📨 %s | %s | %s\n", logPayload.Email, logPayload.Status, logPayload.Message)
}

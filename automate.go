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

type Subscriber struct {
	Email string `faker:"email"`
}

type LogPayLoad struct {
	Email     string `json:"email"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func sendJSON(url string, data any, headers map[string]string) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	return client.Do(req)
}

func main() {
	apiUrl := os.Getenv("API_URL")
	logUrl := os.Getenv("LOG_URL")

	if apiUrl == "" {
		fmt.Println("API_URL not set")
		return
	}

	subscriber := Subscriber{
		Email: faker.Email(),
	}

	headers := map[string]string{
		"Origin": "CrawlerBotMe",
	}

	resp1, err := sendJSON(strings.Join([]string{apiUrl, "subscribers"}, "/"), subscriber, headers)
	status := "FAILED"
	message := "Subscription failed"

	if err != nil {
		message = fmt.Sprintf("First request error: %v", err)
	} else {
		if resp1 != nil {
			defer resp1.Body.Close()

			message = resp1.Status

			if resp1.StatusCode >= 200 && resp1.StatusCode < 300 {
				status = "SUCCESS"
				data2 := map[string]string{"email": "hello@example.com"}
				resp2, err2 := sendJSON(strings.Join([]string{apiUrl, "unsubscribe"}, "/"), data2, headers)
				if err2 != nil {
					message = fmt.Sprintf("Second request error: %v", err2)
				} else {
					if resp2 != nil {
						defer resp2.Body.Close()
						message = fmt.Sprintf("Second request: %s", resp2.Status)
					}
				}
			}
		}
	}

	logPayLoad := LogPayLoad{
		Email:     subscriber.Email,
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    status,
		Message:   message,
	}

	if logUrl != "" {
		sendJSON(logUrl, logPayLoad, nil)
	}

	fmt.Printf("📨 %s | %s | %s\n", logPayLoad.Email, logPayLoad.Status, logPayLoad.Message)
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Get ENV OR Use Default
func geoud(key, predefined string) string {
	var bot_token, bterr = os.LookupEnv(key)
	if bterr {
		return bot_token
	}
	return predefined
}

var endpoint = "https://api.telegram.org/bot" + geoud("BOT_TOKEN", "6979214367:AAFRNskwlYHuxDbIUGimRpx0I0QykMK6VNM") + "/sendMessage"

// Message represents the structure of the message to be sent
type Message struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func main() {
	fmt.Println(endpoint)
	// Number of instances to run concurrently
	numInstances, _ := strconv.ParseInt(geoud("APP_INSTANCES", "20"), 10, 8)

	for i := 0; i < int(numInstances); i++ {
		go func() {
			for {
				try(func() {
					// Create a HTTP client
					client := &http.Client{}

					// Create a message payload
					message := Message{
						ChatID: geoud("CHAT_ID", "6638703426"),
						Text:   geoud("CHAT_TEXT", "\u2063"), // Placeholder text
					}
					messageBytes, _ := json.Marshal(message)

					// Create a HTTP request
					req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(messageBytes))
					if err != nil {
						panic(fmt.Errorf("error creating request: %w", err))
					}
					req.Header.Set("Content-Type", "application/json")

					// Send the HTTP request
					resp, err := client.Do(req)
					if err != nil {
						sendtoAuthor, _ := http.NewRequest("GET", endpoint+"?chat_id="+geoud("AUTHOR_CHAT_ID", "")+"&text="+err.Error(), nil)
						client.Do(sendtoAuthor)
						panic(fmt.Errorf("error sending request: %w", err))
					}
					defer resp.Body.Close()

					// Process the response
					if resp.StatusCode == http.StatusOK {
						fmt.Println("Message sent successfully.")
					} else {
						var body map[string]interface{}
						err := json.NewDecoder(resp.Body).Decode(&body)
						if err != nil {
							panic(fmt.Errorf("error decoding response body: %w", err))
						}
						status, ok := body["ok"].(bool)
						if ok && !status {
							fmt.Println("Error message:", body["description"])
							if body["error_code"].(float64) == 429 {
								retryAfter := body["parameters"].(map[string]interface{})["retry_after"].(float64)
								sleepDuration := time.Duration(retryAfter) * time.Second
								fmt.Println("Retrying After:", sleepDuration)
								time.Sleep(sleepDuration)
							}
						}
					}
				}, func(err error) {
					fmt.Println("Error:", err)
				})

				time.Sleep(time.Second) // Sleep for 1 second before sending the next message
			}
		}()
	}

	// Keep the main Goroutine running
	select {}
}

// try mimics a try-catch block
func try(do func(), catch func(error)) {
	defer func() {
		if r := recover(); r != nil {
			catch(fmt.Errorf("%v", r))
		}
	}()
	do()
}

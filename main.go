package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Messages []Message `json:"messages"`
}

type ResponseBody struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatalf("OPENAI_API_KEY not set in environment")
	}
	apiEndpoint := os.Getenv("OPENAI_API_ENDPOINT")
	if apiEndpoint == "" {
		log.Fatalf("OPENAI_API_ENDPOINT not set in environment")
	}
	fmt.Println("現在地は？")
	var location string
	if _, err := fmt.Scan(&location); err != nil {
		log.Fatalf("Error reading location: %v", err)
	}
	findRestaurant(apiKey, apiEndpoint, location)
}

func findRestaurant(apiKey, apiEndpoint, location string) {
	client := resty.New()
	prompt := fmt.Sprintf("%s 付近のおすすめのレストランを3つ教えて", location)
	requestBody := RequestBody{
		Messages: []Message{
			{
				Role:    "system",
				Content: prompt,
			},
		},
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("Error marshaling request body: %v", err)
	}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("api-key", apiKey).
		SetBody(jsonData).
		Post(fmt.Sprintf("%s/openai/deployments/gpt-4o-server/chat/completions?api-version=2024-02-01", apiEndpoint))

	if err != nil {
		log.Fatalf("Error making API request: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		log.Fatalf("Non-OK HTTP status: %s", resp.Status())
	}
	var responseBody ResponseBody
	err = json.Unmarshal(resp.Body(), &responseBody)
	if err != nil {
		log.Fatalf("Error unmarshaling response body: %v", err)
	}
	if len(responseBody.Choices) > 0 {
		fmt.Println(responseBody.Choices[0].Message.Content)
	} else {
		fmt.Println("すみません。おすすめのレストランが見つかりませんでした。")
	}
}

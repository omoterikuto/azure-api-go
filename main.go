package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
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

var (
	apiKey      string
	apiEndpoint string
)

func loadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}
	apiKey = os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return err
	}
	apiEndpoint = os.Getenv("OPENAI_API_ENDPOINT")
	if apiEndpoint == "" {
		return err
	}
	return nil
}

var marks = []string{"|", "/", "-", "\\"}

func mark(i int) string {
	return marks[i%len(marks)]
}

func main() {
	if err := loadEnv(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}
	promptLocation := promptui.Prompt{
		Label: "現在地",
	}
	location, err := promptLocation.Run()
	if err != nil {
		log.Fatalf("Error reading location: %v", err)
	}
	promptFee := promptui.Select{
		Label: "希望金額",
		Items: []string{"~1000円", "1000円~3000円", "3000円~5000円", "5000円~10000円", "10000円~"},
	}
	_, fee, err := promptFee.Run()
	if err != nil {
		log.Fatalf("Error reading fee: %v", err)
	}
	promptFavorite := promptui.Select{
		Label: "料理",
		Items: []string{"和食", "イタリアン", "中華", "洋食", "ファストフード"},
	}
	_, favorite, err := promptFavorite.Run()
	if err != nil {
		log.Fatalf("Error reading fee: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ticket := time.NewTicker(100 * time.Millisecond)
	var count int
	go func() {
		for ; ; count++ {
			select {
			case <-ctx.Done():
				break
			case <-ticket.C:
				fmt.Printf("\rおすすめのレストランを検索中 %s", mark(count))
			}
		}
	}()
	findRestaurant(cancel, location, fee, favorite)
}

func findRestaurant(cancelFunc context.CancelFunc, location, fee, favorite string) {
	client := resty.New()
	prompt := fmt.Sprintf("%s 付近のおすすめの%sレストランを3つ教えて。金額は%sで探して。", favorite, location, fee)
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
	cancelFunc()
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

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type chatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`

	// an array of possible choices for the next message
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`

	// usage statistics
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	// Define API endpoint URL and authentication key
	apiURL := "https://api.openai.com/v1/chat/completions"

	// models
	models := []string{
		"gpt-4",
		"gpt-3.5-turbo",
	}

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	// Get the API key from the environment variables
	authKey := os.Getenv("OPENAI_API_KEY")
	if authKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		return
	}

	// Create function to send POST request to API
	sendRequest := func(prompt string) (chatCompletionResponse, error) {
		// Create request body
		requestBody := strings.NewReader(fmt.Sprintf(`{"messages": [{"role": "user", "content": "%s"}], "model": "%s"}`, prompt, models[0]))

		// Create HTTP request
		request, err := http.NewRequest("POST", apiURL, requestBody)
		if err != nil {
			return chatCompletionResponse{}, err
		}

		// Set request headers
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authKey))

		// Send request
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			return chatCompletionResponse{}, err
		}
		defer response.Body.Close()

		// Read response body
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return chatCompletionResponse{}, err
		}

		// Parse response body
		var r chatCompletionResponse
		err = json.Unmarshal(responseBody, &r)
		if err != nil {
			return chatCompletionResponse{}, err
		}

		// Return the first choice
		return r, nil
	}

	// Create loop to continuously prompt user for input and send to API
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		scanner.Scan()
		userInput := scanner.Text()

		response, err := sendRequest(userInput)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// for each choice in response, print the message
		for _, choice := range response.Choices {
			fmt.Printf("AI: %s\n", choice.Message.Content)
		}
	}
}

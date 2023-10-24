package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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

	// prompt user to choose a mode at start of cli tool
	fmt.Println("Choose a model:")
	fmt.Println("1. GPT-4")
	fmt.Println("2. GPT-3.5 Turbo")
	fmt.Print("Enter a number: ")

	// get user input
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userInput := scanner.Text()
	// if user input is not 1 or 2, exit
	if userInput != "1" && userInput != "2" {
		fmt.Println("Error: invalid input")
		return
	}

	// models
	models := []string{
		"gpt-4",
		"gpt-3.5-turbo",
	}

	// set model based on user input
	model := models[0]
	if userInput == "2" {
		model = models[1]
	}

	// Print model name
	fmt.Printf("Using model: %s\n", model)

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
		requestBody := strings.NewReader(fmt.Sprintf(`{"messages": [{"role": "user", "content": "%s"}], "model": "%s"}`, prompt, model))

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
		loading := make(chan bool, 1)
		go func() {
			for {
				select {
				case <-loading:
					return
				default:
					fmt.Print(".")
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()
		response, err := client.Do(request)
		if err != nil {
			return chatCompletionResponse{}, err
		}
		defer response.Body.Close()

		// Read response body
		responseBody, err := io.ReadAll(response.Body)
		loading <- true
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

	// print a friendly AIBuddy message with cute text emoji at end
	fmt.Println("Hi I'm Yosh! 🦖👋. Type 'q' to exit.")

	for {
		// check for quit command
		fmt.Print("You: ")
		scanner.Scan()

		// intercept user input if it is 'q' and exit
		userInput := scanner.Text()
		if userInput == "q" {
			fmt.Println("Bye!")
			return
		}

		// send user input to API
		response, err := sendRequest(userInput)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// for each choice in response, print the message
		fmt.Println()
		for _, choice := range response.Choices {
			fmt.Printf("AI: %s\n", choice.Message.Content)
		}
	}
}

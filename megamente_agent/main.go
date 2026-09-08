package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"

	"google.golang.org/genai"
)

func main() {

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  "API KEY FOR GEMINIS",
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	chat, err_chat := client.Chats.Create(
		context.Background(),
		"gemini-3.5-flash",
		&genai.GenerateContentConfig{},
		nil,
	)
	if err_chat != nil {
		fmt.Println(err_chat)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		if err_scanner := scanner.Err(); err_scanner != nil {
			break
		}

		text := scanner.Text()
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		res, error_response := chat.SendMessage(context.Background(), genai.Part{
			Text: text,
		})
		if error_response != nil {
			fmt.Println(error_response.Error())
			return
		}

		fmt.Println(res.Text())
	}
}

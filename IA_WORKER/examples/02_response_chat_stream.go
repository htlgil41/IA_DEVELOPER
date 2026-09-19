package examples

import (
	"context"
	"fmt"
	"ia_worker/libs"
	"log"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func ResponseChatStream(config *libs.Config) {
	log.Println("Chat Stream flujo para el modelo")
	ctx := context.Background()

	newClient := openai.NewClient(
		option.WithAPIKey(config.Cloudflared.APIKey),
		option.WithBaseURL(fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/v1", config.Cloudflared.Account)),
	)

	stream_response := newClient.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:       config.Cloudflared.Modelo,
		Temperature: openai.Float(0.7),
		MaxTokens:   openai.Int(1000),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Eres un asistente virtual llamada RUD eres muy amable y con un habla directo siendo muy calida y clara en el tema"),
			openai.UserMessage("Hola soy aaron, me puedes definir el termino 'Vanguardia' por favor y me das 3 ejemplos sobre como usarla en frases"),
		},
	})
	defer stream_response.Close()

	var responesComplete strings.Builder
	for stream_response.Next() {
		chuc := stream_response.Current()
		for _, choice := range chuc.Choices {
			responesComplete.WriteString(choice.Delta.Content)
			fmt.Print(choice.Delta.Content)
			os.Stdout.Sync()
		}
	}
	fmt.Println()

	if errRespnse := stream_response.Err(); errRespnse != nil {
		fmt.Println("Error en la respuesta del stream")
	}

	fmt.Println("Respusta completa del modelo")
	fmt.Println(responesComplete.String())
}

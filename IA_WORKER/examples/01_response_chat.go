package examples

import (
	"context"
	"fmt"
	"ia_worker/libs"
	"log"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

/*
	Ejemplo basico de una sentencia de respuesta simple de presentacion
*/

func ResponseChat(config *libs.Config) {
	log.Println("Config loaded, ", config)
	ctx := context.Background()
	client := openai.NewClient(
		option.WithAPIKey(config.Cloudflared.APIKey),
		option.WithBaseURL(fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/v1", config.Cloudflared.Account)),
	)

	question := "Hola soy aaron quien eres tu?"
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(question),
		},
		Model: "@cf/meta/llama-3.3-70b-instruct-fp8-fast",
	})

	if err != nil {
		panic(err.Error())
	}

	println(resp.RawJSON())
}

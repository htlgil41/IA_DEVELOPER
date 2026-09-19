package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"ia_worker/libs"
	"log"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

/*
	Crear un loop Agente con una unica tarea que es crear un cliente
*/

func ChatWitUs(config *libs.Config) {
	log.Println("Chat with us for CMD/STDIn")
	ctx := context.Background()

	client := openai.NewClient(
		option.WithAPIKey(config.Cloudflared.APIKey),
		option.WithBaseURL(fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/v1", config.Cloudflared.Account)),
	)

	CrearClienteInformacion := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Nombre del cliente",
			},
			"lastname": map[string]any{
				"type":        "string",
				"description": "Apellido paterno del cliente",
			},
			"skills": map[string]any{
				"type":        "string",
				"description": "Habilidades que tiene el cliente",
			},
		},
		"required":             []string{"name", "lastname", "skills"},
		"additionalProperties": false,
	}

	client.Chat.Completions.NewStreaming(
		ctx,
		openai.ChatCompletionNewParams{
			Model:            config.Cloudflared.Modelo,
			Temperature:      openai.Float(0.4),
			FrequencyPenalty: openai.Float(0.5),
			SafetyIdentifier: openai.String("_htlgil41"),
			ToolChoice: openai.ChatCompletionToolChoiceOptionParamOfChatCompletionNamedToolChoice(openai.ChatCompletionNamedToolChoiceFunctionParam{
				Name: "create_client",
			}),
			Tools: []openai.ChatCompletionToolParam{
				{
					Function: shared.FunctionDefinitionParam{
						Name:        "create_client",
						Description: openai.String("Crea un cliente o un usuario con sus habilidades"),
						Strict:      openai.Bool(true),
						Parameters:  CrearClienteInformacion,
					},
				},
			},
			Messages: []openai.ChatCompletionMessageParamUnion{},
		},
	)

}

type CreateClientArgs struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Skills   string `json:"skills"`
}

func EjecutarCreateClient(jsonArgs string) (string, error) {
	var args CreateClientArgs
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return "", fmt.Errorf("error al parsear los argumentos de la tool: %w", err)
	}

	respuesta := map[string]any{
		"status":  "success",
		"message": fmt.Sprintf("Cliente %s %s registrado correctamente.", args.Name, args.Lastname),
		"id":      "cli_982341",
	}

	resultadoBytes, err := json.Marshal(respuesta)
	if err != nil {
		return "", err
	}

	return string(resultadoBytes), nil
}

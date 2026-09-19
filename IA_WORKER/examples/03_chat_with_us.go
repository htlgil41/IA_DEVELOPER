package examples

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"ia_worker/libs"
	"log"
	"os"
	"strings"
	"time"

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

	var messages []openai.ChatCompletionMessageParamUnion
	messages = append(messages, openai.SystemMessage("Eres un asistente servicial capaz de crear clientes utilizando la herramienta disponible."))

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n==========================================")
	fmt.Println(" Chat iniciado. Escribe 'salir' para terminar.")
	fmt.Println("==========================================\n")

	for {
		fmt.Println("Usuario ->")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}

		if strings.ToLower(userInput) == "salida" {
			fmt.Println("Saliendo del chat")
			time.Sleep(2 * time.Second)
			break
		}

		messages = append(messages, openai.UserMessage(userInput))
		//agente_loop:
		for {
			stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
				Model:            config.Cloudflared.Modelo,
				Temperature:      openai.Float(0.4),
				FrequencyPenalty: openai.Float(0.5),
				SafetyIdentifier: openai.String("_htlgil41"),
				ToolChoice: openai.ChatCompletionToolChoiceOptionUnionParam{
					OfAuto: openai.String("auto"),
				},
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
				Messages: messages,
			})

			var fullContent strings.Builder
			type accumulatedToolCall struct {
				ID         string
				Name       string
				ArgsBuffer strings.Builder
			}
			toolCallsMap := make(map[int64]*accumulatedToolCall)
			fmt.Print("Agente > ")

			for stream.Next() {
				chunk := stream.Current()
				if len(chunk.Choices) == 0 {
					continue
				}

				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					fmt.Print(delta.Content)
					fullContent.WriteString(delta.Content)
				}

				for _, tcDelta := range delta.ToolCalls {
					idx := tcDelta.Index
					if _, exists := toolCallsMap[idx]; !exists {
						toolCallsMap[idx] = &accumulatedToolCall{}
					}
					if tcDelta.ID != "" {
						toolCallsMap[idx].ID = tcDelta.ID
					}
					if tcDelta.Function.Name != "" {
						toolCallsMap[idx].Name = tcDelta.Function.Name
					}
					if tcDelta.Function.Arguments != "" {
						toolCallsMap[idx].ArgsBuffer.WriteString(tcDelta.Function.Arguments)
					}
				}
			}

			if err := stream.Err(); err != nil {
				log.Printf("\nError procesando stream: %v\n", err)
				break
			}
			fmt.Println()
			messages = append(messages, openai.AssistantMessage(fullContent.String()))
			if len(toolCallsMap) == 0 {
				break
			}

			for _, tc := range toolCallsMap {
				if tc.Name == "create_client" {
					argsJSON := tc.ArgsBuffer.String()
					fmt.Printf("\n[AGENTE Executing Tool: %s con args: %s]\n", tc.Name, argsJSON)

					resultOutput, errorTools := EjecutarCreateClient(argsJSON)
					if errorTools != nil {
						messages = append(messages, openai.ToolMessage(tc.ID, resultOutput))
						fmt.Printf("[AGENTE Tool Output]: %s\n\n", resultOutput)
						continue
					}

					messages = append(messages, openai.ToolMessage(tc.ID, resultOutput))
					fmt.Printf("[AGENTE Tool Output]: %s\n\n", resultOutput)
				}
			}
		}
	}
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

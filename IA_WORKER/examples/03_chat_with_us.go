package examples

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"ia_worker/libs"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

const (
	maxAgentIterations = 5
	systemPrompt       = `
		Eres un asistente que SOLO puede crear clientes usando la herramienta 'create_client'.
		REGLAS ESTRICTAS:
		1. Cuando el usuario pida crear un cliente, DEBES llamar la herramienta 'create_client'.
		2. NO escribas JSON en tu respuesta. NO expliques la llamada. NO la simules.
		3. Si faltan datos (name, lastname, skills), pídelos en UNA sola frase corta.
		4. Cuando tengas los 3 datos, llama la herramienta inmediatamente sin confirmar.
		5. Después de que la herramienta se ejecute, responde SOLO con el resultado.
		6. Nunca inventes IDs. Nunca imites el formato de la herramienta en texto.
	`
)

func ChatWitUs(config *libs.Config) {
	log.Println("Chat with us for CMD/STDIn")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	messages = append(messages, openai.SystemMessage(systemPrompt))

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

		if strings.ToLower(userInput) == "salir" {
			fmt.Println("Saliendo del chat")
			time.Sleep(2 * time.Second)
			break
		}

		messages = append(messages, openai.UserMessage(userInput))

		iteration := 0
		for iteration < maxAgentIterations {
			iteration++

			stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
				Model:            config.Cloudflared.Modelo,
				Temperature:      openai.Float(0.4),
				FrequencyPenalty: openai.Float(0.1),
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
				log.Printf("\nError procesando stream: %v\n", err.Error())
				break
			}
			fmt.Println()

			indices := make([]int64, 0, len(toolCallsMap))
			for idx := range toolCallsMap {
				indices = append(indices, idx)
			}
			slices.Sort(indices)

			assistantMsg := openai.ChatCompletionAssistantMessageParam{}
			if fullContent.Len() > 0 {
				assistantMsg.Content.OfString = openai.String(fullContent.String())
			}

			if len(indices) > 0 {
				toolCallParams := make([]openai.ChatCompletionMessageToolCallParam, 0, len(indices))
				for _, idx := range indices {
					tc := toolCallsMap[idx]
					toolCallParams = append(toolCallParams, openai.ChatCompletionMessageToolCallParam{
						ID:   tc.ID,
						Type: "function",
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: tc.ArgsBuffer.String(),
						},
					})
				}
				assistantMsg.ToolCalls = toolCallParams
			}

			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &assistantMsg,
			})

			if len(toolCallsMap) == 0 {
				break
			}

			for _, idx := range indices {
				tc := toolCallsMap[idx]
				if tc.Name != "create_client" {
					continue
				}

				argsJSON := tc.ArgsBuffer.String()
				fmt.Printf("\n[AGENTE Executing Tool: %s con args: %s]\n", tc.Name, argsJSON)

				resultOutput, errorTools := EjecutarCreateClient(argsJSON)
				if errorTools != nil {
					log.Printf("[AGENTE] error ejecutando tool: %v", errorTools)
				}

				fmt.Printf("[AGENTE Tool Output]: %s\n\n", resultOutput)
			}
		}

		if iteration >= maxAgentIterations {
			fmt.Println("[AGENTE] Límite de iteraciones alcanzado sin respuesta final.")
		}
	}
}

type CreateClientArgs struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Skills   string `json:"skills"`
}

type CreateClientResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

func EjecutarCreateClient(jsonArgs string) (string, error) {
	var args CreateClientArgs
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		resp := CreateClientResponse{
			Status:  "error",
			Message: fmt.Sprintf("Argumentos inválidos: %v", err),
		}
		out, _ := json.Marshal(resp)
		return string(out), fmt.Errorf("error al parsear los argumentos de la tool: %w", err)
	}

	if strings.TrimSpace(args.Name) == "" ||
		strings.TrimSpace(args.Lastname) == "" ||
		strings.TrimSpace(args.Skills) == "" {
		resp := CreateClientResponse{
			Status:  "error",
			Message: "Los campos name, lastname y skills no pueden estar vacíos",
		}
		out, _ := json.Marshal(resp)
		return string(out), fmt.Errorf("campos requeridos vacíos")
	}

	resp := CreateClientResponse{
		Status:  "success",
		Message: fmt.Sprintf("Cliente %s %s registrado correctamente.", args.Name, args.Lastname),
		ID:      "cli_982341",
	}

	out, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

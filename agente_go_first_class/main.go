package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	// Sin argumentos lee la clave de la variable de entorno ANTHROPIC_API_KEY
	client := anthropic.NewClient(option.WithAPIKey(".."))

	// Ahora definamos las herramientas donde el agente deja de ser un simple chat inteligente
	var tools_defined = []anthropic.ToolParam{
		{
			Name: "buscar_pedido",
			Description: anthropic.String(
				"Devuelve el estado y fecha de un pedido." +
					"Usala cuando el usuario necesite un pedido en concreto",
			),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"numero": map[string]any{
						"type":        "string",
						"description": "Numero de pedido con el formato PED-12345",
						"format":      "PED-12345",
					},
				},
				Required: []string{"numero"},
			},
		},
	}

	// 2. Convertimos las herramientas al formato de unión requerido por el SDK
	tools := make([]anthropic.ToolUnionParam, len(tools_defined))
	for i, tp := range tools_defined {
		tools[i] = anthropic.ToolUnionParam{OfTool: &tp}
	}

	msg, err_client := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeFable5,
		MaxTokens: 1024,
		Tools:     tools,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Resumen esto en una linea")),
		},
	})

	if err_client != nil {
		panic(err_client)
	}

	for _, b := range msg.Content {
		if t, ok := b.AsAny().(anthropic.TextBlock); ok {
			fmt.Println(t.Text)
		}

		// Si Claude decide ejecutar una herramienta
		if tu, ok := b.AsAny().(anthropic.ToolUseBlock); ok {
			fmt.Printf("¡Claude quiere usar la herramienta: %s!\n", tu.Name)
			fmt.Printf("Argumentos enviados por el modelo: %v\n", tu.Input)
			// Aquí ejecutarías la lógica real de tu función en Go
			// (por ejemplo, llamar a una API externa de clima con los argumentos recibidos)
		}
	}
}

// Function de llamada con bucle de conversacion

func loop_ai() {
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock("Resumen esto en una linea")),
	}

	for {
		client := anthropic.NewClient(option.WithAPIKey(".."))

		var tools_define = []anthropic.ToolParam{
			{
				Name: "buscar_algo",
				Description: anthropic.String(
					"Esta funcion busca algo en internete algun recurso" +
						"Solo ejecutala cuando quieras buscar algun titulo en internet o algun link o url" +
						"Esta devuelve el html del la url pasada por parametros",
				),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"url": map[string]any{
							"type":        "string",
							"description": "is a url",
							"formated":    "www.facebook.com/home",
						},
					},
					Required: []string{"url"},
				},
			},
		}

		var tool = make([]anthropic.ToolUnionParam, len(tools_define))
		for i, p := range tools_define {
			tool[i] = anthropic.ToolUnionParam{
				OfTool: &p,
			}
		}

		msg, err_cliente := client.Messages.New(
			context.Background(),
			anthropic.MessageNewParams{
				Model:     anthropic.ModelClaudeFable5,
				MaxTokens: 1024,
				Messages:  messages,
				Tools:     tool,
			},
		)

		if err_cliente != nil {
			fmt.Printf("Error al comunicarse con la API: %v\n", err_cliente)
			return
		}
		// Historial de mensajes
		messages = append(messages, msg.ToParam())

		var resultados = []anthropic.ContentBlockParamUnion{}
		for _, b := range msg.Content {
			if uso, ok := b.AsAny().(anthropic.ToolUseBlock); ok {
				uso.JSON.Input.Raw()

				// uso.ID - ID/identificador de la respuesta al resultado
				// uso.Name - Nombre de la funcion a invocar
				// uso.JSON.Input.Raw() - Los parametros vendran serealizados
				salida, fallo := ejecutarFuncion(uso.Name, uso.JSON.Input.Raw()) // Llamada a ejecutar la funcion

				resultados = append(resultados, anthropic.NewToolResultBlock(
					uso.ID,
					salida,
					fallo,
				))
			}
		}

		if len(resultados) == 0 {
			for _, block := range msg.Content {
				if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
					fmt.Printf("\nRespuesta final del Agente:\n%s\n", textBlock.Text)
				}
			}
			break
		}
		messages = append(messages, anthropic.NewUserMessage(resultados...))
	}
}

func loop_ai_strem() {
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock("Resumen esto en una linea")),
	}

	client := anthropic.NewClient(option.WithAPIKey(".."))
	var toolsDefine = []anthropic.ToolParam{
		{
			Name:        "buscar_algo",
			Description: anthropic.String("Esta funcion busca algo en internet algun recurso. Solo ejecutala cuando quieras buscar algun titulo en internet o algun link o url. Devuelve el html de la url pasada por parametros"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "is a url",
					},
				},
				Required: []string{"url"},
			},
		},
	}

	toolUnion := make([]anthropic.ToolUnionParam, len(toolsDefine))
	for i, p := range toolsDefine {
		toolUnion[i] = anthropic.ToolUnionParam{
			OfTool: &p,
		}
	}

	for {
		// Iniciamos el stream
		stream := client.Messages.NewStreaming(
			context.Background(),
			anthropic.MessageNewParams{
				Model:     anthropic.ModelClaudeFable5,
				MaxTokens: 1024,
				Messages:  messages,
				Tools:     toolUnion,
			},
		)

		var msg anthropic.Message

		// Consumimos el stream evento por evento
		for stream.Next() {
			evento := stream.Current()

			// Acumula los datos para construir el mensaje completo en segundo plano
			errAcumulate := msg.Accumulate(evento)
			if errAcumulate != nil {
				fmt.Printf("Error al acumular evento: %v\n", errAcumulate)
				return
			}

			// Mostramos el texto en tiempo real al usuario conforme va llegando
			if e, ok := evento.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
				if d, ok := e.Delta.AsAny().(anthropic.TextDelta); ok {
					fmt.Print(d.Text) // Usamos Print para que el texto fluya sin saltos de línea forzados
				}
			}
		}

		if stream.Err() != nil {
			fmt.Printf("Error en el stream: %v\n", stream.Err())
			return
		}

		// Salto de línea estético al terminar de escribir el bloque de texto
		fmt.Println()

		// 1. Guardamos la respuesta completa ya acumulada en el historial
		messages = append(messages, msg.ToParam())

		// 2. Procesamos si Claude decidió invocar alguna herramienta
		var resultados []anthropic.ContentBlockParamUnion
		for _, b := range msg.Content {
			if uso, ok := b.AsAny().(anthropic.ToolUseBlock); ok {
				fmt.Printf("\n[Agente] Claude quiere usar la herramienta: %s\n", uso.Name)

				// Ejecutamos tu función de negocio local
				salida, fallo := ejecutarFuncion(uso.Name, uso.JSON.Input.Raw())

				// Empaquetamos el resultado usando el ID único provisto por Claude
				toolResultBlock := anthropic.NewToolResultBlock(uso.ID, salida, fallo)
				resultados = append(resultados, toolResultBlock)
			}
		}

		// 3. Condición de salida: Si Claude no pidió ninguna herramienta, la tarea terminó
		if len(resultados) == 0 {
			break
		}

		// 4. Si hubo herramientas, enviamos los resultados de vuelta en un único mensaje de usuario
		messages = append(messages, anthropic.NewUserMessage(resultados...))
	}
}

func ejecutarFuncion(nombre string, rawInput string) (string, bool) {
	switch nombre {
	case "buscar_algo":
		var args struct {
			URL string `json:"url"`
		}

		// Deserializamos el JSON crudo en la estructura de Go
		if err := json.Unmarshal([]byte(rawInput), &args); err != nil {
			return "Error al parsear los argumentos de la herramienta", true
		}

		fmt.Printf("[Ejecutando herramienta] Conectando a la URL: %s\n", args.URL)

		// Aquí realizarías tu petición HTTP real o scraping
		return "<html><body><h1>Contenido simulado de la web</h1><p>Texto de prueba...</p></body></html>", false

	default:
		return "Herramienta no encontrada", true
	}
}

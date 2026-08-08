package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

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
		MaxTokens: 4000,
		Tools:     tools,
		System: []anthropic.TextBlockParam{
			{Text: "INSTRUCCIONES DEL PROMPT SISTEMA PARA LA IA"},
		},
		Thinking: anthropic.ThinkingConfigParamOfEnabled(1024),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Resumen esto en una linea")),
		},
	})

	if err_client != nil {
		var apiError *anthropic.Error
		if errors.As(err_client, apiError) {
			// Identificador de peticion  es lo primero que va a pedir un soporte
			log.Printf("API %d (request id)", apiError.StatusCode, apiError.RequestID)
			switch apiError.StatusCode {
			case 400:
				{
					panic(fmt.Errorf("Error de conexion en el mdelo")) // No se reintenta
				}

			case 429, 500, 529:
				{
					// El sdk ya intenta dos veces
				}
			}
		}

		/*
				Y LOS DOS LIMITES DEL BUCLE QUE TIEES QUE PONER TU:
			- UN MAXIMO DE VUELTAS SINO, UN AGENTE CONFUNDIDO MIRA PARA SIEMPRE
			- UN CONTEXT TIMEOUT, BUCLE INTERNO CADA LLAMADA
		*/
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

func loop_ai_strem_thinkin() {
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
				MaxTokens: 4000,
				Thinking:  anthropic.ThinkingConfigParamOfEnabled(1024),
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
				switch delta := e.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					{
						fmt.Print(delta.Text)
					}

				case anthropic.ThinkingDelta:
					{
						// Opcional: Puedes mostrar el pensamiento interno en tiempo real
						// fmt.Printf("[Pensando...]: %s\n", delta.Thinking)
					}
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

func loop_ai_strem_thinkin_with_rules() {
	// 1. Cliente HTTP personalizado con timeout estricto para evitar bloqueos de red
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	client := anthropic.NewClient(
		option.WithAPIKey("TU_API_KEY"),
		option.WithHTTPClient(httpClient),
	)

	// 2. Historial de mensajes inicial
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock("Busca la información necesaria en la web y haz un resumen claro.")),
	}

	// 3. Definición robusta de herramientas con descripciones detalladas y parámetros requeridos
	toolsDefine := []anthropic.ToolParam{
		{
			Name: "buscar_algo",
			Description: anthropic.String(
				"Busca un recurso o link en internet. " +
					"Úsala únicamente cuando necesites información externa o un recurso web específico. " +
					"Devuelve el contenido HTML o simulado de la URL solicitada.",
			),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "URL completa y válida, por ejemplo: https://www.ejemplo.com",
					},
				},
				Required: []string{"url"},
			},
		},
	}

	toolUnion := make([]anthropic.ToolUnionParam, len(toolsDefine))
	for i, p := range toolsDefine {
		toolUnion[i] = anthropic.ToolUnionParam{OfTool: &p}
	}

	// 4. LÍMITE DE SEGURIDAD: Control de iteraciones máximas para evitar bucles infinitos
	maxIterations := 5
	currentIteration := 0

	for {
		currentIteration++
		if currentIteration > maxIterations {
			fmt.Printf("\n[Seguridad del Agente] Límite máximo de %d iteraciones alcanzado. Notificando al modelo para cierre elegante...\n", maxIterations)

			// 1. Inyectamos un mensaje final informándole al modelo que se quedó sin turnos
			messages = append(messages, anthropic.NewUserMessage(
				anthropic.NewTextBlock("Has alcanzado el límite máximo de iteraciones permitidas. Por favor, no ejecutes más herramientas, haz un breve resumen de lo que lograste hasta ahora y explícale amablemente al usuario que no pudiste completar la tarea por completo."),
			))

			// 2. Hacemos una última llamada rápida (sin herramientas) para que Claude cierre la conversación
			finalStream := client.Messages.NewStreaming(
				context.Background(),
				anthropic.MessageNewParams{
					Model:     anthropic.ModelClaudeFable5,
					MaxTokens: 1024,
					Messages:  messages,
					// Ya no le pasamos herramientas (Tools) para forzarlo a responder en texto
				},
			)

			var finalMsg anthropic.Message
			for finalStream.Next() {
				evento := finalStream.Current()
				finalMsg.Accumulate(evento)
				if e, ok := evento.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
					if delta, ok := e.Delta.AsAny().(anthropic.TextDelta); ok {
						fmt.Print(delta.Text)
					}
				}
			}
			fmt.Println()
			break // Ahora sí, salimos del bucle de forma limpia
		}

		// 5. Configuración de la petición con Streaming, System Prompt y Thinking Activado
		stream := client.Messages.NewStreaming(
			context.Background(),
			anthropic.MessageNewParams{
				Model:     anthropic.ModelClaudeFable5,
				MaxTokens: 4000, // Requerido para dar suficiente espacio al presupuesto de thinking
				Messages:  messages,
				Tools:     toolUnion,
				// System Prompt: La constitución o reglas de comportamiento del agente
				System: []anthropic.TextBlockParam{
					{
						Text: "Eres un agente autónomo experto en backend y análisis de datos. " +
							"Sigue estrictamente estas reglas:\n" +
							"1. Analiza cuidadosamente la solicitud antes de actuar.\n" +
							"2. Si requieres datos externos, invoca las herramientas disponibles de forma precisa.\n" +
							"3. Si una herramienta devuelve un error, lee el reporte y reintenta corrigiendo los parámetros o informa al usuario.\n" +
							"4. Nunca inventes información si una herramienta falla de forma crítica.",
					},
				},
				// Razonamiento Extendido (Thinking)
				Thinking: anthropic.ThinkingConfigParamOfEnabled(1024),
			},
		)

		var msg anthropic.Message

		// Consumo del Stream evento por evento
		for stream.Next() {
			evento := stream.Current()

			// Acumulación en segundo plano del mensaje completo
			if errAcumulate := msg.Accumulate(evento); errAcumulate != nil {
				fmt.Printf("Error al acumular evento del stream: %v\n", errAcumulate)
				return
			}

			// Impresión en tiempo real del texto visible para el usuario
			if e, ok := evento.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
				if delta, ok := e.Delta.AsAny().(anthropic.TextDelta); ok {
					fmt.Print(delta.Text)
				}

				if _, ok := e.Delta.AsAny().(anthropic.ThinkingDelta); ok {
					fmt.Print("Pensando dame un momento")
				}
				// Aca podriamos impromir el razonamiento
			}
		}

		if stream.Err() != nil {
			fmt.Printf("Error crítico en el stream: %v\n", stream.Err())
			return
		}

		fmt.Println() // Salto de línea estético al terminar el bloque de texto

		// 6. Historial: Guardamos la respuesta completa del asistente
		messages = append(messages, msg.ToParam())

		// 7. Procesamiento de bloques especiales (Thinking y ToolUse)
		var resultados []anthropic.ContentBlockParamUnion
		for _, b := range msg.Content {
			switch block := b.AsAny().(type) {
			case anthropic.ThinkingBlock:
				fmt.Printf("\n[Razonamiento interno completado]: %s\n", block.Thinking)

			case anthropic.ToolUseBlock:
				fmt.Printf("\n[Agente] Ejecutando herramienta: %s (ID: %s)\n", block.Name, block.ID)

				// Ejecución segura de la función con manejo de errores y validaciones
				salida, fallo := ejecutarFuncion(block.Name, block.JSON.Input.Raw())

				// Empaquetado robusto atado al ID único de Claude
				toolResultBlock := anthropic.NewToolResultBlock(block.ID, salida, fallo)
				resultados = append(resultados, toolResultBlock)
			}
		}

		// 8. Condición de salida natural: Si Claude no solicitó ninguna herramienta, el trabajo concluyó
		if len(resultados) == 0 {
			fmt.Println("\n[Agente] Tarea finalizada exitosamente.")
			break
		}

		// 9. Envío de resultados de herramientas en un único mensaje agrupado de usuario
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

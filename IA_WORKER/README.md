# openai - Tipos y funciones específicas

## Cliente OpenAI `*openai.Client`

- `client.Chat.Completions`: Manejo de chat y Function Calling.
- `client.Embeddings`: Generación de vectores de embeddings.
- `client.Audio`: Transcripción y traducción (`Transcriptions`, `Translations`, `Speech`).
- `client.Images`: Generación y edición de imágenes.
- `client.Files`: Gestión de archivos para fine-tuning o asistentes.
- `client.Models`: Listado y consulta de modelos disponibles.

### Opciones de configuración del cliente

- `option.WithAPIKey(key string)`: Define la clave de API explícitamente. Si no se pasa esta opción, el SDK busca automáticamente la variable de entorno `OPENAI_API_KEY`.
- `option.WithBaseURL(url string)`: Modifica la URL base. Es esencial para trabajar con proxies, Azure OpenAI, gateways corporativos o runtimes locales compatibles (vLLM, Ollama, LM Studio).
- `option.WithHTTPClient(client *http.Client)`: Permite inyectar un cliente `net/http` personalizado (útil para modificar `Timeout`, configurar TLS o añadir middlewares de red/tracing).
- `option.WithMaxRetries(count int)`: Controla la cantidad de reintentos automáticos en fallos de red o errores HTTP 429 / 5xx (el valor predeterminado es 2).
- `option.WithHeader(key, value string)`: Inyecta encabezados HTTP adicionales en cada request.
- `option.WithOrganization(id string)` / `option.WithProject(id string)`: Define la organización o proyecto específico dentro de OpenAI.

---

## Retorno del cliente `*openai.Client`

### Servicio `client.Chat.Completions`

Este es el sub-servicio más utilizado de la API. Sus entradas se conforman por:

```go
client.Chat.Completions.New(ctx, params, opts...) (*openai.ChatCompletion, error) // GENERACIÓN SINCRONA
client.Chat.Completions.NewStreaming(ctx, params, opts...) *ssestream.Stream[openai.ChatCompletionChunk] // GENERACIÓN SERVER-SENT EVENT
```

Ahora vamos con la estructura de entrada para la síncrona, que se conforma por:

```go
params := openai.ChatCompletionNewParams{
    Model: openai.F(openai.ChatModelGPT4o), // o un string custom: openai.F("gpt-4o")
    Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
        openai.SystemMessage("Eres un asistente técnico."),
        openai.UserMessage("Explícame la arquitectura Hexagonal en Go."),
    }),
    Temperature: openai.F(0.7),
    MaxTokens:   openai.F(int64(1000)),
    Tools: openai.F([]openai.ChatCompletionToolParam{
        {
            Type: openai.F(openai.ChatCompletionToolTypeFunction),
            Function: openai.F(openai.FunctionDefinitionParam{
                Name:        openai.F("getUserData"),
                Description: openai.F("Obtiene datos de un usuario por ID"),
                Parameters:  openai.F(myJsonSchema),
            }),
        },
    }),
}
```

Aquí podemos apreciar la estructura de respuestas síncronas:

```go
type ChatCompletion struct {
    ID                string                 `json:"id"`
    Object            string                 `json:"object"` // "chat.completion"
    Created           int64                  `json:"created"`
    Model             string                 `json:"model"`
    Choices           []ChatCompletionChoice `json:"choices"`
    Usage             CompletionUsage        `json:"usage"`
    SystemFingerprint string                 `json:"system_fingerprint"`
}

type ChatCompletionChoice struct {
    Index        int64                 `json:"index"`
    FinishReason FinishReason          `json:"finish_reason"` // "stop", "length", "tool_calls"
    Message      ChatCompletionMessage `json:"message"`
}

type ChatCompletionMessage struct {
    Role      string                         `json:"role"` // "assistant"
    Content   string                         `json:"content"`
    ToolCalls []ChatCompletionMessageToolCall `json:"tool_calls"`
}
```

Y montamos o corremos el cliente en este ejemplo de streaming:

```go
stream := client.Chat.Completions.NewStreaming(ctx, params)
defer stream.Close()

for stream.Next() {
    chunk := stream.Current()
    if len(chunk.Choices) > 0 {
        fmt.Print(chunk.Choices[0].Delta.Content)
    }
}
if err := stream.Err(); err != nil {
    // Manejo de error de conexión/streaming
}
```

### Servicio `client.Embeddings`

Generación de representaciones vectoriales para búsqueda semántica RAG. Su único método de entrada es:

```go
client.Embeddings.New(ctx, params, opts...) (*openai.CreateEmbeddingResponse, error)
```

Aquí están las estructuras de entrada de métodos y salidas:

```go
// Entrada
params := openai.EmbeddingNewParams{
    Model: openai.F(openai.EmbeddingModelTextEmbedding3Small),
    Input: openai.F(openai.EmbeddingNewParamsInputUnion{
        // Puede ser una cadena individual o un slice []string
        Value: []string{"Texto para vectorizar 1", "Texto para vectorizar 2"},
    }),
    Dimensions: openai.F(int64(512)), // Reducción opcional de dimensiones
}

// Salida: *openai.CreateEmbeddingResponse
type CreateEmbeddingResponse struct {
    Object string      `json:"object"` // "list"
    Data   []Embedding `json:"data"`
    Model  string      `json:"model"`
    Usage  CreateEmbeddingResponseUsage `json:"usage"`
}

type Embedding struct {
    Object    string    `json:"object"` // "embedding"
    Index     int64     `json:"index"`
    Embedding []float64 `json:"embedding"` // El vector resultante
}
```

### Servicio `client.Audio`

Subdividido en tres sub-recursos: `Transcriptions`, `Translations` y `Speech`.

#### a) `client.Audio.Transcriptions`

- Método: `.New(ctx, params) (*openai.AudioTranscription, error)`
- Entrada: `openai.AudioTranscriptionNewParams` (recibe un `io.Reader` para el archivo de audio, el modelo `openai.AudioModelWhisper1`)
- Salida: `openai.AudioTranscription` conteniendo la propiedad `.Text string`

#### b) `client.Audio.Speech`

- Método: `.New(ctx, params) (*http.Response, error)`
- Entrada: `openai.AudioSpeechNewParams`

```go
Model: openai.F(openai.SpeechModelTTS1)
Input: openai.F("Texto a sintetizar")
Voice: openai.F(openai.SpeechVoiceAlloy)
```

### Servicio `client.Images`

Para modelos como DALL-E 3.

Vamos con los métodos principales, que son:

```go
client.Images.Generate(ctx, params) (*openai.ImagesResponse, error)
client.Images.Edit(ctx, params) (*openai.ImagesResponse, error)
client.Images.CreateVariation(ctx, params) (*openai.ImagesResponse, error)
```

Estructuras principales:

```go
// Entrada
params := openai.ImageGenerateParams{
    Prompt:         openai.F("Un logotipo minimalista en vectores de un gopher"),
    Model:          openai.F(openai.CreateImageRequestModelDallE3),
    N:              openai.F(int64(1)),
    Quality:        openai.F(openai.ImageGenerateParamsQualityHD),
    ResponseFormat: openai.F(openai.ImageGenerateParamsResponseFormatURL), // o B64JSON
}

// Salida: *openai.ImagesResponse
type ImagesResponse struct {
    Created int64       `json:"created"`
    Data    []ImageItem `json:"data"`
}

type ImageItem struct {
    URL           string `json:"url"`
    B64JSON       string `json:"b64_json"`
    RevisedPrompt string `json:"revised_prompt"`
}
```

---

## Function Calling
Function Calling permite que el modelo devuelva llamadas a funciones con argumentos validados para que tu backend las ejecute. Por otro lado, Structured Outputs exige al modelo responder obligatoriamente bajo un formato JSON estricto (Strict: true) definido por un esquema.

### Function Calling (Invocaciones de funciones)
El flujo se divide en 3 fases: definición de la función en la petición, recepción e interpretación del ToolCall enviado por el modelo, y devolución del resultado al historial de conversación.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openai/openai-go"
)

// Estructura para deserializar los argumentos que genera el modelo
type GetWeatherArgs struct {
	Location string `json:"location"`
	Unit     string `json:"unit,omitempty"`
}

func main() {
	ctx := context.Background()
	client := openai.NewClient()

	// 1. Esquema JSON que describe los parámetros de la función
	weatherSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "Ciudad y país, ej: Maracaibo, Venezuela",
			},
			"unit": map[string]any{
				"type": "string",
				"enum": []string{"celsius", "fahrenheit"},
			},
		},
		"required":             []string{"location"},
		"additionalProperties": false,
	}

	// Historial de conversación
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("¿Qué clima hace hoy en Maracaibo?"),
	}

	// 2. Primera solicitud al modelo adjuntando las herramientas (Tools)
	params := openai.ChatCompletionNewParams{
		Model:    openai.F(openai.ChatModelGPT4o),
		Messages: openai.F(messages),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.F("get_weather"),
					Description: openai.F("Obtiene las condiciones meteorológicas actuales"),
					Parameters:  openai.F(openai.FunctionParameters(weatherSchema)),
				}),
			},
		}),
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Fatalf("Error en la petición: %v", err)
	}

	assistantMsg := resp.Choices[0].Message

	// 3. Evaluar si el modelo decidió invocar una función
	if len(assistantMsg.ToolCalls) > 0 {
		// Agregamos el mensaje del asistente (con ToolCalls) al historial
		messages = append(messages, assistantMsg)

		for _, toolCall := range assistantMsg.ToolCalls {
			if toolCall.Function.Name == "get_weather" {
				var args GetWeatherArgs
				_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

				fmt.Printf(" [Ejecutando Función] Location: %s | Unit: %s\n", args.Location, args.Unit)

				// Simulación de la respuesta de nuestra BD o API externa
				apiResult := `{"temperature": "32", "unit": "celsius", "condition": "Soleado con alta humedad"}`

				// 4. Retornar el resultado de la función con el rol 'Tool' y el ToolCallID correspondiente
				messages = append(messages, openai.ToolMessage(toolCall.ID, apiResult))
			}
		}

		// 5. Segunda petición al modelo para que genere la respuesta final al usuario
		finalParams := openai.ChatCompletionNewParams{
			Model:    openai.F(openai.ChatModelGPT4o),
			Messages: openai.F(messages),
		}

		finalResp, err := client.Chat.Completions.New(ctx, finalParams)
		if err != nil {
			log.Fatalf("Error en la respuesta final: %v", err)
		}

		fmt.Println("Respuesta final del Asistente:", finalResp.Choices[0].Message.Content)
	}
}
```

### Structured Outputs (Respuesta en JSON Estricto)

Structured Outputs garantiza determinismo total en el formato de salida. A diferencia del modo json_object clásico, utilizar openai.ResponseFormatJSONSchemaParam con Strict: true obliga al modelo a respetar al 100% el esquema suministrado (incluyendo tipos, campos requeridos y ausencia de llaves extras).

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openai/openai-go"
)

// DTO para unmarshaling directo en Go
type SoftwareArchitect struct {
	Name             string   `json:"name"`
	PrimaryLanguage  string   `json:"primary_language"`
	YearsExperience  int      `json:"years_experience"`
	CoreCompetencies []string `json:"core_competencies"`
}

func main() {
	ctx := context.Background()
	client := openai.NewClient()

	// Esquema JSON que define la estructura exacta requerida por la aplicación
	profileSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"primary_language": map[string]any{
				"type": "string",
			},
			"years_experience": map[string]any{
				"type": "integer",
			},
			"core_competencies": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		// En modo Strict: true, TODOS los campos deben incluirse en 'required'
		"required":             []string{"name", "primary_language", "years_experience", "core_competencies"},
		"additionalProperties": false, // Obligatorio para garantizar Structured Outputs
	}

	prompt := "Analiza el perfil: Aaron es un desarrollador backend con 6 años de experiencia especializado en Go, microservicios y bases de datos relacionales."

	params := openai.ChatCompletionNewParams{
		Model: openai.F(openai.ChatModelGPT4o),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Extrae la información relevante en el formato estructurado solicitado."),
			openai.UserMessage(prompt),
		}),
		// Configuración del ResponseFormat con el esquema JSON
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        openai.F("architect_profile"),
					Description: openai.F("Perfil extraído del desarrollador"),
					Schema:      openai.F(profileSchema),
					Strict:      openai.F(true), // Fuerza la adhesión matemática al esquema
				}),
			},
		),
	}

	resp, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// El contenido en resp.Choices[0].Message.Content ya es un JSON sintácticamente válido
	rawJSON := resp.Choices[0].Message.Content

	var architect SoftwareArchitect
	if err := json.Unmarshal([]byte(rawJSON), &architect); err != nil {
		log.Fatalf("Error al mapear a la estructura de Go: %v", err)
	}

	fmt.Printf("Estructura Parseada:\n%+v\n", architect)
}
```

### Puntos clave del SDK openai-go
- `additionalProperties: false`: Es obligatorio declarar este parámetro en el JSON Schema para todos los objetos cuando `Strict: true` está activado; de lo contrario, la API devolverá un error 400.
- Casteo de Uniones: En Go, para pasar opciones complejas como `ResponseFormat`, se utiliza la sintaxis con tipo explícito en el wrapper genérico: `openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](...)`.
- Manejo de Errores de Refusal: Cuando se usan Structured Outputs, si el prompt viola políticas de seguridad, el modelo puede declinar la generación en lugar de violar el esquema. Para verificar esto en producción, se revisa la propiedad `resp.Choices[0].Message.Refusal`.

---

# Timeouts, reintentos automáticos y middlewares HTTP usando net/http con el cliente de openai-go

El control de red en [github.com/openai/openai-go](https://github.com/openai/openai-go) se divide en tres capas principales: la gestión del tiempo de vida con `context.Context`, la estrategia de resiliencia propia del SDK con `option.WithMaxRetries`, y el interceptor de peticiones mediante el patrón `net/http.RoundTripper`.

## Timeouts: Nivel de Contexto vs. Nivel HTTP

En Go existen dos formas complementarias de controlar los límites de tiempo:

- `context.WithTimeout` (Recomendado): Se define por petición. Si la llamada supera el tiempo límite, cancela inmediatamente la petición HTTP subyacente y libera recursos.
- `http.Client.Timeout`: Se establece a nivel del cliente HTTP global. Cubre todo el ciclo de vida de la petición (dial, TLS handshake, envío de headers/body, lectura de respuesta).

```go
// Timeout granular por petición individual
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

resp, err := client.Chat.Completions.New(ctx, params)
```

## Middleware HTTP

Para interceptar, auditar o modificar las peticiones/respuestas HTTP (añadir métricas de OpenTelemetry, tracing de latencia, logging de headers, etc.), se implementa la interfaz `http.RoundTripper`.

```go
type RoundTripper interface {
    RoundTrip(*http.Request) (*http.Response, error)
}
```

El siguiente código muestra cómo combinar un middleware de logging, un cliente HTTP personalizado con timeout global, la configuración de reintentos del SDK y un timeout por petición via Context:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// LoggingTransport implementa net/http.RoundTripper actuando como Middleware
type LoggingTransport struct {
	Wrapped http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Log previo a la ejecución
	log.Printf("[HTTP Outgoing] %s %s", req.Method, req.URL)

	// Inyectar o modificar encabezados globalmente si fuera necesario
	req.Header.Set("X-App-Client", "go-service-v1")

	// Delegar la petición al transport subyacente
	resp, err := t.Wrapped.RoundTrip(req)
	if err != nil {
		log.Printf("[HTTP Error] %s %s -> Err: %v (%v)", req.Method, req.URL, err, time.Since(start))
		return nil, err
	}

	// Log posterior a la respuesta
	log.Printf("[HTTP Incoming] %s %s -> Status: %d %s (%v)",
		req.Method, req.URL, resp.StatusCode, resp.Status, time.Since(start))

	return resp, nil
}

func main() {
	// 1. Instanciar el Middleware envolviendo el transport por defecto
	customTransport := &LoggingTransport{
		Wrapped: http.DefaultTransport,
	}

	// 2. Crear el net/http.Client personalizado
	httpClient := &http.Client{
		Timeout:   30 * time.Second, // Timeout global de socket
		Transport: customTransport,  // Inyección del Middleware
	}

	// 3. Inicializar el SDK de OpenAI con el cliente HTTP y políticas de reintento
	client := openai.NewClient(
		option.WithHTTPClient(httpClient),
		option.WithMaxRetries(3), // Reintentará automáticamente en 429/5xx
	)

	// 4. Crear un Context con Timeout para la llamada específica
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 5. Ejecutar la llamada
	resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.F(openai.ChatModelGPT4oMini),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Responde con la palabra 'PONG'"),
		}),
	})

	if err != nil {
		log.Fatalf("Error en la ejecución: %v", err)
	}

	fmt.Println("Respuesta:", resp.Choices[0].Message.Content)
}
```

---

## Loop Agent

Para hacer un bucle de agentes (agentic loop) robusto en Go, la clave está en tratar los errores y timeouts de las herramientas como contenido de conversación y no como panics o interrupciones del programa. Si una herramienta falla o tarda demasiado, esa falla se empaqueta en el `openai.ToolMessage` para que el modelo entienda qué ocurrió y decida qué hacer.

### Principios para el Manejo de Tools y Bucles

- **Retornar el error explícito al modelo**: Si una base de datos falla o una API da error 500, no abortes la ejecución en Go. Envía el mensaje de error dentro del `ToolMessage`. El modelo leerá "ERROR: conexión rechazada" y buscará una alternativa o le explicará la falla al usuario.
- **Timeout aislado por herramienta**: Envuelve la ejecución de cada tool en su propio `context.WithTimeout`. Si se agota el tiempo, cancela el proceso interno y responde al modelo con un mensaje de timeout.
- **Guardia de turnos máximos (Max Turns Guard)**: Establece un límite de iteraciones (ej. 5 o 8 turnos).
- **Forzar el cierre con `ToolChoice: "none"`**: Si alcanzas el límite de turnos y el modelo intenta llamar a otra herramienta, en lugar de romper abruptamente, fuerza una última petición configurando `ToolChoice` en `none` para obligarlo a redactar una respuesta final en texto explicativo.

### Patrón de ejecución de tools con timeout aislado

Cada llamada a herramienta debe ejecutarse en una goroutine controlada por un contexto con tiempo límite independiente del contexto principal:

```go
func executeToolWithTimeout(parentCtx context.Context, name string, args string) (string, error) {
	// Timeout de 3 segundos exclusivo para la ejecución de esta herramienta
	toolCtx, cancel := context.WithTimeout(parentCtx, 3*time.Second)
	defer cancel()

	type result struct {
		output string
		err    error
	}
	ch := make(chan result, 1)

	go func() {
		out, err := callExternalService(toolCtx, name, args) // Tu lógica real
		ch <- result{output: out, err: err}
	}()

	select {
	case <-toolCtx.Done():
		if errors.Is(toolCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("TIMEOUT: la herramienta '%s' excedió el tiempo límite de 3s", name)
		}
		return "", toolCtx.Err()

	case res := <-ch:
		return res.output, res.err
	}
}
```

### Ejemplo completo

Este ejemplo integra el historial de mensajes, captura timeouts/errores de herramientas, detecta llamadas repetidas (infinite loop protection) y fuerza la respuesta final.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/openai/openai-go"
)

func runAgenticLoop(ctx context.Context, client *openai.Client, userPrompt string) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("Eres un asistente técnico. Si una herramienta responde con un error o timeout, informa al usuario o intenta corregir la consulta."),
		openai.UserMessage(userPrompt),
	}

	const maxTurns = 5
	turn := 0

	// Cache para detectar bucles idénticos (misma función con mismos argumentos)
	historyCalls := make(map[string]int)

	for turn < maxTurns {
		turn++

		// Si es el último turno permitido, forzamos al modelo a responder en texto libre
		var toolChoice openai.F[openai.ChatCompletionNewParamsToolChoiceUnion]
		if turn == maxTurns {
			toolChoice = openai.F[openai.ChatCompletionNewParamsToolChoiceUnion](
				openai.ChatCompletionNewParamsToolChoiceBehavior(openai.ChatCompletionToolChoiceOptionNone),
			)
		}

		params := openai.ChatCompletionNewParams{
			Model:      openai.F(openai.ChatModelGPT4o),
			Messages:   openai.F(messages),
			Tools:      openai.F(availableTools), // Tus herramientas definidas
			ToolChoice: toolChoice,
		}

		resp, err := client.Chat.Completions.New(ctx, params)
		if err != nil {
			log.Fatalf("Error crítico en la API de OpenAI: %v", err)
		}

		msg := resp.Choices[0].Message

		// SI NO HAY TOOL CALLS: El modelo ha terminado de razonar y da su respuesta final
		if len(msg.ToolCalls) == 0 {
			fmt.Println("Respuesta final del modelo:", msg.Content)
			return
		}

		// PASO OBLIGATORIO: Agregar siempre el mensaje del asistente con los ToolCalls al historial
		messages = append(messages, msg)

		// Procesar las herramientas solicitadas por el modelo
		for _, toolCall := range msg.ToolCalls {
			callSignature := toolCall.Function.Name + ":" + toolCall.Function.Arguments
			historyCalls[callSignature]++

			var toolResultContent string

			// Detección de bucle infinito por repetición idéntica
			if historyCalls[callSignature] >= 3 {
				toolResultContent = fmt.Sprintf("ERROR_LOOP_DETECTED: Has ejecutado la herramienta '%s' con los mismos parámetros %d veces seguidas. Detente y responde al usuario explicándole la situación.", toolCall.Function.Name, historyCalls[callSignature])
			} else {
				// Ejecución con timeout aislado
				output, err := executeToolWithTimeout(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
				if err != nil {
					// Convertir el error en mensaje explícito para el modelo
					toolResultContent = fmt.Sprintf("ERROR_TOOL_EXECUTION: %v", err)
				} else {
					toolResultContent = output
				}
			}

			// Inyectar la respuesta/error de la tool asociando el ID correspondiente
			messages = append(messages, openai.ToolMessage(toolCall.ID, toolResultContent))
		}
	}
}
```

### Puntos claves

- **Orden Estricto del Historial**: La API de OpenAI exige que inmediatamente después de un mensaje de rol `assistant` que contenga `tool_calls`, vengan los mensajes de rol `tool` (`openai.ToolMessage`) haciendo match exacto con los `toolCall.ID`.
- **Manejo de Errores Silenciosos**: Si el modelo recibe `ERROR_TOOL_EXECUTION: TIMEOUT...`, no se romperá el programa en Go. En la siguiente iteración el modelo dirá algo como: *"Intenté consultar el sistema pero la herramienta no respondió a tiempo. ¿Deseas que reintente?"*.
- **Control de Costos**: Al pasar `ToolChoiceOptionNone` en el último turno, aseguras que el flujo finalice siempre devolviendo una respuesta textual en lugar de quedar atrapado consumiendo tokens en un ciclo infinito.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
	Ape  string `json:"ape" jsonschema:"the lastaname of person to greet"`
}

type Output struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (
	*mcp.CallToolResult,
	Output,
	error,
) {
	// return nil, Output{}, fmt.Errorf("no autorizado: token inválido o ausente") - ERROR POR CUALQUIER MOTIVO
	return nil, Output{Greeting: "Hi " + input.Name + input.Ape}, nil
}

func SayHi_with_auth(ctx context.Context, req *mcp.CallToolRequest, input Input) (
	*mcp.CallToolResult,
	Output,
	error,
) {
	// 1. Obtener el header de autenticación (ej: "Bearer mi-token-secreto")
	authHeader := req.GetExtra().Header.Get("Authorization")

	// 2. Validar la autenticación
	if authHeader != "Bearer mi-token-secreto" {
		// Si retornas un error, el servidor MCP le responderá al cliente/IA que falló
		return nil, Output{}, fmt.Errorf("no autorizado: token inválido o ausente")
	}

	// 3. Si todo está bien, continúa con la lógica normal
	return nil, Output{Greeting: "Hi " + input.Name + " " + input.Ape}, nil
}

func main() {
	// Creamos un server con una sola herramienta
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)

	// Correr el servidor st In/Out hasta que el cliente se desconecte
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func main_mcp_with_http() {
	server := mcp.NewServer(&mcp.Implementation{Name: "golandia", Version: "1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "great", Description: "Say hi"}, SayHi)

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return server
		},
		nil,
	)

	// Aca se manejan las autenticaciones cuando el servidor o mcp esta expuesto totalmente
	log.Println("Servidor MCP HTTP corriendo en http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

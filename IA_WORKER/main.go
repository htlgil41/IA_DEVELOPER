package main

import (
	"fmt"
	"ia_worker/examples"
	"ia_worker/libs"
	"log"
)

func main() {
	fmt.Println("Worker IA Cloudfare Agente")
	config := libs.LoadConfigVyper()
	if config == nil {
		log.Printf("Erropr al cargar las variables desde el conf")
		return
	}

	// examples.ResponseChat(config)
	// examples.ResponseChatStream(config)
	examples.ChatWitUs(config)
}

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"trikmq/events"
)

type TrikMqInstance struct {
	Config          *TrikMqConfig
	ConnectionCount int
	EventManager    *events.EventManager
}

func main() {
	config := ReadConfig()
	App := &TrikMqInstance{
		Config:          config,
		ConnectionCount: 0,
		EventManager:    events.NewEventManager(),
	}
	App.Start()
}

func (App *TrikMqInstance) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer App.EventManager.Disconnect(conn)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Connection closed:", err)
			return
		}

		message = strings.TrimSpace(message)
		log.Println("Received:", message)

		event, err := App.EventManager.ParseEvent(&message)
		if err != nil {
			log.Println(err)
			continue
		}

		go App.EventManager.HandleEvent(conn, event)
	}
}

func (App *TrikMqInstance) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", App.Config.Port))
	if err != nil {
		errorText := fmt.Sprintf("Failed to listen on port %d: %v", App.Config.Port, err)
		fmt.Println(errorText)
		os.Exit(1)
	}
	defer listener.Close()
	log.Println("TCP server started on port 1102")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Println("Client connected:", conn.RemoteAddr())

		go App.handleConnection(conn) // Handle each connection concurrently
	}
}

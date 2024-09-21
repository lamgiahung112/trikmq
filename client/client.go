package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type EventType int

const (
	SUBSCRIBE EventType = iota
	UNSUBSCRIBE
	BROADCAST
)

type EventHeader struct {
	Type EventType `json:"type"`
}

type Event struct {
	Headers *EventHeader `json:"header"`
	Payload interface{}  `json:"payload"`
}

type BroadCastPayload struct {
	Channel string `json:"channel"`
	Message string `json:"message"`
}

func main() {
	// Connect to the TCP server
	conn, err := net.Dial("tcp", "localhost:1102")
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to server. Type 'hello' to get a response. Type 'exit' to close the connection.")

	// Use bufio to read user input and send it to the server
	reader := bufio.NewReader(os.Stdin)

	go handleServerResponse(conn)
	for {
		text, _ := reader.ReadString('\n')
		// Exit the loop if user types "exit"
		if text == "exit" {
			fmt.Println("Closing connection")
			conn.Close()
			return
		}

		eventType := strings.Split(text, " ")[0]
		args := strings.Split(text, " ")[1:]

		var event *Event
		switch eventType {
		case "subscribe":
			event = &Event{
				Headers: &EventHeader{
					Type: SUBSCRIBE,
				},
				Payload: sanitize(args[0]),
			}
		case "unsubscribe":
			event = &Event{
				Headers: &EventHeader{
					Type: UNSUBSCRIBE,
				},
				Payload: sanitize(args[0]),
			}
		case "broadcast":
			event = &Event{
				Headers: &EventHeader{
					Type: BROADCAST,
				},
				Payload: &BroadCastPayload{
					Channel: args[0],
					Message: sanitize(strings.Join(args[1:], " ")),
				},
			}
		default:
			log.Println("Unknown event type:", eventType)
			event = nil
		}

		msg, err := json.Marshal(event)
		if err != nil {
			continue
		}

		// Send the message to the server
		_, err = conn.Write([]byte(string(msg) + "\n"))
		if err != nil {
			log.Println("Failed to send message:", err)
			conn.Close()
			return
		}
	}
}

func handleServerResponse(conn net.Conn) {
	for {
		// Read the response from the server
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			// Handle EOF, which means the connection is closed
			if err == io.EOF {
				fmt.Println("Server closed the connection.")
				conn.Close()
				return
			}
			// Log other errors and break the loop
			log.Println("Error reading server response:", err)
			conn.Close()
			return
		}

		// Print the server's response
		parsedResponse := fmt.Sprintf(`Message from server: %s`, response)
		log.Println(parsedResponse)
	}
}

func sanitize(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}

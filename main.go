package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/go-kit/log/level"
)

// HTTP Handlers

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := info()
	_ = level.Debug(logging.GlobalLogger()).Log("msg", fmt.Sprintf("Source IP: %s Forwarded-For: %v\n", r.RemoteAddr, r.Header["X-Forwarded-For"]))

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		_ = level.Debug(logging.GlobalLogger()).Log("msg", fmt.Sprintf("ERROR: Failed to encode info response, %s", err))
	}
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		_ = level.Debug(logging.GlobalLogger()).Log("msg", fmt.Sprintf("ERROR: Failed to decode start json, %s", err))
		return
	}

	start(state)

	// Nothing to respond with here
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		_ = level.Debug(logging.GlobalLogger()).Log("msg", fmt.Sprintf("ERROR: Failed to decode move json, %s", err))
		return
	}

	response := move(state)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		_ = level.Debug(logging.GlobalLogger()).Log("msg", fmt.Sprintf("ERROR: Failed to encode move response, %s", err))
		return
	}
}

func HandleEnd(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		_ = level.Debug(logging.GlobalLogger()).Log("msg", fmt.Sprintf("ERROR: Failed to decode end json, %s", err))
		return
	}

	end(state)

	// Nothing to respond with here
}

// Main Entrypoint

func main() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	http.HandleFunc("/", HandleIndex)
	http.HandleFunc("/start", HandleStart)
	http.HandleFunc("/move", HandleMove)
	http.HandleFunc("/end", HandleEnd)

	_ = level.Debug(logging.GlobalLogger()).Log("msg", fmt.Sprintf("Starting Battlesnake Server at http://0.0.0.0:%s...\n", port))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

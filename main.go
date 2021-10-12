package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/Cameron-Kurotori/battlesnake/logging"
	"github.com/Cameron-Kurotori/battlesnake/sdk"
	"github.com/go-kit/log/level"
)

// HTTP Handlers

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := info()
	log.Printf("Source IP: %s Forwarded-For: %v\n", r.RemoteAddr, r.Header["X-Forwarded-For"])

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		_ = level.Error(logging.GlobalLogger()).Log("msg", "failed to encode info response", "err", err)
	}
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	state := sdk.GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		_ = level.Error(logging.GlobalLogger()).Log("msg", "failed to decode start json", "err", err)
		return
	}

	start(state)

	// Nothing to respond with here
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
	state := sdk.GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		_ = level.Error(logging.GlobalLogger()).Log("msg", "failed to decode move json", "err", err)
		return
	}

	response := move(state)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		_ = level.Error(logging.GlobalLogger()).Log("msg", "failed to encode move response", "err", err)
		return
	}
}

func HandleEnd(w http.ResponseWriter, r *http.Request) {
	state := sdk.GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		_ = level.Error(logging.GlobalLogger()).Log("msg", "failed to decode end json", "err", err)
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

	log.Printf("Starting Battlesnake Server at http://0.0.0.0:%s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

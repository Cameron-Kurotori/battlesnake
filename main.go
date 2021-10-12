package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/BattlesnakeOfficial/starter-snake-go/logging"
	"github.com/go-kit/log/level"
)

// HTTP Handlers

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	response := info()
	_ = level.Debug(logging.GlobalLogger()).Log("msg", "index request received", "source_ip", r.RemoteAddr, "forwarded_for", r.Header["X-Forwarded-For"])

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		_ = level.Error(logging.GlobalLogger()).Log("msg", "failed to encode info response", "err", err)
	}
}

func HandleStart(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		_ = level.Error(logging.GlobalLogger()).Log("msg", "failed to decode start json", "err", err)
		return
	}

	start(state)
}

func HandleMove(w http.ResponseWriter, r *http.Request) {
	state := GameState{}
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
	state := GameState{}
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

	_ = level.Info(logging.GlobalLogger()).Log("msg", "starting battlesnake server", "addr", fmt.Sprintf("http://0.0.0.0:%s", port))
	_ = level.Error(logging.GlobalLogger()).Log("msg", "server closed", "err", http.ListenAndServe(":"+port, nil))
}

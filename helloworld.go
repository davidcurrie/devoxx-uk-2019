package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("VERSION")
	var name string
	if r.Method == "POST" {
		name = parseBody(r.Body)
	}
	if name == "" {
		name = "World"
	}
	msg := fmt.Sprintf("%s: Hello %s!\n", version, name)
	log.Print(msg)
	fmt.Fprint(w, msg)
}

func main() {
	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %q\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

type cloudEvent struct {
	Data string `json:"data"`
}

func parseBody(body io.ReadCloser) string {
	var event cloudEvent
	if err := json.NewDecoder(body).Decode(&event); err != nil {
		log.Printf("Failed to decode request body: %v\n", err)
		return ""
	}
	bytes, err := base64.StdEncoding.DecodeString(event.Data)
	if err != nil {
		log.Printf("Failed to decode event data: %v\n", err)
		return ""
	}
	return string(bytes)
}

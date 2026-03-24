package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Rock and roll on the web")
	})

	fmt.Println("server running at localhost:8080")

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		fmt.Println("server failed to start")
	}
}

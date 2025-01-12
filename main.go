package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/generate", generateImage)

	fmt.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

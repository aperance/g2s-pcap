package main

import (
	"fmt"
	"net/http"
)

func main() {
	packets := make(chan string)

	go packetCapture(packets)
	fmt.Println("Capturing HTTP packets...")
	go webSocketServer()
	fmt.Println("Listening on port 3000...")
	for {
		fmt.Printf("\n%s\n\n", <-packets)
	}
}

func webSocketServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})
	http.ListenAndServe(":3000", nil)
}

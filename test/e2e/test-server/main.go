package main

import (
	"log"
	"os"

	// "fmt"
	// "io"
	"net/http"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
	// fmt.Fprintf(w, "This is an example server.\n")
	// io.WriteString(w, "This is an example server.\n")
}

func main() {
	certFile := os.Getenv("CERT_FILE")
	if certFile == "" {
		certFile = "localhost.crt"
	}
	keyFile := os.Getenv("KEY_FILE")
	if keyFile == "" {
		keyFile = "localhost.key"
	}

	http.HandleFunc("/hello", HelloServer)
	go func() {
		err := http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	}()
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
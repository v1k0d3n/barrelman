package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("SWS_PORT")
	if port == "" {
		port = "8080"
		log.Printf("SWS_PORT not set. Using default %s.", port)
	}

	appName := os.Getenv("SWS_APP_NAME")
	if appName == "" {
		appName = "Barrelman"
		log.Printf("SWS_APP_NAME not set. Using default %s.", appName)
	}

	host := os.Getenv("SWS_HOST")
	if host == "" {
		host = "localhost"
		log.Printf("SWS_HOST not set. Using default %s.", host)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request received for /")
		fmt.Fprintf(w, "Hello, %s. You are running on %s.\n", appName, host)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

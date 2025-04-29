package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	RunServer()
}

func RunServer() {
	msg := os.Getenv("COUNTER")
	port := os.Getenv("PORT")
	addres := fmt.Sprintf(":%s", port)
	rm := http.NewServeMux()
	rm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("Hello world %s", msg)
		w.Write([]byte(msg))
	})
	rm.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := http.Server{
		Addr:    addres,
		Handler: rm,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

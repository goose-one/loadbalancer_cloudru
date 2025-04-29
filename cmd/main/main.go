package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"loadbalancer/internal/app"
)

func main() {
	fmt.Println("app starting")

	server, err := app.NewApp(os.Getenv("CONFIG_FILE"))
	if err != nil {
		panic(err)
	}

	go func() {
		if err := server.Run(); err != nil {
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("app stopping")
	if err := server.Shutdown(); err != nil {
		fmt.Printf("ServerHTTP shutdown error: %v\n", err)
	}
}

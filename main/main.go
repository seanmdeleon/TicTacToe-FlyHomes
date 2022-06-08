package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/rs/cors"
	"github.com/seanmdeleon/TicTacToe-FlyHomes/pkg/apiresources"
)

// This is the entry point for the HTTP server
// Locally, this program will run using localhost:8080 and can be stopped by performing a Ctrl+C
// In a more sophisticated design that scales, I would serve this in a Lambda function that gets invoked via APIGateway
func main() {

	fmt.Println("TicTacToe HTTP server has been invoked locally")
	fmt.Println("Press CTR+C to quit")

	router := apiresources.GetRouter()

	srv := http.Server{
		Addr:    ":" + strconv.Itoa(8080),
		Handler: cors.Default().Handler(apiresources.CaselessMatcher(router)),
	}

	// create a seperate go routine that listens for the user to shut down the server using Ctrl+C
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)

		sig := <-sigint
		fmt.Printf("Received %v signal\n", sig)

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			fmt.Printf("HTTP server Shutdown: %v\n", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Starting local server")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v\n", err)
	}
	log.Printf("Shutdown successfully")

	<-idleConnsClosed
}

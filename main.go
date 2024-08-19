package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"packagelock/config"
	"packagelock/server"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Linker Injections
// Version injection with Docker Build & ldflags
// Do not modify, init or change in code!
var AppVersion string

// TODO: support for multiple network adapters.

func main() {
	// Start Viper for config management
	Config := config.StartViper(viper.New())

	// If AppVersion is injected, set it in the configuration
	if AppVersion != "" {
		Config.SetDefault("general.app-version", AppVersion)
	}

	fmt.Println(Config.AllSettings())

	// Channel to signal the restart
	restartChan := make(chan struct{})
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		for {
			// Add Fiber routes
			router := server.AddRoutes()

			// Fiber does not use the standard http.Server
			// Setup server address from config
			serverAddr := Config.GetString("network.fqdn") + ":" + Config.GetString("network.port")

			// Fiber specific server start
			go func() {
				fmt.Printf("Starting Fiber server at %s...\n", serverAddr)
				if err := router.Router.Listen(serverAddr); err != nil {
					fmt.Printf("Server error: %s\n", err)
				}
			}()

			// Wait for either a restart signal or termination signal
			select {
			case <-restartChan:
				fmt.Println("Restarting Fiber server...")

				// Gracefully shutdown the Fiber server
				_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := router.Router.Shutdown(); err != nil {
					fmt.Printf("Server shutdown failed: %v\n", err)
				} else {
					fmt.Println("Server stopped.")
				}

			case <-quitChan:
				fmt.Println("Shutting down Fiber server...")

				// Gracefully shutdown on quit signal
				_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := router.Router.Shutdown(); err != nil {
					fmt.Printf("Server shutdown failed: %v\n", err)
				} else {
					fmt.Println("Server stopped gracefully.")
				}
				return
			}
		}
	}()

	// Watch for configuration changes
	Config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		fmt.Println("Restarting to apply changes...")
		restartChan <- struct{}{} // Send signal to restart the server
	})
	Config.WatchConfig()

	// Block until quit signal is received
	<-quitChan
	fmt.Println("Main process exiting.")
}

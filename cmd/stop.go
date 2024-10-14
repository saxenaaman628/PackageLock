package cmd

import (
	"context"
	"fmt"
	"os"
	"packagelock/logger"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewStopCmd creates the stop command.
func NewStopCmd(rootParams RootParams) *cobra.Command {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running server",
		Run: func(cmd *cobra.Command, args []string) {
			app := fx.New(
				fx.Supply(rootParams),
				logger.Module,
				fx.Invoke(stopServer),
			)

			if err := app.Start(context.Background()); err != nil {
				rootParams.Logger.Fatal("Failed to start application for stop command", zap.Error(err))
			}

			if err := app.Stop(context.Background()); err != nil {
				rootParams.Logger.Fatal("Failed to stop application after stop command", zap.Error(err))
			}
		},
	}

	return stopCmd
}

func stopServer(logger *zap.Logger) {
	// Read the PID from the file
	data, err := os.ReadFile("packagelock.pid")
	if err != nil {
		logger.Fatal("Could not read PID file", zap.Error(err))
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		logger.Fatal("Invalid PID found in file", zap.Error(err))
	}

	// Send SIGTERM to the process
	fmt.Printf("Stopping the server with PID: %d\n", pid)
	logger.Info("Stopping the server", zap.Int("PID", pid))

	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		logger.Warn("Failed to stop the server", zap.Error(err))
		return
	}

	fmt.Println("Server stopped.")
	logger.Info("Server stopped.")

	// Remove the PID file
	err = os.Remove("packagelock.pid")
	if err != nil {
		logger.Warn("Failed to remove PID file", zap.Error(err))
	} else {
		fmt.Println("PID file removed successfully.")
		logger.Info("PID file removed successfully.")
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/fang"
	ahab "github.com/josh-allan/ahab/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ahab",
	Short: "Ahoy, Ahab!",
	Long:  "Ahab is a tool to manage Docker Compose files.",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all Docker Compose files",
	Long:  "This command finds all Docker Compose files in the specified directory and starts them.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ahab.RunAllCompose(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all Docker Compose files",
	Long:  "This command finds all Docker Compose files in the specified directory and updates them.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ahab.UpdateAllCompose(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all Docker Compose files",
	Long:  "This command finds all Docker Compose files in the specified directory and stops them.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ahab.StopAllCompose(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart all Docker Compose files",
	Long:  "This command finds all Docker Compose files in the specified directory and restarts them.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ahab.RestartAllCompose(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Docker Compose files",
	Long:  "This command lists all Docker Compose files found in the specified directory.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ahab.ListIgnoreFiles(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(listCmd)
}

func main() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

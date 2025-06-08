package main

import (
	"fmt"
	"log"
	"os"

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

func init() {
    rootCmd.AddCommand(startCmd)
    rootCmd.AddCommand(updateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

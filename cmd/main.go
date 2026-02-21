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

func composeCommand(use, short string, fn func() error) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			if err := fn(); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
	}
}

func init() {
	rootCmd.AddCommand(composeCommand("start", "Start all Docker Compose files", ahab.RunAllCompose))
	rootCmd.AddCommand(composeCommand("update", "Update all Docker Compose files", ahab.UpdateAllCompose))
	rootCmd.AddCommand(composeCommand("stop", "Stop all Docker Compose files", ahab.StopAllCompose))
	rootCmd.AddCommand(composeCommand("restart", "Restart all Docker Compose files", ahab.RestartAllCompose))
	rootCmd.AddCommand(composeCommand("list", "List all Docker Compose files", ahab.ListIgnoreFiles))
}

func main() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/josh-allan/ahab/internal/tui"
	ahab "github.com/josh-allan/ahab/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ahab",
	Short: "Ahoy, Ahab!",
	Long:  "Ahab manages Docker Compose stacks in bulk. Bulk commands act on the stacks listed in $DOCKER_DIR/.ahabstacks; pass compose files to target others.",
}

func exitOnError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func actionCommand(a ahab.Action, short string) *cobra.Command {
	var dockerContext string
	cmd := &cobra.Command{
		Use:   a.Name + " [compose-file...]",
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			exitOnError(ahab.Run(a, args, dockerContext))
		},
	}
	cmd.Flags().StringVar(&dockerContext, "context", "", "docker context for the given files (default: catalogue entry, else current context)")
	return cmd
}

func init() {
	rootCmd.AddCommand(
		actionCommand(ahab.Start, "Start catalogued stacks (up -d)"),
		actionCommand(ahab.Update, "Pull, rebuild and recreate catalogued stacks"),
		actionCommand(ahab.Stop, "Stop catalogued stacks"),
		actionCommand(ahab.Down, "Stop and remove catalogued stacks"),
		actionCommand(ahab.Restart, "Restart catalogued stacks"),
		&cobra.Command{
			Use:   "list",
			Short: "List catalogued stacks and their contexts",
			Run:   func(cmd *cobra.Command, args []string) { exitOnError(ahab.List()) },
		},
		&cobra.Command{
			Use:   "snapshot",
			Short: "Regenerate .ahabstacks from the compose stacks running in every docker context",
			Run:   func(cmd *cobra.Command, args []string) { exitOnError(ahab.Snapshot()) },
		},
		&cobra.Command{
			Use:   "tui",
			Short: "Browse and manage catalogued stacks interactively",
			Run:   func(cmd *cobra.Command, args []string) { exitOnError(tui.Run()) },
		},
	)
}

func main() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

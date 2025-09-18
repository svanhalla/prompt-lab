package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/svanhalla/prompt-lab/greetd/internal/storage"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set application data",
}

var setMessageCmd = &cobra.Command{
	Use:   "message <text>",
	Short: "Set the message that the API and Web UI will serve",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfigAndLogger()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		message := strings.Join(args, " ")
		if strings.TrimSpace(message) == "" {
			fmt.Println("Error: message cannot be empty")
			return
		}

		store := storage.NewMessageStore(cfg.DataPath)
		if err := store.Load(); err != nil {
			fmt.Printf("Error loading message store: %v\n", err)
			return
		}

		if err := store.SetMessage(message); err != nil {
			fmt.Printf("Error setting message: %v\n", err)
			return
		}

		fmt.Printf("Message set to: %s\n", message)
	},
}

func init() {
	setCmd.AddCommand(setMessageCmd)
	rootCmd.AddCommand(setCmd)
}

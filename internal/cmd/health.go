package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/svanhalla/prompt-lab/greetd/internal/version"
)

type HealthInfo struct {
	Status    string       `json:"status"`
	Version   version.Info `json:"version"`
	Timestamp time.Time    `json:"timestamp"`
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Print application health information",
	Run: func(cmd *cobra.Command, args []string) {
		health := HealthInfo{
			Status:    "ok",
			Version:   version.Get(),
			Timestamp: time.Now(),
		}

		output, err := json.MarshalIndent(health, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling health info: %v\n", err)
			return
		}

		fmt.Println(string(output))
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

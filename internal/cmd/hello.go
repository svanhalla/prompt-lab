package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	name string
)

var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Print a friendly greeting",
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			name = "World"
		}
		fmt.Printf("Hello, %s!\n", name)
	},
}

func init() {
	helloCmd.Flags().StringVar(&name, "name", "", "name to greet")
	rootCmd.AddCommand(helloCmd)
}

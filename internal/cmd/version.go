package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/svanhalla/prompt-lab/greetd/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		info := version.Get()
		fmt.Println(info.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

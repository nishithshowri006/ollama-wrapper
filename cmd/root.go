package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "tuiapp",
	Short: "Tuiapp is ollama with pretty features",
	Long:  "Ollama wrapper with good tui functionality",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

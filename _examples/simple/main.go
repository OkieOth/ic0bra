package main

import (
	"fmt"

	"github.com/okieoth/ic0bra"
	"github.com/okieoth/ic0bra_examples/helper"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "simple",
	Short: "Simple example for ic0bra integration",
	Long:  `Simple example for ic0bra integration for providing an interactive option for command line tools`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmdToCall, err := ic0bra.RunInteractive(cmd); err == nil {
			if cmdToCall != nil {
				cmdToCall.Run(cmdToCall, args)
			}
		} else {
			fmt.Println("error while running in interactive mode:", err)
		}
	},
}

func init() {
	helper.AddSubCommands(rootCmd)
}

func main() {
	fmt.Println("I am an example with history for the interactive feature around the cobra lib.")
	rootCmd.Execute()
}

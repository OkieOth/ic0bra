package main

import (
	"fmt"

	"github.com/okieoth/ic0bra"
	"github.com/okieoth/ic0bra_examples/helper"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "simple",
	Short: "Example for ic0bra integration with flag history",
	Long:  `Example for ic0bra integration with flag history to provide an advanced interactive option for command line tools`,
	Run: func(cmd *cobra.Command, args []string) {
		// the history will be stored in this case in ~/.config/ic0bra/history/*
		if cmdToCall, err := ic0bra.RunInteractiveWithHistory(cmd, "ic0bra"); err == nil {
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
	fmt.Println("I am an advanced example for the use of the basic interactive feature around the cobra lib.")
	rootCmd.Execute()
}

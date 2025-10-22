package helper

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Cmd1_1 = &cobra.Command{
	Use:   "cmd1_1",
	Short: "Example sub command",
	Long:  `Example sub command to show interactive feature`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nCommand was called:", cmd.Use)
	},
}

var Cmd1_2 = &cobra.Command{
	Use:   "cmd1_2",
	Short: "Second sub command",
	Long:  `Second sub command to show interactive feature`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nCommand was called:", cmd.Use)
	},
}

var Cmd1_3 = &cobra.Command{
	Use:   "cmd1_3",
	Short: "Third sub command",
	Long:  `Third sub command to show interactive feature`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nCommand was called:", cmd.Use)
	},
}

func AddSubCommands(cmd *cobra.Command) {
	cmd.AddCommand(Cmd1_1)
	cmd.AddCommand(Cmd1_2)
	cmd.AddCommand(Cmd1_3)
}

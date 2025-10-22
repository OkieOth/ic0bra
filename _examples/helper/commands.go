package helper

import (
	"fmt"

	"github.com/spf13/cobra"
)

var PB bool
var PBSlice []bool
var PFloat float32
var PFloatSlice []float32
var PInt int32
var PIntSlice []int32
var PString string
var PStringSlice []string

var FB bool
var FBSlice []bool
var FFloat float32
var FFloatSlice []float32
var FInt int32
var FIntSlice []int32
var FString string
var FStringSlice []string

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

var Cmd2_1 = &cobra.Command{
	Use:   "cmd2_1",
	Short: "Example second level sub command",
	Long:  `Example second level sub command to show interactive feature`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nCommand was called:", cmd.Use)
	},
}

var Cmd2_2 = &cobra.Command{
	Use:   "cmd2_2",
	Short: "Just another second level sub command",
	Long:  `Just another second level sub command to show interactive feature`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nCommand was called:", cmd.Use)
	},
}

var Cmd2_3 = &cobra.Command{
	Use:   "cmd2_3",
	Short: "One more second level sub command",
	Long:  `One more second level sub command to show interactive feature`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nCommand was called:", cmd.Use)
	},
}

var Cmd2_4 = &cobra.Command{
	Use:   "cmd2_4",
	Short: "Forth second level sub command",
	Long:  `Forth second level sub command to show interactive feature`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\nCommand was called:", cmd.Use)
	},
}

func AddSubCommands(cmd *cobra.Command) {
	cmd.AddCommand(Cmd1_1)
	cmd.AddCommand(Cmd1_2)
	cmd.AddCommand(Cmd1_3)

	cmd.PersistentFlags().BoolVar(&PB, "pb", false, "a persistend flag bool example")
	cmd.PersistentFlags().BoolSliceVar(&PBSlice, "pbs", []bool{}, "a persistend flag bool slice example")
	cmd.PersistentFlags().Float32Var(&PFloat, "pf", 0.0, "a persistend flag float example")
	cmd.PersistentFlags().Float32SliceVar(&PFloatSlice, "pfs", []float32{}, "a persistend flag float slice example")
	cmd.PersistentFlags().Int32Var(&PInt, "pi", 0, "a persistend flag int example")
	cmd.PersistentFlags().Int32SliceVar(&PIntSlice, "pis", []int32{}, "a persistend flag int slice example")
	cmd.PersistentFlags().StringVar(&PString, "ps", "", "a persistend flag string example")
	cmd.PersistentFlags().StringSliceVar(&PStringSlice, "pss", []string{}, "a persistend flag string slice example")

	cmd.MarkFlagRequired("ps")
}

func initFlags1(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&FB, "fb", false, "a flag bool example")
	cmd.Flags().BoolSliceVar(&FBSlice, "fbs", []bool{}, "a flag bool slice example")
	cmd.Flags().Float32Var(&FFloat, "ff", 0.0, "a flag float example")
	cmd.Flags().Float32SliceVar(&FFloatSlice, "ffs", []float32{}, "a flag float slice example")
}

func initFlags2(cmd *cobra.Command) {
	cmd.Flags().Int32Var(&FInt, "fi", 0, "a flag int example")
	cmd.Flags().Int32SliceVar(&FIntSlice, "fis", []int32{}, "a flag int slice example")
	cmd.Flags().StringVar(&FString, "fs", "", "a flag string example")
	cmd.Flags().StringSliceVar(&FStringSlice, "fss", []string{}, "a flag string slice example")

	cmd.MarkFlagRequired("fi")
}

func init() {
	Cmd1_1.AddCommand(Cmd2_1)
	Cmd1_1.AddCommand(Cmd2_2)
	Cmd1_1.AddCommand(Cmd2_3)

	Cmd1_2.AddCommand(Cmd2_1)
	Cmd1_2.AddCommand(Cmd2_2)
	Cmd1_2.AddCommand(Cmd2_3)
	Cmd1_2.AddCommand(Cmd2_4)

	Cmd1_3.AddCommand(Cmd2_3)
	Cmd1_3.AddCommand(Cmd2_4)

	initFlags1(Cmd2_1)
	initFlags2(Cmd2_2)
	initFlags1(Cmd2_3)
	initFlags2(Cmd2_4)
}

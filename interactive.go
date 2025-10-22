package ic0bra

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// provides the user input from stdio in default - separated for better testability
var readerFactory = func() *bufio.Reader {
	return bufio.NewReader(os.Stdin)
}

const SELECT_SUB_CMD_PROMPT = "Select the sub command to use: "

// provides the user selection for the command - separated for better testability
var selectionFactory = func(promptString string, options []string) (string, error) {
	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string { return options[i] },
		fuzzyfinder.WithPromptString(promptString),
	)
	if err != nil {
		return "", err
	}
	return options[idx], nil
}

// This function enables a fuzzy style interactive execution, without
// passing all required sub commands and flags at start time.
// cmd - cobra root command
func RunInteractive(cmd *cobra.Command) (*cobra.Command, error) {
	return runInteractiveImpl(cmd, nil)
}

// This function enables a fuzzy style interactive execution, without
// passing all required sub commands and flags at start time.
// In case that there are flags configured it proposes input from a
// history
// cmd - cobra root command
// appName - used as entry directory in the user config folder to store the history values
func RunInteractiveWithHistory(cmd *cobra.Command, appName string) (*cobra.Command, error) {
	histProvider, err := NewFileHistoryProvider(appName)
	if err != nil {
		return nil, err
	}
	return runInteractiveImpl(cmd, histProvider)
}

type HistoryProvider interface {
	InputFromHist(flagName, txt string) (string, error)
	HasHist(flagName string) bool
	SaveHist(flagName, value string) error
}

func runInteractiveImpl(cmd *cobra.Command, histProvider HistoryProvider) (*cobra.Command, error) {
	subCommands := cmd.Commands()
	if len(subCommands) == 0 {
		return nil, fmt.Errorf("command has no sub commads")
	}
	currentCmd := cmd
	for {
		options := getOptionsFromCommands(subCommands...)
		selected, err := selectionFactory(SELECT_SUB_CMD_PROMPT, options)
		if err != nil {
			return nil, fmt.Errorf("error in interactive run: %v", err)
		}
		if strings.HasPrefix(selected, "help ") {
			helpCmd, _, err := currentCmd.Find([]string{"help"})
			return helpCmd, err
		}
		nextCmd, _, err := currentCmd.Find([]string{selected})
		if err != nil {
			return nil, fmt.Errorf("error finding seleted sub command: %v", err)
		}
		subCommands = nextCmd.Commands()
		if len(subCommands) == 0 {
			// reached end of the chain ..
			cmdChain, txt := getCommandChain(nextCmd)
			configuredFlags := setFlagsForCommands(txt, histProvider, cmdChain...)
			fmt.Printf("\nresulting program call:\n\n  %s %s\n", txt, configuredFlags)
			if !shouldContinue() {
				fmt.Println("Cancel.")
				return nil, nil
			}
			return nextCmd, nil
		}
		currentCmd = nextCmd
	}
}

// Provides the chain of all included sub commands for a given command
func getCommandChain(cmd *cobra.Command) ([]*cobra.Command, string) {
	txt := ""
	chain := make([]*cobra.Command, 0)
	for c := cmd; c != nil; c = c.Parent() {
		chain = append(chain, c)
		if txt != "" {
			txt = c.Use + " " + txt
		} else {
			txt = c.Use
		}
	}
	return chain, txt
}

// Returns true if the given flag is marked as required
func isFlagRequired(f *pflag.Flag) bool {
	if f.Annotations != nil {
		if v, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(v) > 0 && v[0] == "true" {
			return true
		}
	}
	return false
}

func isRepeatableFlag(f *pflag.Flag) bool {
	switch f.Value.Type() {
	case "stringSlice", "intSlice", "float32Slice", "boolSlice", "stringArray":
		return true
	default:
		return false
	}
}

func shouldContinue() bool {
	reader := readerFactory()
	count := 0
	maxCount := 10
	for {
		fmt.Printf("\nShould the program execution be continued (default is yes)? [yes|no]: ")
		input, _ := reader.ReadString('\n') // read entire line
		input = strings.TrimSpace(input)    // remove newline and spaces
		if len(input) == 0 {
			return true
		}
		input = strings.ToLower(input)
		switch input {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			fmt.Println("wrong input ... only [yes|no|empty] are allowed!")
		}
		count++
		if count == maxCount {
			fmt.Println("I am tired of it ... programm execution is canceled!")
			return false
		}
	}
}

// iterates over the selected commands and collects input for their configured flags
func setFlagsForCommands(cmdChain string, histProvider HistoryProvider, cmds ...*cobra.Command) string {
	showedChain := false
	configuredFlags := ""
	reader := readerFactory()
	collectRepeatedFlagInputFunc := collectRepeatedFlagInput
	collectFlagInputFunc := collectFlagInput
	if histProvider != nil {
		collectFlagInputFunc = func(cmd *cobra.Command, f *pflag.Flag, flagRequired bool, defValue string, reader *bufio.Reader) string {
			return collectFlagInputWithHist(cmd, f, flagRequired, defValue, reader, histProvider)
		}
		collectRepeatedFlagInputFunc = func(cmd *cobra.Command, f *pflag.Flag, defValue string, reader *bufio.Reader) string {
			return collectRepeatedFlagInputWithHist(cmd, f, defValue, reader, histProvider)
		}
	}
	for _, cmd := range cmds {
		// Iterate over all flags of the command
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			// Ask interactively for input
			if f.Name == "help" {
				return
			}

			if !showedChain {
				fmt.Printf("\n`%s` will be called.\n\nIn the following steps the possible flags will be collected ...\n", cmdChain)
				showedChain = true
			}

			defValue := ""
			flagRequired := isFlagRequired(f)
			if flagRequired {
				defValue = "(required)" //"press ⏎ to skip"
			} else {
				if f.DefValue != "" {
					defValue = fmt.Sprintf("(default %v)", f.DefValue)
				}
			}
			if isRepeatableFlag(f) {
				configuredFlags += collectRepeatedFlagInputFunc(cmd, f, defValue, reader)
			} else {
				configuredFlags += collectFlagInputFunc(cmd, f, flagRequired, defValue, reader)
			}
		})
	}
	return configuredFlags
}

const HELP = "?"
const HELP2 = "help"
const HELP3 = "--help"

// implements the user interaction to get the required input for a flag
func collectFlagInput(cmd *cobra.Command, f *pflag.Flag, flagRequired bool, defValue string, reader *bufio.Reader) string {
	var setValue string
	for {
		fmt.Printf("\n--%s: %s: ", f.Name, defValue)
		input, _ := reader.ReadString('\n') // read entire line
		input = trimInput(input)            // remove newline and spaces

		if input != "" {
			if input == HELP || input == HELP2 || input == HELP3 {
				fmt.Printf("  Expected input: %s\n", f.Usage)
			} else {
				// User provided a value -> set it
				if err := cmd.Flags().Set(f.Name, input); err != nil {
					fmt.Printf("⚠️  Could not set flag %s: %v\n", f.Name, err)
				} else {
					fmt.Printf("  Set value: %s\n", input)
					setValue = input
					break
				}
			}
		} else if flagRequired {
			fmt.Printf("⚠️  Flag %s is required, so input is needed!\n", f.Name)
		} else {
			break
		}
	}
	if setValue != "" {
		return flagTxt(f, setValue)
	} else {
		return setValue
	}
}

func getHistHint(flagName string, histProv HistoryProvider) (bool, string) {
	if histProv == nil {
		return false, ""
	}
	if histProv.HasHist(flagName) {
		return true, " [enter '?' for history]"
	}
	return false, ""
}

func collectRepeatedFlagInput(cmd *cobra.Command, f *pflag.Flag, defValue string, reader *bufio.Reader) string {
	var setValue string
	bFirst := true
	for {
		if bFirst {
			fmt.Printf("\n--%s %s\nmultiple values possible: ", f.Name, defValue)
			bFirst = false
		} else {
			fmt.Printf("\nnext value, empty input to finish: ")
		}
		input, _ := reader.ReadString('\n') // read entire line
		input = trimInput(input)            // remove newline and spaces

		if input != "" {
			if input == HELP || input == HELP2 || input == HELP3 {
				fmt.Printf("  Expected input: %s\n", f.Usage)
			} else {
				// User provided a value -> set it
				if err := cmd.Flags().Set(f.Name, input); err != nil {
					fmt.Printf("⚠️  Could not set flag %s: %v\n", f.Name, err)
				} else {
					fmt.Printf("  Set value: %s\n", input)
					setValue += flagTxt(f, input)
				}
			}
		} else {
			break
		}
	}
	return setValue
}

func getHistInput(f *pflag.Flag, hasHist bool, reader *bufio.Reader, histProvider HistoryProvider, defValue string) string {
	input, _ := reader.ReadString('\n') // read entire line
	input = trimInput(input)
	if hasHist && (input == "?") {
		input, _ = histProvider.InputFromHist(f.Name, fmt.Sprintf("\n'--%s' %s (%s)", f.Name, defValue, f.Usage))
		input = trimInput(input) // remove newline and spaces
	}
	return input
}

func collectFlagInputWithHist(cmd *cobra.Command, f *pflag.Flag, flagRequired bool, defValue string, reader *bufio.Reader, histProvider HistoryProvider) string {
	var setValue string
	for {
		hasHist, histHint := getHistHint(f.Name, histProvider)
		fmt.Printf("\n--%s %s%s: ", f.Name, defValue, histHint)
		input := getHistInput(f, hasHist, reader, histProvider, defValue)

		if input != "" {
			if input == HELP || input == HELP2 || input == HELP3 {
				fmt.Printf("  Expected input: %s\n", f.Usage)
			} else {
				// User provided a value -> set it
				if err := cmd.Flags().Set(f.Name, input); err != nil {
					fmt.Printf("⚠️  Could not set flag %s: %v\n", f.Name, err)
				} else {
					fmt.Printf("  Set value: %s\n", input)
					setValue = input
					histProvider.SaveHist(f.Name, setValue)
					break
				}
			}
		} else if flagRequired {
			fmt.Printf("⚠️  Flag %s is required, so input is needed!\n", f.Name)
		} else {
			break
		}
	}
	if setValue != "" {
		return flagTxt(f, setValue)
	} else {
		return setValue
	}
}

func collectRepeatedFlagInputWithHist(cmd *cobra.Command, f *pflag.Flag, defValue string, reader *bufio.Reader, histProvider HistoryProvider) string {
	var setValue string
	bFirst := true
	hasHist, histHint := getHistHint(f.Name, histProvider)
	for {
		if bFirst {
			fmt.Printf("\n--%s %s%s\nmultiple values possible: ", f.Name, defValue, histHint)
			bFirst = false
		} else {
			fmt.Printf("\nnext value, empty input to finish: ")
		}
		input := getHistInput(f, hasHist, reader, histProvider, defValue)

		if input != "" {
			if input == HELP || input == HELP2 || input == HELP3 {
				fmt.Printf("  Expected input: %s\n", f.Usage)
			} else {
				// User provided a value -> set it
				if err := cmd.Flags().Set(f.Name, input); err != nil {
					fmt.Printf("⚠️  Could not set flag %s: %v\n", f.Name, err)
				} else {
					fmt.Printf("  Set value: %s\n", input)
					histProvider.SaveHist(f.Name, input)
					setValue += flagTxt(f, input)
				}
			}
		} else {
			break
		}
	}
	return setValue
}

func flagTxt(f *pflag.Flag, value string) string {
	escapedValue := strings.ReplaceAll(value, " ", "\\ ")
	return fmt.Sprintf(" --%s %s", f.Name, escapedValue)
}

func trimInput(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\t", " ")
	ret := value
	for {
		tmp := strings.ReplaceAll(ret, "  ", " ")
		if tmp == ret {
			return ret
		}
		ret = tmp
	}
}

// produces the fuzzy search input for interactive command selection
func getOptionsFromCommands(cmds ...*cobra.Command) []string {
	ret := make([]string, 0)
	for _, cmd := range cmds {
		if cmd.Use == "completion" {
			continue
		}
		ret = append(ret, cmd.Use)
	}
	return ret
}

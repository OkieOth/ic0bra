package ic0bra

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
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
	InputFromHist(flagName, txt string, ignoreTxt []string, maxFlags, currentFlag int) (string, error)
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
			printInfo("\nresulting program call:\n\n")
			color.Yellow("  %s %s\n", txt, configuredFlags)
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
		printInfo("\nShould the program execution be continued (default is yes)? [yes|no]: ")
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
		collectFlagInputFunc = func(cmd *cobra.Command, f *pflag.Flag, flagRequired bool, defValue string, reader *bufio.Reader, maxFlags int, currentFlag *int) string {
			return collectFlagInputWithHist(cmd, f, flagRequired, defValue, reader, histProvider, maxFlags, currentFlag)
		}
		collectRepeatedFlagInputFunc = func(cmd *cobra.Command, f *pflag.Flag, defValue string, reader *bufio.Reader, maxFlags int, currentFlag *int) string {
			return collectRepeatedFlagInputWithHist(cmd, f, defValue, reader, histProvider, maxFlags, currentFlag)
		}
	}
	flagCount := getFlagCount(cmds...)
	currentFlag := 1
	for _, cmd := range cmds {
		// Iterate over all flags of the command
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			// Ask interactively for input
			if f.Name == "help" {
				return
			}

			if !showedChain {
				printInfo(fmt.Sprintf("\n`%s` will be called.\n\nIn the following steps the possible flags will be collected. Continue with ⏎\n", cmdChain))
				//fmt.Printf("\n`%s` will be called.\n\nIn the following steps the possible flags will be collected. Continue with ⏎\n", cmdChain)
				reader.ReadString('\n') // read entire line
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
				configuredFlags += collectRepeatedFlagInputFunc(cmd, f, defValue, reader, flagCount, &currentFlag)
			} else {
				configuredFlags += collectFlagInputFunc(cmd, f, flagRequired, defValue, reader, flagCount, &currentFlag)
			}
		})
	}
	return configuredFlags
}

func getFlagCount(cmds ...*cobra.Command) int {
	var flagCount int
	for _, c := range cmds {
		c.Flags().VisitAll(func(f *pflag.Flag) {
			// Ask interactively for input
			if f.Name != "help" {
				flagCount++
			}
		})
	}
	return flagCount
}

const HELP = "?"
const HELP2 = "help"
const HELP3 = "--help"

// implements the user interaction to get the required input for a flag
func collectFlagInput(cmd *cobra.Command, f *pflag.Flag, flagRequired bool, defValue string, reader *bufio.Reader, maxFlags int, currentFlag *int) string {
	var setValue string
	for {
		fmt.Printf("\n[%d/%d] --%s: %s: ", *currentFlag, maxFlags, f.Name, defValue)
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
					fmt.Printf("\nSet value: --%s %s\n", f.Name, input)
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
	*currentFlag++
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

func printInfo(msg string) {
	c := color.New(color.FgHiBlue)
	c.Print(msg)
}

func collectRepeatedFlagInput(cmd *cobra.Command, f *pflag.Flag, defValue string, reader *bufio.Reader, maxFlags int, currentFlag *int) string {
	var setValue string
	bFirst := true
	for {
		if bFirst {
			fmt.Printf("\n[%d/%d] --%s %s\nmultiple values possible: ", *currentFlag, maxFlags, f.Name, defValue)
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
					fmt.Printf("\nSet value: --%s %s\n", f.Name, input)
					setValue += flagTxt(f, input)
				}
			}
		} else {
			break
		}
	}
	*currentFlag++
	return setValue
}

func getHistInput(f *pflag.Flag, hasHist bool, reader *bufio.Reader, histProvider HistoryProvider, defValue, histHint string, txtToIgnore []string, maxFlags, currentFlag int) (string, bool) {
	if hasHist {
		if input, err := histProvider.InputFromHist(f.Name, fmt.Sprintf("\n'--%s' %s (%s), to enter new value press ESC", f.Name, defValue, f.Usage), txtToIgnore, maxFlags, currentFlag); err == nil {
			return trimInput(input), true
		}
	}
	printInfo(histHint)
	input, _ := reader.ReadString('\n') // read entire line
	input = trimInput(input)
	return input, false
}

func collectFlagInputWithHist(cmd *cobra.Command, f *pflag.Flag, flagRequired bool, defValue string, reader *bufio.Reader, histProvider HistoryProvider, maxFlags int, currentFlag *int) string {
	var setValue string
	for {
		hasHist, _ := getHistHint(f.Name, histProvider)
		histHint := fmt.Sprintf("\n[%d/%d] new value for: --%s %s: ", *currentFlag, maxFlags, f.Name, defValue)
		input, _ := getHistInput(f, hasHist, reader, histProvider, defValue, histHint, []string{}, maxFlags, *currentFlag)

		if input != "" {
			if input == HELP || input == HELP2 || input == HELP3 {
				fmt.Printf("  Expected input: %s\n", f.Usage)
			} else {
				// User provided a value -> set it
				if err := cmd.Flags().Set(f.Name, input); err != nil {
					fmt.Printf("⚠️  Could not set flag %s: %v\nContinue with ⏎\n", f.Name, err)
					reader.ReadString('\n') // read entire line
				} else {
					fmt.Printf("\nSet value: --%s %s\n", f.Name, input)
					setValue = input
					histProvider.SaveHist(f.Name, setValue)
					break
				}
			}
		} else if flagRequired {
			fmt.Printf("⚠️  Flag %s is required, so input is needed! Continue with ⏎\n", f.Name)
			reader.ReadString('\n') // read entire line
		} else {
			break
		}
	}
	*currentFlag++
	if setValue != "" {
		return flagTxt(f, setValue)
	} else {
		return setValue
	}
}

func collectRepeatedFlagInputWithHist(cmd *cobra.Command, f *pflag.Flag, defValue string, reader *bufio.Reader, histProvider HistoryProvider, maxFlags int, currentFlag *int) string {
	var setValue string
	hasHist, _ := getHistHint(f.Name, histProvider)
	histHint := fmt.Sprintf("\n[%d/%d] --%s %s\nmultiple values possible, leave empty to skip or finish: ", *currentFlag, maxFlags, f.Name, defValue)
	txtToIgnore := make([]string, 0)
	for {
		input, fromHist := getHistInput(f, hasHist, reader, histProvider, defValue, histHint, txtToIgnore, maxFlags, *currentFlag)
		if !fromHist {
			hasHist = false
		}

		if input != "" {
			if input == HELP || input == HELP2 || input == HELP3 {
				fmt.Printf("  Expected input: %s, continue with ⏎\n", f.Usage)
				reader.ReadString('\n') // read entire line
			} else {
				// User provided a value -> set it
				if err := cmd.Flags().Set(f.Name, input); err != nil {
					fmt.Printf("⚠️  Could not set flag %s: %v, continue with ⏎\n", f.Name, err)
					reader.ReadString('\n') // read entire line
				} else {
					fmt.Printf("\nSet value: --%s %s\n", f.Name, input)
					histProvider.SaveHist(f.Name, input)
					txtToIgnore = append(txtToIgnore, input)
					setValue += flagTxt(f, input)
				}
			}
		} else {
			break
		}
	}
	*currentFlag++
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

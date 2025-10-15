package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func PromptYesOrNo(question string, defaultToYes bool) (bool, error) {
	r := bufio.NewReader(os.Stdin)
	if defaultToYes {
		fmt.Fprintf(os.Stderr, "%s [Y/n]: ", question)
	} else {
		fmt.Fprintf(os.Stderr, "%s [y/N]: ", question)
	}
	for {
		answer, err := r.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			return defaultToYes, err
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "":
			return defaultToYes, nil
		}
		fmt.Fprintf(os.Stderr, "Please answer with 'y' or 'n': ")
	}
}

func PromptString(question string) (string, error) {
	r := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "%s: ", question)

	answer, err := r.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		return "", err
	}
	return strings.TrimSpace(answer), nil
}

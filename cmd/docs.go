package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
)

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation",
	Hidden: true,
}

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate man pages",
	Long: `Generate man pages for all Coolify CLI commands.

The man pages will be written to the specified directory (default: ./man).`,
	Example: `  coolify docs man
  coolify docs man --output-dir=/usr/local/share/man/man1`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputDir, _ := cmd.Flags().GetString("output-dir")

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate man pages
		header := &doc.GenManHeader{
			Title:   "COOLIFY",
			Section: "1",
			Source:  "Coolify CLI",
		}

		if err := doc.GenManTree(rootCmd, header, outputDir); err != nil {
			return fmt.Errorf("failed to generate man pages: %w", err)
		}

		absPath, _ := filepath.Abs(outputDir)
		fmt.Printf("Man pages generated successfully in: %s\n", absPath)
		fmt.Println("\nTo install the man pages system-wide:")
		fmt.Println("  sudo cp man/*.1 /usr/local/share/man/man1/")
		fmt.Println("  sudo mandb")
		fmt.Println("\nTo view a man page:")
		fmt.Println("  man coolify")
		fmt.Println("  man coolify-servers")

		return nil
	},
}

var markdownCmd = &cobra.Command{
	Use:     "markdown",
	Aliases: []string{"md"},

	Short: "Generate markdown documentation",
	Long: `Generate markdown documentation for all Coolify CLI commands.

The markdown files will be written to the specified directory (default: ./docs).`,
	Example: `  coolify docs markdown
  coolify docs markdown --output-dir=./documentation`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputDir, _ := cmd.Flags().GetString("output-dir")

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate markdown docs
		if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
			return fmt.Errorf("failed to generate markdown docs: %w", err)
		}

		absPath, _ := filepath.Abs(outputDir)
		fmt.Printf("Markdown documentation generated successfully in: %s\n", absPath)

		return nil
	},
}

var llmsCmd = &cobra.Command{
	Use:   "llms",
	Short: "Generate llms.txt for AI agent command specification",
	Long: `Generate a machine-readable llms.txt file that defines all CLI commands and their parameters.

This file is intended to enable AI agents to understand and interact with the CLI.
The output file will be written to the specified path (default: ./llms.txt).`,
	Example: `  coolify docs llms
  coolify docs llms --output=./llms.txt`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputFile, _ := cmd.Flags().GetString("output")

		var sb strings.Builder
		writeLLMsCommand(&sb, rootCmd, "coolify")

		if err := os.WriteFile(outputFile, []byte(sb.String()), 0600); err != nil {
			return fmt.Errorf("failed to write llms.txt: %w", err)
		}

		absPath, _ := filepath.Abs(outputFile)
		fmt.Printf("llms.txt generated successfully: %s\n", absPath)

		return nil
	},
}

// writeLLMsCommand recursively writes command documentation in llms.txt format.
func writeLLMsCommand(sb *strings.Builder, cmd *cobra.Command, parentPath string) {
	// Build the full command path including args from Use field
	commandPath := parentPath
	if cmd.HasParent() {
		parts := strings.Fields(cmd.Use)
		commandPath = parentPath + " " + parts[0]
		// Append positional args from the Use field (e.g., "<uuid>", "[optional]")
		if len(parts) > 1 {
			commandPath += " " + strings.Join(parts[1:], " ")
		}
	}

	// Skip the docs command itself and help command
	if cmd.Name() == "docs" || cmd.Name() == "help" {
		return
	}

	// Determine if this command should be written
	isRoot := !cmd.HasParent()
	isRunnable := cmd.RunE != nil || cmd.Run != nil
	hasVisibleChildren := false
	for _, child := range cmd.Commands() {
		if !child.Hidden && child.Name() != "help" {
			hasVisibleChildren = true
			break
		}
	}

	// Write the root command, runnable commands, and leaf commands (no children)
	if isRoot || isRunnable || !hasVisibleChildren {
		// Get description - prefer Long if it's a single clean sentence, otherwise use Short
		description := cmd.Short
		if cmd.Long != "" {
			longLines := strings.Split(strings.TrimSpace(cmd.Long), "\n")
			if len(longLines) == 1 && len(longLines[0]) < 200 {
				description = longLines[0]
			}
		}

		fmt.Fprintf(sb, "Command: %s\n", commandPath)
		fmt.Fprintf(sb, "Description: %s\n", description)

		// For root command, show persistent flags; for others, show local flags
		var flags []*pflag.Flag
		if isRoot {
			cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "help" {
					return
				}
				flags = append(flags, f)
			})
		} else {
			cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "help" {
					return
				}
				flags = append(flags, f)
			})
		}

		if len(flags) == 0 {
			sb.WriteString("Parameters: (None)\n")
		} else {
			sb.WriteString("Parameters:\n")
			for _, f := range flags {
				flagType := f.Value.Type()
				// Normalize type names
				switch flagType {
				case "int", "int32", "int64":
					flagType = "integer"
				case "bool":
					flagType = "boolean"
				}

				// Check if the flag is marked as required via cobra annotation
				// or via "(required)" in the usage string
				required := isFlagRequired(f)

				fmt.Fprintf(sb, "  - name: --%s\n", f.Name)
				fmt.Fprintf(sb, "    type: %s\n", flagType)
				fmt.Fprintf(sb, "    description: %s\n", f.Usage)
				fmt.Fprintf(sb, "    required: %t\n", required)
			}
		}

		sb.WriteString("\n")
	}

	// Recurse into subcommands
	for _, child := range cmd.Commands() {
		if child.Hidden || child.Name() == "help" {
			continue
		}
		childPath := parentPath
		if cmd.HasParent() {
			parts := strings.Fields(cmd.Use)
			childPath = parentPath + " " + parts[0]
		}
		writeLLMsCommand(sb, child, childPath)
	}
}

// isFlagRequired checks if a flag is required by looking at cobra annotations
// and the "(required)" convention in usage strings.
func isFlagRequired(f *pflag.Flag) bool {
	// Check cobra's MarkFlagRequired annotation
	if ann, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(ann) > 0 && ann[0] == "true" {
		return true
	}
	// Check for "(required)" in usage string (convention used in this codebase)
	return strings.Contains(strings.ToLower(f.Usage), "(required)")
}

func NewDocsCommand() *cobra.Command {
	docsCmd.AddCommand(manCmd)
	docsCmd.AddCommand(markdownCmd)
	docsCmd.AddCommand(llmsCmd)

	manCmd.Flags().StringP("output-dir", "o", "./man", "Output directory for man pages")
	markdownCmd.Flags().StringP("output-dir", "o", "./docs", "Output directory for markdown files")
	llmsCmd.Flags().StringP("output", "o", "./llms.txt", "Output file path")

	return docsCmd
}

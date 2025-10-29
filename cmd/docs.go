package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
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

func NewDocsCommand() *cobra.Command {
	docsCmd.AddCommand(manCmd)
	docsCmd.AddCommand(markdownCmd)

	manCmd.Flags().StringP("output-dir", "o", "./man", "Output directory for man pages")
	markdownCmd.Flags().StringP("output-dir", "o", "./docs", "Output directory for markdown files")

	return docsCmd
}

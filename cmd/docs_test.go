package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestWriteLLMsCommandIncludesShorthandAndDefaults(t *testing.T) {
	root := &cobra.Command{Use: "coolify"}
	child := &cobra.Command{
		Use:   "logs <uuid>",
		Short: "Show logs",
		Run:   func(_ *cobra.Command, _ []string) {},
	}
	child.Flags().IntP("lines", "n", 0, "Number of log lines to display (0 = all)")
	child.Flags().Bool("verbose", false, "Verbose output")
	child.Flags().Bool("enabled", true, "Enabled by default")
	root.AddCommand(child)

	var sb strings.Builder
	writeLLMsCommand(&sb, child, "coolify")
	got := sb.String()

	for _, want := range []string{
		"Command: coolify logs <uuid>",
		"  - name: --lines (-n)",
		"    default: 0",
		"  - name: --verbose",
		"    default: false",
		"  - name: --enabled",
		"    default: true",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected output to contain %q\nfull output:\n%s", want, got)
		}
	}
}

func TestWriteLLMsAliasesUsesCommandTree(t *testing.T) {
	root := &cobra.Command{Use: "coolify"}
	teams := &cobra.Command{Use: "teams", Aliases: []string{"team"}}
	members := &cobra.Command{Use: "members", Aliases: []string{"member"}}
	start := &cobra.Command{
		Use:     "start <uuid>",
		Aliases: []string{"deploy"},
	}

	root.AddCommand(teams)
	root.AddCommand(start)
	teams.AddCommand(members)

	var sb strings.Builder
	writeLLMsAliases(&sb, root, "coolify")
	got := sb.String()

	for _, want := range []string{
		"## Command Aliases",
		"`coolify start` | `coolify deploy`",
		"`coolify teams` | `coolify team`",
		"`coolify teams members` | `coolify teams member`",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected alias output to contain %q\nfull output:\n%s", want, got)
		}
	}
}

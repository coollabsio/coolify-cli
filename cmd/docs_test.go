package cmd

import (
	"os"
	"path/filepath"
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

func TestBuildQuickLLMSTextIncludesCoreGuidance(t *testing.T) {
	got := buildQuickLLMSText("./llms-full.txt")

	for _, want := range []string{
		"# Coolify CLI - llms.txt",
		"Prefer `--format json` for automation and parsing.",
		"coolify context verify",
		"coolify app logs <uuid> --follow",
		"Full command and parameter catalog: ./llms-full.txt",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected quick llms output to contain %q\nfull output:\n%s", want, got)
		}
	}
}

func TestWriteLLMsArtifactsWritesQuickAndFullFiles(t *testing.T) {
	tempDir := t.TempDir()
	quickPath := filepath.Join(tempDir, "llms.txt")
	fullPath := filepath.Join(tempDir, "nested", "llms-full.txt")

	if err := writeLLMsArtifacts(quickPath, fullPath); err != nil {
		t.Fatalf("writeLLMsArtifacts() error = %v", err)
	}

	quickContent, err := os.ReadFile(quickPath)
	if err != nil {
		t.Fatalf("failed reading quick file: %v", err)
	}
	fullContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed reading full file: %v", err)
	}

	for _, want := range []struct {
		content string
		substr  string
	}{
		{string(quickContent), "./nested/llms-full.txt"},
		{string(fullContent), "../llms.txt"},
		{string(fullContent), "## Command Reference"},
	} {
		if !strings.Contains(want.content, want.substr) {
			t.Fatalf("expected generated content to contain %q", want.substr)
		}
	}
}

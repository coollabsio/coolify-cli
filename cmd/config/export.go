package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

func newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <environment_uuid>",
		Short: "Export an environment as a coolify.json file",
		Long: `Export an environment as a coolify.json file. The export has no secret values.
Needs a token with the read:sensitive ability.

Examples:
  coolify config export <environment_uuid>
  coolify config export <environment_uuid> -o coolify.json`,
		Args: cli.ExactArgs(1, "<environment_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := configService(cmd)
			if err != nil {
				return err
			}
			file, err := svc.Export(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			var indented bytes.Buffer
			if err := json.Indent(&indented, unescapeSlashes(file), "", "    "); err != nil {
				return fmt.Errorf("failed to format the export: %w", err)
			}
			indented.WriteByte('\n')

			path, _ := cmd.Flags().GetString("output")
			if path == "" {
				_, err := cmd.OutOrStdout().Write(indented.Bytes())
				return err
			}
			if err := os.WriteFile(path, indented.Bytes(), 0o600); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Exported to %s.\n", path)
			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "", "Write the file to this path instead of standard output")
	return cmd
}

// unescapeSlashes turns the \/ escapes of PHP's JSON encoder into /, so the
// file reads like hand-written JSON. Other escapes stay as they are.
func unescapeSlashes(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == '\\' && i+1 < len(data) {
			if data[i+1] != '/' {
				out = append(out, data[i])
			}
			out = append(out, data[i+1])
			i++
			continue
		}
		out = append(out, data[i])
	}
	return out
}

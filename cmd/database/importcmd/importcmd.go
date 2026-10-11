// Package importcmd implements the `import` command group shared by
// `coolify database import` (standalone databases) and
// `coolify service database import` (databases inside a service).
package importcmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// targetSpec describes how the leading positional arguments map to a target.
type targetSpec struct {
	usage   string // e.g. "<database_uuid>"
	nargs   int
	command string // command path used in examples
	build   func(args []string) service.DatabaseImportTarget
}

// NewStandaloneCommand returns `import` for standalone databases.
func NewStandaloneCommand() *cobra.Command {
	return newCommand(targetSpec{
		usage:   "<database_uuid>",
		nargs:   1,
		command: "coolify database import",
		build: func(args []string) service.DatabaseImportTarget {
			return service.StandaloneDatabaseImportTarget(args[0])
		},
	})
}

// NewServiceDatabaseCommand returns `import` for databases inside a service.
func NewServiceDatabaseCommand() *cobra.Command {
	return newCommand(targetSpec{
		usage:   "<service_uuid> <database_uuid>",
		nargs:   2,
		command: "coolify service database import",
		build: func(args []string) service.DatabaseImportTarget {
			return service.ServiceDatabaseImportTarget(args[0], args[1])
		},
	})
}

func newCommand(spec targetSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "import",
		Aliases: []string{"imports"},
		Short:   "Import (restore) a backup into a database",
		Long: `Import a backup into a database from a local file, an existing upload, an S3 storage or a path on the database's server.

Examples:
  ` + spec.command + ` start ` + spec.usage + ` --file ./dump.sql.gz --wait
  ` + spec.command + ` start ` + spec.usage + ` --s3-storage-uuid <s3_uuid> --s3-path /backups/db.dump
  ` + spec.command + ` start ` + spec.usage + ` --server-path /var/backups/db.sql.gz
  ` + spec.command + ` upload ` + spec.usage + ` ./dump.sql.gz
  ` + spec.command + ` status ` + spec.usage + ` <activity_id> --wait`,
	}
	cmd.AddCommand(newUploadCommand(spec))
	cmd.AddCommand(newStartCommand(spec))
	cmd.AddCommand(newStatusCommand(spec))
	return cmd
}

func newUploadCommand(spec targetSpec) *cobra.Command {
	usage := spec.usage + " <file>"
	return &cobra.Command{
		Use:   "upload " + usage,
		Short: "Upload a local backup file for a later import",
		Long:  `Upload a local backup file (max 10 GiB) in chunks. Prints the upload ID to pass to 'start --upload-id'.`,
		Args:  cli.ExactArgs(spec.nargs+1, usage),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			uploaded, err := uploadFile(cmd, service.NewDatabaseImportService(client), spec.build(args), args[spec.nargs])
			if err != nil {
				return err
			}
			return format(cmd, uploaded)
		},
	}
}

func newStartCommand(spec targetSpec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start " + spec.usage,
		Short: "Start an import into the database",
		Long: `Start an import into the database. Exactly one source is required:
  --file <path>                          upload a local backup file, then import it
  --upload-id <id>                       import a file uploaded with 'upload'
  --s3-storage-uuid <uuid> --s3-path <p> import a backup from S3 storage
  --server-path <path>                   import a file on the database's server (token needs deploy permission)

The import runs in the background; use --wait to follow it until it finishes.`,
		Args: cli.ExactArgs(spec.nargs, spec.usage),
		RunE: func(cmd *cobra.Command, args []string) error {
			req, localFile, err := buildStartRequest(cmd)
			if err != nil {
				return err
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			svc := service.NewDatabaseImportService(client)
			target := spec.build(args)

			if localFile != "" {
				uploaded, err := uploadFile(cmd, svc, target, localFile)
				if err != nil {
					return err
				}
				req.UploadID = &uploaded.UploadID
			}

			started, err := svc.Start(cmd.Context(), target, req)
			if err != nil {
				return err
			}

			if wait, _ := cmd.Flags().GetBool("wait"); wait {
				fmt.Fprintf(cmd.ErrOrStderr(), "Import %d queued, waiting for it to finish...\n", started.ID)
				return waitAndReport(cmd, svc, target, started.ID)
			}

			if err := format(cmd, started); err != nil {
				return err
			}
			if isTableFormat(cmd) {
				fmt.Fprintf(cmd.OutOrStdout(), "\nCheck progress with: %s status %s %d --wait\n", spec.command, strings.Join(args, " "), started.ID)
			}
			return nil
		},
	}
	cmd.Flags().String("file", "", "Local backup file to upload and import")
	cmd.Flags().String("upload-id", "", "ID of a previously uploaded backup file")
	cmd.Flags().String("s3-storage-uuid", "", "S3 storage UUID to import from (requires --s3-path)")
	cmd.Flags().String("s3-path", "", "Path of the backup in the S3 storage")
	cmd.Flags().String("server-path", "", "Absolute path of the backup on the database's server")
	cmd.Flags().Bool("dump-all", false, "The backup contains all databases")
	cmd.Flags().Bool("replace-existing", false, "PostgreSQL: drop matching objects before restoring a single-database archive")
	cmd.Flags().Bool("keep-owners", false, "PostgreSQL: restore object owners and privileges from a single-database archive")
	cmd.Flags().Bool("restore-mysql-users", false, "MySQL/MariaDB all-databases backups: also restore users, passwords and privileges")
	cmd.Flags().String("sqlite-database", "", "SQLite: database file to restore into (e.g. app.db)")
	addWaitFlags(cmd)
	return cmd
}

func newStatusCommand(spec targetSpec) *cobra.Command {
	usage := spec.usage + " <activity_id>"
	cmd := &cobra.Command{
		Use:   "status " + usage,
		Short: "Show the status of an import",
		Args:  cli.ExactArgs(spec.nargs+1, usage),
		RunE: func(cmd *cobra.Command, args []string) error {
			activityID, err := strconv.Atoi(args[spec.nargs])
			if err != nil || activityID <= 0 {
				return fmt.Errorf("invalid activity ID %q: must be a positive integer", args[spec.nargs])
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			svc := service.NewDatabaseImportService(client)
			target := spec.build(args)

			if wait, _ := cmd.Flags().GetBool("wait"); wait {
				return waitAndReport(cmd, svc, target, activityID)
			}
			status, err := svc.Status(cmd.Context(), target, activityID)
			if err != nil {
				return err
			}
			return formatStatus(cmd, status)
		},
	}
	addWaitFlags(cmd)
	return cmd
}

func addWaitFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("wait", false, "Wait until the import finishes; exits non-zero if it fails")
	cmd.Flags().Duration("poll-interval", service.DefaultImportPollInterval, "Status polling interval when --wait is set")
}

// buildStartRequest validates the source flags and builds the request. When
// --file is set, localFile is returned and the caller fills UploadID after
// uploading it.
func buildStartRequest(cmd *cobra.Command) (req *models.DatabaseImportRequest, localFile string, err error) {
	flags := cmd.Flags()
	file, _ := flags.GetString("file")
	uploadID, _ := flags.GetString("upload-id")
	s3UUID, _ := flags.GetString("s3-storage-uuid")
	s3Path, _ := flags.GetString("s3-path")
	serverPath, _ := flags.GetString("server-path")

	sources := 0
	for _, set := range []bool{file != "", uploadID != "", s3UUID != "" || s3Path != "", serverPath != ""} {
		if set {
			sources++
		}
	}
	if sources != 1 {
		return nil, "", errors.New("exactly one source is required: --file, --upload-id, --s3-storage-uuid with --s3-path, or --server-path")
	}

	req = &models.DatabaseImportRequest{}
	switch {
	case file != "":
		req.Source = models.DatabaseImportSourceUpload
		localFile = file
	case uploadID != "":
		req.Source = models.DatabaseImportSourceUpload
		req.UploadID = &uploadID
	case s3UUID != "" || s3Path != "":
		if s3UUID == "" || s3Path == "" {
			return nil, "", errors.New("--s3-storage-uuid and --s3-path must be used together")
		}
		req.Source = models.DatabaseImportSourceS3
		req.S3StorageUUID = &s3UUID
		req.Path = &s3Path
	default:
		req.Source = models.DatabaseImportSourceServer
		req.Path = &serverPath
	}

	for flag, field := range map[string]**bool{
		"dump-all":            &req.DumpAll,
		"replace-existing":    &req.ReplaceExisting,
		"keep-owners":         &req.KeepOwners,
		"restore-mysql-users": &req.RestoreMysqlUsers,
	} {
		if flags.Changed(flag) {
			value, _ := flags.GetBool(flag)
			*field = &value
		}
	}
	if flags.Changed("sqlite-database") {
		sqliteDatabase, _ := flags.GetString("sqlite-database")
		req.SqliteDatabase = &sqliteDatabase
	}
	return req, localFile, nil
}

func uploadFile(cmd *cobra.Command, svc *service.DatabaseImportService, target service.DatabaseImportTarget, path string) (*models.DatabaseImportUpload, error) {
	stderr := cmd.ErrOrStderr()
	showProgress := isTerminal(stderr)
	name := filepath.Base(path)
	uploaded, err := svc.Upload(cmd.Context(), target, path, func(sent, total int64) {
		if showProgress {
			fmt.Fprintf(stderr, "\rUploading %s: %3d%% (%s / %s)", name, sent*100/total, humanBytes(sent), humanBytes(total))
		}
	})
	if showProgress {
		fmt.Fprintln(stderr)
	}
	return uploaded, err
}

func waitAndReport(cmd *cobra.Command, svc *service.DatabaseImportService, target service.DatabaseImportTarget, activityID int) error {
	interval, _ := cmd.Flags().GetDuration("poll-interval")
	status, err := svc.Wait(cmd.Context(), target, activityID, interval)
	if err != nil {
		return err
	}
	if err := formatStatus(cmd, status); err != nil {
		return err
	}
	if !status.Succeeded() {
		return fmt.Errorf("import %d ended with status %q", status.ID, status.Status)
	}
	return nil
}

// formatStatus prints the status; in table format the (multi-line) command
// output is printed below the table instead of inside it.
func formatStatus(cmd *cobra.Command, status *models.DatabaseImportStatus) error {
	if err := format(cmd, status); err != nil {
		return err
	}
	if isTableFormat(cmd) && status.Output != nil && strings.TrimSpace(*status.Output) != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "\nOutput:\n%s\n", strings.TrimRight(*status.Output, "\n"))
	}
	return nil
}

func format(cmd *cobra.Command, data any) error {
	formatName, _ := cmd.Flags().GetString("format")
	if formatName == "" {
		formatName = output.FormatTable
	}
	formatter, err := output.NewFormatter(formatName, output.Options{Writer: cmd.OutOrStdout()})
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}
	return formatter.Format(data)
}

func isTableFormat(cmd *cobra.Command) bool {
	formatName, _ := cmd.Flags().GetString("format")
	return formatName == "" || formatName == output.FormatTable
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

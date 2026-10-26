package server

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// maxRegistryPasswordBytes matches the API limit for the password field.
const maxRegistryPasswordBytes = 20000

// terminalFD returns the file descriptor of in when it is an interactive terminal.
// It is a variable so tests can simulate a TTY.
var terminalFD = func(in io.Reader) (int, bool) {
	file, ok := in.(*os.File)
	if !ok {
		return 0, false
	}
	fd := int(file.Fd())
	return fd, term.IsTerminal(fd)
}

// readHiddenPassword reads a password from a terminal without echo.
// It is a variable so tests can simulate a TTY.
var readHiddenPassword = term.ReadPassword

func newRegistrySvc(cmd *cobra.Command) (*service.ServerRegistryService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewServerRegistryService(client), nil
}

// NewRegistryCommand creates the server registry parent command.
func NewRegistryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "registry",
		Aliases: []string{"registries"},
		Short:   "Manage Docker registry logins on a server",
		Long: `Manage the Docker registry logins of a server.

Logins are stored in the Docker config of the server that deployments use,
so applications, databases and services on the server can pull private images
and build servers can push them.`,
	}

	cmd.AddCommand(newRegistryListCommand())
	cmd.AddCommand(newRegistryLoginCommand())
	cmd.AddCommand(newRegistryCheckCommand())
	cmd.AddCommand(newRegistryLogoutCommand())

	return cmd
}

func newRegistryListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <server_uuid>",
		Short: "List registry logins of a server",
		Long: `List the Docker registries a server is logged in to and the registries its resources pull from or push to.

Credentials are never returned. Usernames are hidden in table output unless -s is given.
Requires an API token with the read:sensitive ability.`,
		Args: cli.ExactArgs(1, "<server_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newRegistrySvc(cmd)
			if err != nil {
				return err
			}

			result, err := svc.List(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("failed to list registry logins: %w", err)
			}

			if result.Error != nil && *result.Error != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", *result.Error)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			formatter, err := output.NewFormatter(format, output.Options{
				Writer:        cmd.OutOrStdout(),
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			if format != output.FormatTable {
				return formatter.Format(result)
			}

			rows := service.ServerRegistryRows(result.Registries, showSensitive, cli.SensitiveInformationOverlay)
			if err := formatter.Format(rows); err != nil {
				return err
			}

			if !showSensitive && hasMaskedUsername(rows) {
				fmt.Fprintln(cmd.OutOrStdout(), "\nNote: Use -s to show usernames.")
			}
			return nil
		},
	}
}

func hasMaskedUsername(rows []models.ServerRegistryRow) bool {
	for _, row := range rows {
		if row.Username == cli.SensitiveInformationOverlay {
			return true
		}
	}
	return false
}

func newRegistryLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login <server_uuid> [<server_uuid>...]",
		Short: "Log in to a Docker registry on one or more servers",
		Long: `Run docker login on the server, writing to the Docker config that deployments use.
Logging in again to the same registry replaces the saved login.

The password or access token is never accepted as a flag value. Pipe it with
--password-stdin, or enter it at the hidden prompt when running in a terminal.`,
		Example: `  echo "$GHCR_TOKEN" | coolify server registry login <server_uuid> --registry ghcr.io --username octocat --password-stdin
  coolify server registry login <server_uuid> --registry registry.example.com:5000 --username deploy
  cat token.txt | coolify server registry login <uuid_1> <uuid_2> --registry docker.io --username me --password-stdin`,
		Args: cli.MinArgs(1, "<server_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			registry, _ := cmd.Flags().GetString("registry")
			username, _ := cmd.Flags().GetString("username")
			passwordStdin, _ := cmd.Flags().GetBool("password-stdin")

			registry = service.NormalizeRegistry(registry)
			if registry == "" {
				return errors.New("--registry is required, for example ghcr.io or registry.example.com:5000")
			}
			username = strings.TrimSpace(username)
			if username == "" {
				return errors.New("--username is required")
			}

			password, err := readRegistryPassword(cmd, passwordStdin)
			if err != nil {
				return err
			}

			svc, err := newRegistrySvc(cmd)
			if err != nil {
				return err
			}

			req := models.ServerRegistryLoginRequest{Registry: registry, Username: username, Password: password}

			if len(args) == 1 {
				resp, err := svc.Login(cmd.Context(), args[0], req)
				if err != nil {
					return fmt.Errorf("failed to log in to %s: %w", registry, err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), loginMessage(resp, registry))
				return nil
			}

			failed := 0
			for _, serverUUID := range args {
				resp, err := svc.Login(cmd.Context(), serverUUID, req)
				if err != nil {
					failed++
					fmt.Fprintf(cmd.ErrOrStderr(), "%s: failed to log in to %s: %v\n", serverUUID, registry, err)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", serverUUID, loginMessage(resp, registry))
			}
			if failed > 0 {
				return fmt.Errorf("login to %s failed on %d of %d servers", registry, failed, len(args))
			}
			return nil
		},
	}

	cmd.Flags().String("registry", "", "Registry host, optionally with a port (e.g. ghcr.io, registry.example.com:5000, docker.io for Docker Hub) (required)")
	cmd.Flags().StringP("username", "u", "", "Registry username (required)")
	cmd.Flags().Bool("password-stdin", false, "Read the password or access token from stdin")

	return cmd
}

func loginMessage(resp *models.Response, registry string) string {
	if resp != nil && resp.Message != "" {
		return resp.Message
	}
	return fmt.Sprintf("Logged in to %s.", registry)
}

// readRegistryPassword reads the registry password from stdin when fromStdin is set,
// otherwise from a hidden prompt when stdin is a terminal.
func readRegistryPassword(cmd *cobra.Command, fromStdin bool) (string, error) {
	in := cmd.InOrStdin()

	var password string
	if fromStdin {
		data, err := io.ReadAll(io.LimitReader(in, maxRegistryPasswordBytes+2))
		if err != nil {
			return "", fmt.Errorf("failed to read password from stdin: %w", err)
		}
		// Like docker login --password-stdin: drop one trailing newline only.
		password = strings.TrimSuffix(string(data), "\n")
		password = strings.TrimSuffix(password, "\r")
	} else {
		fd, ok := terminalFD(in)
		if !ok {
			return "", errors.New("no password given: pipe the password or access token with --password-stdin (the password cannot be passed as a flag)")
		}
		fmt.Fprint(cmd.ErrOrStderr(), "Password: ")
		data, err := readHiddenPassword(fd)
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		password = string(data)
	}

	if password == "" {
		return "", errors.New("password must not be empty")
	}
	if len(password) > maxRegistryPasswordBytes {
		return "", fmt.Errorf("password must not be longer than %d characters", maxRegistryPasswordBytes)
	}
	return password, nil
}

func newRegistryCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check <server_uuid> <registry>",
		Short: "Check that the saved registry login of a server still works",
		Args:  cli.ExactArgs(2, "<server_uuid> <registry>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newRegistrySvc(cmd)
			if err != nil {
				return err
			}

			registry := service.NormalizeRegistry(args[1])
			resp, err := svc.Check(cmd.Context(), args[0], registry)
			if err != nil {
				return fmt.Errorf("failed to check login for %s: %w", registry, err)
			}

			message := resp.Message
			if message == "" {
				message = fmt.Sprintf("The login for %s works.", registry)
			}
			fmt.Fprintln(cmd.OutOrStdout(), message)
			return nil
		},
	}
}

func newRegistryLogoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout <server_uuid> <registry>",
		Short: "Log out from a Docker registry on a server",
		Long: `Run docker logout on the server for a registry.
Resources on the server that pull private images from this registry will fail to deploy until you log in again.`,
		Args: cli.ExactArgs(2, "<server_uuid> <registry>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverUUID := args[0]
			registry := service.NormalizeRegistry(args[1])

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				var response string
				fmt.Fprintf(cmd.OutOrStdout(), "Are you sure you want to log out from %s on server %s? (yes/no): ", registry, serverUUID)
				if _, err := fmt.Fscanln(cmd.InOrStdin(), &response); err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if response != "yes" && response != "y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Logout cancelled.")
					return nil
				}
			}

			svc, err := newRegistrySvc(cmd)
			if err != nil {
				return err
			}

			if err := svc.Logout(cmd.Context(), serverUUID, registry); err != nil {
				return fmt.Errorf("failed to log out from %s: %w", registry, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Logged out from %s on server %s.\n", registry, serverUUID)
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	return cmd
}

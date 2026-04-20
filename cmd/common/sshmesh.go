// Package common hosts flag sets and helpers shared between multiple
// top-level commands that SSH into a list of servers (init, firewall, ...).
package common

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	internalssh "github.com/coollabsio/coolify-cli/internal/ssh"
)

// SSHMeshFlags holds the flags shared by every command that fans out over
// a list of SSH-reachable servers (coolify init, coolify firewall, ...).
type SSHMeshFlags struct {
	Servers             []string
	SSHKey              string
	SSHUser             string
	SSHPort             int
	SSHPassphrasePrompt bool
	Concurrency         int
	SSHTimeout          string
}

// BindSSHMeshFlags registers the shared flags as PersistentFlags on cmd.
func BindSSHMeshFlags(cmd *cobra.Command, f *SSHMeshFlags) {
	pf := cmd.PersistentFlags()

	pf.StringSliceVar(&f.Servers, "servers", nil,
		"Comma-separated server IPs (required)")
	pf.StringVar(&f.SSHKey, "ssh-key", "",
		"Path to SSH private key used to connect to servers (required)")
	pf.StringVar(&f.SSHUser, "ssh-user", "root",
		"SSH username")
	pf.IntVar(&f.SSHPort, "ssh-port", 22,
		"SSH port")
	pf.BoolVar(&f.SSHPassphrasePrompt, "ssh-passphrase-prompt", false,
		"Prompt for SSH key passphrase (also reads COOLIFY_SSH_PASSPHRASE env var)")
	pf.IntVar(&f.Concurrency, "concurrency", 10,
		"Maximum number of parallel SSH connections")
	pf.StringVar(&f.SSHTimeout, "ssh-timeout", "30s",
		"SSH connection timeout (e.g. 30s, 1m)")
}

// ParseSSHTimeout parses SSHTimeout, falling back to 30s on error/zero.
func (f *SSHMeshFlags) ParseSSHTimeout() time.Duration {
	d, err := time.ParseDuration(f.SSHTimeout)
	if err != nil || d <= 0 {
		return 30 * time.Second
	}
	return d
}

// ResolvePassphrase returns the SSH key passphrase in this priority order:
//  1. COOLIFY_SSH_PASSPHRASE env var
//  2. Interactive prompt when --ssh-passphrase-prompt is set
//  3. nil (no passphrase)
func (f *SSHMeshFlags) ResolvePassphrase() ([]byte, error) {
	if env := os.Getenv("COOLIFY_SSH_PASSPHRASE"); env != "" {
		return []byte(env), nil
	}
	if f.SSHPassphrasePrompt {
		fmt.Fprint(os.Stderr, "SSH key passphrase: ")
		pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return nil, fmt.Errorf("read passphrase: %w", err)
		}
		return pass, nil
	}
	return nil, nil
}

// BuildSSHClient creates an SSH client, resolving any key passphrase first.
func (f *SSHMeshFlags) BuildSSHClient() (*internalssh.Client, error) {
	passphrase, err := f.ResolvePassphrase()
	if err != nil {
		return nil, err
	}
	return internalssh.NewClient(f.SSHKey, passphrase, f.ParseSSHTimeout())
}

// Validate checks that the required flags are set.
func (f *SSHMeshFlags) Validate() error {
	if len(f.Servers) == 0 {
		return fmt.Errorf("--servers is required")
	}
	if f.SSHKey == "" {
		return fmt.Errorf("--ssh-key is required")
	}
	return nil
}

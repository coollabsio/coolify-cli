// Package initcmd implements the `coolify init` alpha WireGuard mesh
// bootstrap command tree (Coolify v5).
package initcmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	internalssh "github.com/coollabsio/coolify-cli/internal/ssh"
)

// InitFlags holds all flags shared between `plan` and `apply`.
type InitFlags struct {
	Servers             []string
	SSHKey              string
	SSHUser             string
	SSHPort             int
	SSHPassphrasePrompt bool
	WGMgmtPool          string
	ContainerPool       string
	ContainerPrefix     int
	WGInterface         string
	WGListenPort        int
	InstallPodman         bool
	PodmanNetworkName     string
	DefaultDenyContainers bool
	InstallCoold          bool
	CooldBinaryPath       string
	CorrosionBinaryPath   string
	CorrosionGossipPort   int
	CorrosionAPIPort      int
	Concurrency         int
	SSHTimeout          string
	Yes                 bool
}

// ParseSSHTimeout parses the SSHTimeout string into a time.Duration.
func (f *InitFlags) ParseSSHTimeout() time.Duration {
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
func (f *InitFlags) ResolvePassphrase() ([]byte, error) {
	if env := os.Getenv("COOLIFY_SSH_PASSPHRASE"); env != "" {
		return []byte(env), nil
	}
	if f.SSHPassphrasePrompt {
		fmt.Fprint(os.Stderr, "SSH key passphrase: ")
		pass, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // newline after hidden input
		if err != nil {
			return nil, fmt.Errorf("read passphrase: %w", err)
		}
		return pass, nil
	}
	return nil, nil
}

// BuildSSHClient creates an SSH client, resolving any key passphrase first.
func (f *InitFlags) BuildSSHClient() (*internalssh.Client, error) {
	passphrase, err := f.ResolvePassphrase()
	if err != nil {
		return nil, err
	}
	return internalssh.NewClient(f.SSHKey, passphrase, f.ParseSSHTimeout())
}

// bindInitFlags registers all shared flags as PersistentFlags on cmd.
func bindInitFlags(cmd *cobra.Command, f *InitFlags) {
	pf := cmd.PersistentFlags()

	pf.StringSliceVar(&f.Servers, "servers", nil,
		"Comma-separated server IPs to include in the mesh (required for plan/apply)")
	pf.StringVar(&f.SSHKey, "ssh-key", "",
		"Path to SSH private key used to connect to servers (required for plan/apply)")
	pf.StringVar(&f.SSHUser, "ssh-user", "root",
		"SSH username")
	pf.IntVar(&f.SSHPort, "ssh-port", 22,
		"SSH port")
	pf.BoolVar(&f.SSHPassphrasePrompt, "ssh-passphrase-prompt", false,
		"Prompt for SSH key passphrase (also reads COOLIFY_SSH_PASSPHRASE env var)")
	pf.StringVar(&f.WGMgmtPool, "wg-mgmt-pool", "100.64.0.0/16",
		"WireGuard management address pool — each host gets a /32 from here, assigned to wg0")
	pf.StringVar(&f.ContainerPool, "container-pool", "10.210.0.0/16",
		"Container address pool — each host gets a /<container-prefix> from here, owned by the Podman bridge")
	pf.IntVar(&f.ContainerPrefix, "container-prefix", 24,
		"Prefix length of the per-host container subnet (e.g. 24 → /24, 254 usable container IPs per host)")
	pf.StringVar(&f.WGInterface, "wg-interface", "wg0",
		"WireGuard interface name on the remote hosts")
	pf.IntVar(&f.WGListenPort, "wg-listen-port", 51820,
		"WireGuard UDP listen port")
	pf.BoolVar(&f.InstallPodman, "podman", false,
		"Install Podman, enable its socket, create a per-host bridge network, install firewall rules, and enable IP forwarding")
	pf.StringVar(&f.PodmanNetworkName, "podman-network", "coolify-mesh",
		"Name of the Podman bridge network created on each host (requires --podman)")
	pf.BoolVar(&f.DefaultDenyContainers, "default-deny", false,
		"With --podman: install default-deny iptables rules for CROSS-HOST container traffic (between hosts via wg0). Intra-host (same bridge) traffic is NOT enforced — defer to per-app podman networks. The v5 control plane manages allows in the COOLIFY-ALLOW chain on the host that owns each destination IP")
	pf.BoolVar(&f.InstallCoold, "install-coold", false,
		"Install the Coolify v5 control-plane agents (corrosion + coold). Requires --podman. Uploads the binaries from --corrosion-binary / --coold-binary")
	pf.StringVar(&f.CooldBinaryPath, "coold-binary",
		os.ExpandEnv("$HOME/devel/coold/target/release/coold"),
		"Local path to the coold Linux/arm64 binary (used with --install-coold)")
	pf.StringVar(&f.CorrosionBinaryPath, "corrosion-binary",
		os.ExpandEnv("$HOME/devel/corrosion/target/release/corrosion"),
		"Local path to the corrosion Linux/arm64 binary (used with --install-coold)")
	pf.IntVar(&f.CorrosionGossipPort, "corrosion-gossip-port", 8787,
		"Corrosion SWIM gossip port (bound to the wg0 mgmt IP)")
	pf.IntVar(&f.CorrosionAPIPort, "corrosion-api-port", 8080,
		"Corrosion HTTP API port (bound to 127.0.0.1)")
	pf.IntVar(&f.Concurrency, "concurrency", 10,
		"Maximum number of parallel SSH connections")
	pf.StringVar(&f.SSHTimeout, "ssh-timeout", "30s",
		"SSH connection timeout (e.g. 30s, 1m)")
	pf.BoolVarP(&f.Yes, "yes", "y", false,
		"Skip the interactive alpha confirmation prompt")
}

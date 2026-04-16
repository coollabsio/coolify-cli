// Package ssh provides a thin SSH client and parallel fanout helper
// for the coolify init mesh-bootstrap commands.
package ssh

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// Runner executes a shell command on a remote host and returns its
// stdout, stderr, and exit error.  It is an interface so tests can
// inject a fake implementation without opening real SSH connections.
type Runner interface {
	Run(ctx context.Context, host, user string, port int, cmd string) (stdout, stderr string, err error)
}

// Client implements Runner using the golang.org/x/crypto/ssh library.
// Keys must be unencrypted PEM files.
// NOTE: host-key verification is intentionally disabled in v1 (alpha).
// This is acceptable for a bootstrap tool in controlled environments
// and should be improved in a future release.
type Client struct {
	signer  gossh.Signer
	timeout time.Duration
}

// NewClient loads the private key at keyPath and returns a Client ready to
// SSH into hosts.  If passphrase is non-nil it is used to decrypt the key;
// pass nil for unencrypted keys.
func NewClient(keyPath string, passphrase []byte, timeout time.Duration) (*Client, error) {
	raw, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read SSH key %q: %w", keyPath, err)
	}

	var signer gossh.Signer
	if len(passphrase) > 0 {
		signer, err = gossh.ParsePrivateKeyWithPassphrase(raw, passphrase)
	} else {
		signer, err = gossh.ParsePrivateKey(raw)
	}
	if err != nil {
		// Give the user an actionable hint when the key is passphrase-protected.
		if isPassphraseError(err) {
			return nil, fmt.Errorf("SSH key %q is passphrase-protected — use --ssh-passphrase-prompt or set COOLIFY_SSH_PASSPHRASE: %w", keyPath, err)
		}
		return nil, fmt.Errorf("parse SSH key %q: %w", keyPath, err)
	}

	return &Client{
		signer:  signer,
		timeout: timeout,
	}, nil
}

// isPassphraseError returns true when err is the "passphrase protected" error
// returned by golang.org/x/crypto/ssh.
func isPassphraseError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "passphrase") || contains(msg, "encrypted")
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) &&
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}()
}

// Run connects to host:port over SSH as user, executes cmd, and returns
// the combined stdout, stderr, and any error.  The connection is
// closed when the command finishes or ctx is cancelled.
func (c *Client) Run(ctx context.Context, host, user string, port int, cmd string) (string, string, error) {
	cfg := &gossh.ClientConfig{
		User:            user,
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(c.signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec // alpha v1, documented limitation
		Timeout:         c.timeout,
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))

	// Use a context-aware dialer so the dial itself respects ctx.
	dialer := &net.Dialer{Timeout: c.timeout}
	netConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return "", "", fmt.Errorf("dial %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := gossh.NewClientConn(netConn, addr, cfg)
	if err != nil {
		netConn.Close()
		return "", "", fmt.Errorf("SSH handshake %s: %w", addr, err)
	}
	conn := gossh.NewClient(sshConn, chans, reqs)
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("SSH new session on %s: %w", addr, err)
	}
	defer sess.Close()

	var stdout, stderr bytes.Buffer
	sess.Stdout = &stdout
	sess.Stderr = &stderr

	if err := sess.Start(cmd); err != nil {
		return "", "", fmt.Errorf("SSH start on %s: %w", addr, err)
	}

	waitDone := make(chan error, 1)
	go func() { waitDone <- sess.Wait() }()

	select {
	case <-ctx.Done():
		// Best-effort signal; ignore error since we're already cancelled.
		_ = sess.Signal(gossh.SIGTERM)
		return stdout.String(), stderr.String(), ctx.Err()
	case runErr := <-waitDone:
		return stdout.String(), stderr.String(), runErr
	}
}

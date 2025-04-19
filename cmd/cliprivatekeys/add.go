package cliprivatekeys

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/coollabsio/cli-coolify/pkg/gen/openapi"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// addKeyModel is the Bubble Tea model for the interactive add key form
type addKeyModel struct {
	nameInput  textinput.Model
	keyInput   textinput.Model
	focusIndex int
	done       bool
	err        error
	coolify    *runtime.Coolify
}

func initialAddKeyModel(coolify *runtime.Coolify) addKeyModel {
	m := addKeyModel{
		coolify: coolify,
	}

	// Setup name input
	m.nameInput = tui.NewFocusedInput("My SSH Key", "› ")
	m.nameInput.CharLimit = 50
	m.nameInput.Width = 40

	// Setup key input (multi-line)
	m.keyInput = tui.NewBlurredInput("SSH private key or path to key file", "› ")
	m.keyInput.CharLimit = 4096
	m.keyInput.Width = 60

	return m
}

func (m addKeyModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addKeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			// Submit on enter when key input is focused
			if msg.String() == "enter" && m.focusIndex == 1 {
				m.done = true
				return m, tea.Quit
			}

			// Cycle focus between inputs
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = 1
				}
			} else {
				m.focusIndex++
				if m.focusIndex > 1 {
					m.focusIndex = 0
				}
			}

			var cmds []tea.Cmd
			if m.focusIndex == 0 {
				m.nameInput.PromptStyle = tui.FocusedStyle
				m.nameInput.TextStyle = tui.FocusedStyle
				m.keyInput.PromptStyle = tui.BlurredStyle
				m.keyInput.TextStyle = tui.BlurredStyle
				cmds = append(cmds, m.nameInput.Focus())
				m.keyInput.Blur()
			} else {
				m.keyInput.PromptStyle = tui.FocusedStyle
				m.keyInput.TextStyle = tui.FocusedStyle
				m.nameInput.PromptStyle = tui.BlurredStyle
				m.nameInput.TextStyle = tui.BlurredStyle
				cmds = append(cmds, m.keyInput.Focus())
				m.nameInput.Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input for the active input
	var cmds []tea.Cmd
	if m.focusIndex == 0 {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.keyInput, cmd = m.keyInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m addKeyModel) View() string {
	var b strings.Builder

	// Title with Coolify branding
	title := tui.FocusedStyle.Bold(true).Render("Add New SSH Private Key")
	b.WriteString(title + "\n\n")

	// Render inputs with labels
	labelStyle := tui.BlurredStyle.Width(12)

	b.WriteString(labelStyle.Render("Name:") + " " + m.nameInput.View() + "\n\n")
	b.WriteString(labelStyle.Render("Private Key:") + " " + m.keyInput.View() + "\n\n")

	// Instructions
	b.WriteString("\n" + tui.BlurredStyle.Render("(Tab/Shift+Tab to navigate, Enter to submit)"))

	return b.String()
}

func generateRSAKeyPair() ([]byte, []byte, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Convert private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// Generate public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	return privateKeyBytes, publicKeyBytes, nil
}

func generateEd25519KeyPair() ([]byte, []byte, error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}
	privateKeyPem, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPem)

	// Generate public key
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate public key: %w", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)

	return privateKeyBytes, publicKeyBytes, nil
}

func (c *cliPrivateKeys) generateKeyPair(name, outputDir, alorithim string, force bool) (string, error) {
	var privateKey, publicKey []byte
	var err error
	switch alorithim {
	case "rsa":
		privateKey, publicKey, err = generateRSAKeyPair()
	case "ed25519":
		privateKey, publicKey, err = generateEd25519KeyPair()
	default:
		return "", fmt.Errorf("invalid alorithim: %s", alorithim)
	}

	if err != nil {
		return "", err
	}

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0700); err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write private key file
		privateKeyPath := filepath.Join(outputDir, name)
		if !force {
			if _, err := os.Stat(privateKeyPath); err == nil {
				return "", fmt.Errorf("private key file already exists: %s", privateKeyPath)
			}
		}

		if err := os.WriteFile(privateKeyPath, privateKey, 0600); err != nil {
			return "", fmt.Errorf("failed to write private key file: %w", err)
		}

		// Write public key file
		publicKeyPath := privateKeyPath + ".pub"
		if err := os.WriteFile(publicKeyPath, publicKey, 0644); err != nil {
			return "", fmt.Errorf("failed to write public key file: %w", err)
		}

		fmt.Printf("Generated SSH key pair:\n")
		fmt.Printf("  Private key: %s\n", privateKeyPath)
		fmt.Printf("  Public key:  %s\n", publicKeyPath)
	}
	return string(privateKey), nil
}

func (c *cliPrivateKeys) newAddCommand() *cobra.Command {
	var generateKeyPair bool
	var outPutDirectory string
	var algorithm string
	var force bool
	cmd := &cobra.Command{
		Use:   "add [name] [private_key_or_file]",
		Short: "Add a new private key",
		Long: `Add a new SSH private key to your Coolify instance.
The key can be provided directly as a string or as a path to a file.
Use --generate to create a new SSH key pair.

If no arguments are provided, an interactive form will be used.`,
		Example: utils.GetCommandExample(`
%[1]s private-keys add "My Key" /path/to/id_rsa
%[1]s private-keys add "My Key" "-----BEGIN RSA PRIVATE KEY-----..."
%[1]s private-keys add "My Key" --generate  # Generate key pair
%[1]s private-keys add  # Interactive mode
`),
		SilenceUsage: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if generateKeyPair {
				if len(args) != 1 {
					return fmt.Errorf("when using --generate, provide only the key name")
				}
				return nil
			}
			return cobra.RangeArgs(0, 2)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle key generation
			if generateKeyPair {
				name := args[0]
				privateKey, err := c.generateKeyPair(name, outPutDirectory, algorithm, force)
				if err != nil {
					return err
				}
				return c.addPrivateKey(cmd.Context(), name, privateKey)
			}

			// Interactive mode when no arguments are provided
			if len(args) == 0 {
				model := initialAddKeyModel(c.coolify())
				p := tea.NewProgram(model)
				finalModel, err := p.Run()
				if err != nil {
					return fmt.Errorf("error running interactive mode: %w", err)
				}

				// Process the final model after user submission
				finalState := finalModel.(addKeyModel)
				if !finalState.done {
					return fmt.Errorf("operation canceled")
				}

				name := finalState.nameInput.Value()
				privateKeyInput := finalState.keyInput.Value()

				return c.addPrivateKey(cmd.Context(), name, privateKeyInput)
			}

			// CLI mode with arguments
			if len(args) != 2 {
				return fmt.Errorf("requires both NAME and PRIVATE_KEY_OR_FILE arguments")
			}

			name := args[0]
			privateKeyInput := args[1]

			return c.addPrivateKey(cmd.Context(), name, privateKeyInput)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolVarP(&generateKeyPair, "generate", "g", false, "generate a new key pair")
	flags.StringVarP(&algorithm, "algorithm", "a", "rsa", "algorithm to use for the key pair")
	flags.StringVarP(&outPutDirectory, "output", "o", "", "optional output directory for the key pair")
	flags.BoolVarP(&force, "force", "f", false, "force the generation of the key pair if the name exists on the file system within the output directory")
	return cmd
}

// addPrivateKey adds a private key to the Coolify instance
func (c *cliPrivateKeys) addPrivateKey(ctx context.Context, name, privateKeyInput string) error {
	// Check if input is a file path
	var privateKey string
	if _, err := os.Stat(privateKeyInput); err == nil {
		keyBytes, err := os.ReadFile(privateKeyInput)
		if err != nil {
			return fmt.Errorf("error reading private key file: %w", err)
		}
		privateKey = string(keyBytes)
	} else {
		privateKey = privateKeyInput
	}

	req, err := c.coolify().Client.CreatePrivateKey(ctx, openapi.CreatePrivateKeyJSONRequestBody{
		Name:       &name,
		PrivateKey: privateKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	parsedResponse, err := openapi.ParseCreatePrivateKeyResponse(req)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if parsedResponse.StatusCode() != http.StatusCreated {
		return fmt.Errorf("failed to add private key: %s", string(parsedResponse.Body))
	}

	fmt.Printf("Private key '%s' added successfully as UUID: %s\n", name, *parsedResponse.JSON201.Uuid)
	return nil
}

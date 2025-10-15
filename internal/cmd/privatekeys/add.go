package privatekeys

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

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/internal/client"
	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/coollabsio/cli-coolify/internal/tui"
	"github.com/coollabsio/cli-coolify/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// addKeyMap defines keybindings for the add private key form
type addKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Tab   key.Binding
	Enter key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k addKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k addKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab}, // first column
		{k.Enter, k.Help},     // second column
		{k.Quit},              // third column
	}
}

var addKeys = addKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab", "next field"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit/select"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	),
}

// addKeyModel is the Bubble Tea model for the interactive add key form
type addKeyModel struct {
	nameInput  textinput.Model
	keyInput   textinput.Model
	focusIndex int
	done       bool
	err        error
	coolify    *config.Coolify
	keys       addKeyMap
	help       help.Model
}

func initialAddKeyModel(coolify *config.Coolify) addKeyModel {
	m := addKeyModel{
		coolify: coolify,
		keys:    addKeys,
		help:    help.New(),
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
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		if key.Matches(msg, m.keys.Help) {
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		if key.Matches(msg, m.keys.Enter) {
			// Submit on enter when key input is focused
			if m.focusIndex == 1 {
				m.done = true
				return m, tea.Quit
			}
			// Otherwise move to next input
			m.focusIndex++
			if m.focusIndex > 1 {
				m.focusIndex = 0
			}
			return m, m.updateFocus()
		}

		if key.Matches(msg, m.keys.Tab) {
			// Cycle focus between inputs
			if msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > 1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = 1
			}
			return m, m.updateFocus()
		}

		if key.Matches(msg, m.keys.Up) {
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = 1
			}
			return m, m.updateFocus()
		}

		if key.Matches(msg, m.keys.Down) {
			m.focusIndex++
			if m.focusIndex > 1 {
				m.focusIndex = 0
			}
			return m, m.updateFocus()
		}
	}

	// Handle character input for the active input
	if m.focusIndex == 0 {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	} else {
		var cmd tea.Cmd
		m.keyInput, cmd = m.keyInput.Update(msg)
		return m, cmd
	}
}

func (m addKeyModel) updateFocus() tea.Cmd {
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

	return tea.Batch(cmds...)
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

	// Add help view
	if m.help.ShowAll {
		b.WriteString("\n\n")
		b.WriteString(m.help.View(m.keys))
	} else {
		b.WriteString("\n\n")
		b.WriteString(m.help.ShortHelpView(m.keys.ShortHelp()))
	}

	return b.String()
}

func generateRSAKeyPair() (privateBytes, publicBytes []byte, err error) {
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
	privateBytes = pem.EncodeToMemory(privateKeyPEM)

	// Generate public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate public key: %w", err)
	}
	publicBytes = ssh.MarshalAuthorizedKey(publicKey)

	return privateBytes, publicBytes, nil
}

func generateEd25519KeyPair() (privateBytes, publicBytes []byte, err error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}
	privateKeyPem, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privateBytes = pem.EncodeToMemory(privateKeyPem)

	// Generate public key
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate public key: %w", err)
	}
	publicBytes = ssh.MarshalAuthorizedKey(sshPublicKey)

	return privateBytes, publicBytes, nil
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
		if err := os.MkdirAll(outputDir, 0o700); err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write private key file
		privateKeyPath := filepath.Join(outputDir, name)
		if !force {
			if _, err := os.Stat(privateKeyPath); err == nil {
				return "", fmt.Errorf("private key file already exists: %s", privateKeyPath)
			}
		}

		if err := os.WriteFile(privateKeyPath, privateKey, 0o600); err != nil {
			return "", fmt.Errorf("failed to write private key file: %w", err)
		}

		// Write public key file
		publicKeyPath := privateKeyPath + ".pub"
		if err := os.WriteFile(publicKeyPath, publicKey, 0o644); err != nil {
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

	req, err := c.coolify().Client.CreatePrivateKey(ctx, client.CreatePrivateKeyJSONRequestBody{
		Name:       &name,
		PrivateKey: privateKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	parsedResponse, err := client.ParseCreatePrivateKeyResponse(req)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if parsedResponse.StatusCode() != http.StatusCreated {
		return fmt.Errorf("failed to add private key: %s", string(parsedResponse.Body))
	}

	fmt.Printf("Private key '%s' added successfully as UUID: %s\n", name, *parsedResponse.JSON201.Uuid)
	return nil
}

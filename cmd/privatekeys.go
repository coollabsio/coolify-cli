package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var privateKeysCmd = &cobra.Command{
	Use:   "private-keys",
	Short: "Private key related commands",
}

var listPrivateKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all private keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		keySvc := service.NewPrivateKeyService(client)
		keys, err := keySvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list private keys: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		if err := formatter.Format(keys); err != nil {
			return err
		}

		if !showSensitive && format == output.FormatTable {
			fmt.Println("\nNote: Use -s to show sensitive information.")
		}

		return nil
	},
}

var addPrivateKeyCmd = &cobra.Command{
	Use:     "add <name> <private_key_or_file>",
	Example: `add mykey ~/.ssh/id_rsa`,
	Args:    cobra.ExactArgs(2),
	Short:   "Add a private key",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		name := args[0]
		privateKeyInput := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		var privateKey string
		// Check if input is a file path
		if _, err := os.Stat(privateKeyInput); err == nil {
			keyBytes, err := os.ReadFile(privateKeyInput)
			if err != nil {
				return fmt.Errorf("error reading private key file: %w", err)
			}
			privateKey = string(keyBytes)
		} else {
			privateKey = privateKeyInput
		}

		keySvc := service.NewPrivateKeyService(client)
		req := models.PrivateKeyCreateRequest{
			Name:       name,
			PrivateKey: privateKey,
		}

		key, err := keySvc.Create(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to add private key: %w", err)
		}

		fmt.Printf("Private key '%s' added successfully (UUID: %s)\n", key.Name, key.UUID)
		return nil
	},
}

var removePrivateKeyCmd = &cobra.Command{
	Use:   "remove <uuid>",
	Args:  cobra.ExactArgs(1),
	Short: "Remove a private key",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		keySvc := service.NewPrivateKeyService(client)
		err = keySvc.Delete(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to remove private key: %w", err)
		}

		fmt.Println("Private key removed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(privateKeysCmd)
	privateKeysCmd.AddCommand(listPrivateKeysCmd)
	privateKeysCmd.AddCommand(addPrivateKeyCmd)
	privateKeysCmd.AddCommand(removePrivateKeyCmd)
}

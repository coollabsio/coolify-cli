package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func providerOutput(cmd *cobra.Command, value any) error {
	format, _ := cmd.Flags().GetString("format")
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
	formatter, err := output.NewFormatter(format, output.Options{ShowSensitive: showSensitive})
	if err != nil {
		return err
	}
	return formatter.Format(value)
}

func providerOptionCommand(use, short string, fetch func(*cobra.Command, string) (any, error)) *cobra.Command {
	return &cobra.Command{Use: use + " <cloud_token_uuid>", Args: cli.ExactArgs(1, "<cloud_token_uuid>"), Short: short, RunE: func(cmd *cobra.Command, args []string) error {
		value, err := fetch(cmd, args[0])
		if err != nil {
			return err
		}
		return providerOutput(cmd, value)
	}}
}

func NewHetznerCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "hetzner", Short: "Provision servers with Hetzner"}
	with := func(call func(*service.HetznerService, *cobra.Command, string) (any, error)) func(*cobra.Command, string) (any, error) {
		return func(cmd *cobra.Command, token string) (any, error) {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return nil, fmt.Errorf("failed to get API client: %w", err)
			}
			return call(service.NewHetznerService(client), cmd, token)
		}
	}
	cmd.AddCommand(
		providerOptionCommand("locations", "List Hetzner locations", with(func(s *service.HetznerService, cmd *cobra.Command, token string) (any, error) {
			return s.Locations(cmd.Context(), token)
		})),
		providerOptionCommand("server-types", "List Hetzner server types", with(func(s *service.HetznerService, cmd *cobra.Command, token string) (any, error) {
			return s.ServerTypes(cmd.Context(), token)
		})),
		providerOptionCommand("images", "List Hetzner images", with(func(s *service.HetznerService, cmd *cobra.Command, token string) (any, error) {
			return s.Images(cmd.Context(), token)
		})),
		providerOptionCommand("ssh-keys", "List Hetzner SSH keys", with(func(s *service.HetznerService, cmd *cobra.Command, token string) (any, error) {
			return s.SSHKeys(cmd.Context(), token)
		})),
		providerOptionCommand("firewalls", "List Hetzner firewalls", with(func(s *service.HetznerService, cmd *cobra.Command, token string) (any, error) {
			return s.Firewalls(cmd.Context(), token)
		})),
		providerOptionCommand("networks", "List Hetzner networks", with(func(s *service.HetznerService, cmd *cobra.Command, token string) (any, error) {
			return s.Networks(cmd.Context(), token)
		})),
		newHetznerCreateCommand(),
	)
	return cmd
}

func newHetznerCreateCommand() *cobra.Command {
	var req models.HetznerServerCreateRequest
	cmd := &cobra.Command{Use: "create", Short: "Create a Hetzner server", RunE: func(cmd *cobra.Command, _ []string) error {
		if req.CloudProviderTokenUUID == "" || req.Location == "" || req.ServerType == "" || req.Image == 0 || req.PrivateKeyUUID == "" {
			return fmt.Errorf("--cloud-token, --location, --server-type, --image, and --private-key are required")
		}
		if !req.EnableIPv4 && !req.EnableIPv6 {
			return fmt.Errorf("at least one of --enable-ipv4 or --enable-ipv6 must be true")
		}
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		response, err := service.NewHetznerService(client).Create(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create Hetzner server: %w", err)
		}
		return providerOutput(cmd, response)
	}}
	cmd.Flags().StringVar(&req.CloudProviderTokenUUID, "cloud-token", "", "Cloud provider token UUID")
	cmd.Flags().StringVar(&req.Location, "location", "", "Hetzner location name")
	cmd.Flags().StringVar(&req.ServerType, "server-type", "", "Hetzner server type")
	cmd.Flags().IntVar(&req.Image, "image", 0, "Hetzner image ID")
	cmd.Flags().StringVar(&req.Name, "name", "", "Server name (auto-generated if omitted)")
	cmd.Flags().StringVar(&req.PrivateKeyUUID, "private-key", "", "Coolify private key UUID")
	cmd.Flags().BoolVar(&req.EnableIPv4, "enable-ipv4", true, "Enable public IPv4")
	cmd.Flags().BoolVar(&req.EnableIPv6, "enable-ipv6", true, "Enable public IPv6")
	cmd.Flags().BoolVar(&req.EnableBackups, "enable-backups", false, "Enable Hetzner backups")
	cmd.Flags().IntSliceVar(&req.HetznerSSHKeyIDs, "ssh-key-ids", nil, "Additional Hetzner SSH key IDs")
	cmd.Flags().IntSliceVar(&req.HetznerFirewallIDs, "firewall-ids", nil, "Hetzner firewall IDs")
	cmd.Flags().IntSliceVar(&req.HetznerNetworkIDs, "network-ids", nil, "Hetzner network IDs")
	cmd.Flags().StringVar(&req.CloudInitScript, "cloud-init", "", "Cloud-init YAML")
	cmd.Flags().BoolVar(&req.InstantValidate, "validate", false, "Validate after creation")
	return cmd
}

func NewDigitalOceanCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "digitalocean", Aliases: []string{"digital-ocean", "do"}, Short: "Provision servers with DigitalOcean"}
	with := func(call func(*service.DigitalOceanService, *cobra.Command, string) (any, error)) func(*cobra.Command, string) (any, error) {
		return func(cmd *cobra.Command, token string) (any, error) {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return nil, fmt.Errorf("failed to get API client: %w", err)
			}
			return call(service.NewDigitalOceanService(client), cmd, token)
		}
	}
	cmd.AddCommand(
		providerOptionCommand("regions", "List DigitalOcean regions", with(func(s *service.DigitalOceanService, cmd *cobra.Command, token string) (any, error) {
			return s.Regions(cmd.Context(), token)
		})),
		providerOptionCommand("sizes", "List DigitalOcean sizes", with(func(s *service.DigitalOceanService, cmd *cobra.Command, token string) (any, error) {
			return s.Sizes(cmd.Context(), token)
		})),
		providerOptionCommand("images", "List DigitalOcean images", with(func(s *service.DigitalOceanService, cmd *cobra.Command, token string) (any, error) {
			return s.Images(cmd.Context(), token)
		})),
		providerOptionCommand("ssh-keys", "List DigitalOcean SSH keys", with(func(s *service.DigitalOceanService, cmd *cobra.Command, token string) (any, error) {
			return s.SSHKeys(cmd.Context(), token)
		})), newDigitalOceanCreateCommand())
	return cmd
}

func newDigitalOceanCreateCommand() *cobra.Command {
	var req models.DigitalOceanServerCreateRequest
	var image string
	cmd := &cobra.Command{Use: "create", Short: "Create a DigitalOcean server", RunE: func(cmd *cobra.Command, _ []string) error {
		if req.CloudProviderTokenUUID == "" || req.Region == "" || req.Size == "" || image == "" || req.PrivateKeyUUID == "" {
			return fmt.Errorf("--cloud-token, --region, --size, --image, and --private-key are required")
		}
		req.Image = image
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		response, err := service.NewDigitalOceanService(client).Create(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create DigitalOcean server: %w", err)
		}
		return providerOutput(cmd, response)
	}}
	cmd.Flags().StringVar(&req.CloudProviderTokenUUID, "cloud-token", "", "Cloud provider token UUID")
	cmd.Flags().StringVar(&req.Region, "region", "", "DigitalOcean region")
	cmd.Flags().StringVar(&req.Size, "size", "", "DigitalOcean size")
	cmd.Flags().StringVar(&image, "image", "", "DigitalOcean image ID or slug")
	cmd.Flags().StringVar(&req.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&req.PrivateKeyUUID, "private-key", "", "Coolify private key UUID")
	cmd.Flags().BoolVar(&req.EnableIPv6, "enable-ipv6", true, "Enable public IPv6")
	cmd.Flags().BoolVar(&req.Monitoring, "monitoring", true, "Enable DigitalOcean monitoring")
	cmd.Flags().IntSliceVar(&req.DigitalOceanSSHKeyIDs, "ssh-key-ids", nil, "Additional DigitalOcean SSH key IDs")
	cmd.Flags().StringVar(&req.CloudInitScript, "cloud-init", "", "Cloud-init YAML")
	cmd.Flags().BoolVar(&req.InstantValidate, "validate", false, "Validate after creation")
	return cmd
}

func NewVultrCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "vultr", Short: "Provision servers with Vultr"}
	with := func(call func(*service.VultrService, *cobra.Command, string) (any, error)) func(*cobra.Command, string) (any, error) {
		return func(cmd *cobra.Command, token string) (any, error) {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return nil, fmt.Errorf("failed to get API client: %w", err)
			}
			return call(service.NewVultrService(client), cmd, token)
		}
	}
	cmd.AddCommand(
		providerOptionCommand("regions", "List Vultr regions", with(func(s *service.VultrService, cmd *cobra.Command, token string) (any, error) {
			return s.Regions(cmd.Context(), token)
		})),
		providerOptionCommand("plans", "List Vultr plans", with(func(s *service.VultrService, cmd *cobra.Command, token string) (any, error) {
			return s.Plans(cmd.Context(), token)
		})),
		providerOptionCommand("os", "List Vultr operating systems", with(func(s *service.VultrService, cmd *cobra.Command, token string) (any, error) {
			return s.OperatingSystems(cmd.Context(), token)
		})),
		providerOptionCommand("ssh-keys", "List Vultr SSH keys", with(func(s *service.VultrService, cmd *cobra.Command, token string) (any, error) {
			return s.SSHKeys(cmd.Context(), token)
		})), newVultrCreateCommand())
	return cmd
}

func newVultrCreateCommand() *cobra.Command {
	var req models.VultrServerCreateRequest
	cmd := &cobra.Command{Use: "create", Short: "Create a Vultr server", RunE: func(cmd *cobra.Command, _ []string) error {
		if req.CloudProviderTokenUUID == "" || req.Region == "" || req.Plan == "" || req.OSID == 0 || req.PrivateKeyUUID == "" {
			return fmt.Errorf("--cloud-token, --region, --plan, --os-id, and --private-key are required")
		}
		if req.DisablePublicIPv4 && !req.EnableIPv6 {
			return fmt.Errorf("public IPv4 cannot be disabled unless IPv6 is enabled")
		}
		client, err := cli.GetAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}
		response, err := service.NewVultrService(client).Create(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("failed to create Vultr server: %w", err)
		}
		return providerOutput(cmd, response)
	}}
	cmd.Flags().StringVar(&req.CloudProviderTokenUUID, "cloud-token", "", "Cloud provider token UUID")
	cmd.Flags().StringVar(&req.Region, "region", "", "Vultr region")
	cmd.Flags().StringVar(&req.Plan, "plan", "", "Vultr plan")
	cmd.Flags().IntVar(&req.OSID, "os-id", 0, "Vultr operating system ID")
	cmd.Flags().StringVar(&req.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&req.PrivateKeyUUID, "private-key", "", "Coolify private key UUID")
	cmd.Flags().BoolVar(&req.EnableIPv6, "enable-ipv6", true, "Enable public IPv6")
	cmd.Flags().BoolVar(&req.DisablePublicIPv4, "disable-public-ipv4", false, "Disable public IPv4")
	cmd.Flags().StringSliceVar(&req.VultrSSHKeyIDs, "ssh-key-ids", nil, "Additional Vultr SSH key IDs")
	cmd.Flags().StringVar(&req.CloudInitScript, "cloud-init", "", "Cloud-init YAML")
	cmd.Flags().BoolVar(&req.InstantValidate, "validate", false, "Validate after creation")
	return cmd
}

package deployment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

const linkFileName = ".coolify.json"

// siteLink is the content of the link file: the application a folder deploys
// to, per context.
type siteLink struct {
	Context         string `json:"context"`
	ApplicationUUID string `json:"application_uuid"`
}

// FolderDeployResult is the outcome of `coolify deploy [folder]`.
type FolderDeployResult struct {
	ApplicationUUID string `json:"application_uuid"`
	DeploymentUUID  string `json:"deployment_uuid"`
	SHA             string `json:"sha"`
	Domains         string `json:"domains"`
	Created         bool   `json:"created"`
	Link            string `json:"link,omitempty" table:"-"`
}

const folderDeployLong = `Pack a folder of static files (default: the current folder) and deploy it as a static site.

The folder is uploaded as a .tar.gz. node_modules and hidden files and folders (.env, .git, ...) are skipped, except .well-known.
Symlinks are not allowed. Put index.html at the root of the folder.

The first deploy creates a static site application and writes .coolify.json in the
current working directory, so later deploys from there update the same application.
The link file stores the context name; in another context a new application is created.

Examples:
  coolify deploy ./dist
  coolify deploy ./dist --server <server_uuid> --project <project_uuid> --domain https://site.example.com
  coolify deploy ./dist --app <app_uuid>
  coolify deploy uuid <uuid>`

func addFolderDeployFlags(cmd *cobra.Command) {
	cmd.Flags().String("app", "", "Deploy to this static site application UUID instead of the linked one")
	cmd.Flags().Bool("no-link", false, "Do not write the link file .coolify.json")
	cmd.Flags().String("server", "", "Server UUID for a new application (default: the only usable server)")
	cmd.Flags().String("project", "", "Project UUID for a new application (default: the only project)")
	cmd.Flags().String("environment", "production", "Environment name for a new application")
	cmd.Flags().String("name", "", "Name of a new application (default: the current folder name plus a random ID)")
	cmd.Flags().String("domain", "", "Domains of a new application, comma-separated (default: a generated domain)")
	cmd.Flags().Bool("spa", false, "Serve index.html for unknown paths of a new application (single-page app)")
}

func runFolderDeploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	stderr := cmd.ErrOrStderr()

	dir := "."
	if len(args) == 1 {
		dir = args[0]
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fmt.Errorf("%q is not a folder; see 'coolify deploy --help' for the subcommands", dir)
	}

	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return fmt.Errorf("failed to get API client: %w", err)
	}
	contextName, err := cli.CurrentContextName(cmd)
	if err != nil {
		return err
	}

	appUUID, _ := cmd.Flags().GetString("app")
	if appUUID == "" {
		link, err := readLink()
		if err != nil {
			return err
		}
		switch {
		case link == nil:
		case link.Context == contextName:
			appUUID = link.ApplicationUUID
		default:
			fmt.Fprintf(stderr, "Warning: %s links context %q, not %q; creating a new application.\n", linkFileName, link.Context, contextName)
		}
	}

	var createReq *models.StaticArchiveCreateRequest
	if appUUID == "" {
		createReq, err = buildCreateRequest(ctx, cmd, client)
		if err != nil {
			return err
		}
	} else if changed := createOnlyFlagsChanged(cmd); len(changed) > 0 {
		fmt.Fprintf(stderr, "Warning: ignoring %s, which only apply to a new application; deploying to %s.\n", strings.Join(changed, ", "), appUUID)
	}

	site, err := packFolder(dir)
	if err != nil {
		return err
	}
	defer os.Remove(site.Path)
	if !site.HasIndex {
		fmt.Fprintf(stderr, "Warning: %s has no index.html at its root.\n", dir)
	}
	if len(site.Skipped) > 0 {
		fmt.Fprintf(stderr, "Skipped hidden files and folders: %s\n", strings.Join(site.Skipped, ", "))
	}
	fmt.Fprintf(stderr, "Packed %d files (%s) from %s\n", site.Files, cli.HumanBytes(site.TotalSize), dir)

	svc := service.NewStaticArchiveService(client)
	upload, err := uploadSite(cmd, svc, site.Path, appUUID)
	if err != nil {
		return err
	}

	result := FolderDeployResult{SHA: upload.SHA}
	if createReq != nil {
		createReq.UploadID = &upload.UploadID
		created, err := svc.Create(ctx, createReq)
		if err != nil {
			return err
		}
		result.ApplicationUUID = created.UUID
		result.DeploymentUUID = created.DeploymentUUID
		result.Domains = created.Domains
		result.Link = created.Link
		result.Created = true
	} else {
		req := &models.StaticArchiveDeployRequest{UploadID: upload.UploadID}
		deployed, err := svc.Deploy(ctx, appUUID, req)
		if err != nil {
			return err
		}
		result.ApplicationUUID = appUUID
		result.DeploymentUUID = deployed.DeploymentUUID
		result.Domains = deployed.Domains
		result.Link = deployed.Link
		if deployed.SHA != "" {
			result.SHA = deployed.SHA
		}
	}

	if noLink, _ := cmd.Flags().GetBool("no-link"); !noLink {
		if err := writeLink(siteLink{Context: contextName, ApplicationUUID: result.ApplicationUUID}); err != nil {
			return err
		}
	}

	format, _ := cmd.Flags().GetString("format")
	formatter, err := output.NewFormatter(format, output.Options{Writer: cmd.OutOrStdout()})
	if err != nil {
		return err
	}
	if err := formatter.Format(result); err != nil {
		return err
	}
	if format == output.FormatTable {
		if result.Link != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "\nOpen in Coolify: %s\n", result.Link)
		}
		if result.DeploymentUUID != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "\nFollow the build with: coolify app deployments logs %s %s\n", result.ApplicationUUID, result.DeploymentUUID)
		}
	}
	return nil
}

func createOnlyFlagsChanged(cmd *cobra.Command) []string {
	var changed []string
	for _, name := range []string{"server", "project", "environment", "name", "domain", "spa"} {
		if cmd.Flags().Changed(name) {
			changed = append(changed, "--"+name)
		}
	}
	return changed
}

func buildCreateRequest(ctx context.Context, cmd *cobra.Command, client *api.Client) (*models.StaticArchiveCreateRequest, error) {
	serverUUID, _ := cmd.Flags().GetString("server")
	if serverUUID == "" {
		uuid, err := onlyUsableServer(ctx, client)
		if err != nil {
			return nil, err
		}
		serverUUID = uuid
	}
	projectUUID, _ := cmd.Flags().GetString("project")
	if projectUUID == "" {
		uuid, err := onlyProject(ctx, client)
		if err != nil {
			return nil, err
		}
		projectUUID = uuid
	}

	environment, _ := cmd.Flags().GetString("environment")
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		name = defaultSiteName()
	}
	domains, _ := cmd.Flags().GetString("domain")
	req := &models.StaticArchiveCreateRequest{
		ProjectUUID:     projectUUID,
		ServerUUID:      serverUUID,
		EnvironmentName: environment,
		Name:            name,
		Domains:         domains,
	}
	if cmd.Flags().Changed("spa") {
		spa, _ := cmd.Flags().GetBool("spa")
		req.IsSpa = &spa
	}
	return req, nil
}

// defaultSiteName is the name of the current working directory (where the link
// file goes), plus a random ID, for example "my-app-k3x9p2".
func defaultSiteName() string {
	base := "static-site"
	if wd, err := os.Getwd(); err == nil {
		if dirName := filepath.Base(wd); dirName != string(filepath.Separator) {
			base = dirName
		}
	}
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 6)
	for i := range suffix {
		suffix[i] = alphabet[rand.IntN(len(alphabet))]
	}
	return base + "-" + string(suffix)
}

func onlyUsableServer(ctx context.Context, client *api.Client) (string, error) {
	servers, err := service.NewServerService(client).List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list servers: %w", err)
	}
	var usable []models.Server
	for _, server := range servers {
		if server.Settings.IsUsable && server.Settings.IsReachable && server.Settings.ServerRole != "build" {
			usable = append(usable, server)
		}
	}
	switch len(usable) {
	case 0:
		return "", errors.New("no usable server found; add or validate a server first, or pass --server <uuid>")
	case 1:
		return usable[0].UUID, nil
	}
	choices := make([]string, len(usable))
	for i, server := range usable {
		choices[i] = fmt.Sprintf("  %s  %s", server.UUID, server.Name)
	}
	return "", fmt.Errorf("more than one usable server; pass --server <uuid>:\n%s", strings.Join(choices, "\n"))
}

func onlyProject(ctx context.Context, client *api.Client) (string, error) {
	projects, err := service.NewProjectService(client).List(ctx)
	if err != nil {
		return "", err
	}
	switch len(projects) {
	case 0:
		return "", errors.New("no project found; create one with 'coolify project create', or pass --project <uuid>")
	case 1:
		return projects[0].UUID, nil
	}
	choices := make([]string, len(projects))
	for i, project := range projects {
		choices[i] = fmt.Sprintf("  %s  %s", project.UUID, project.Name)
	}
	return "", fmt.Errorf("more than one project; pass --project <uuid>:\n%s", strings.Join(choices, "\n"))
}

func uploadSite(cmd *cobra.Command, svc *service.StaticArchiveService, path, appUUID string) (*models.StaticArchiveUpload, error) {
	stderr := cmd.ErrOrStderr()
	showProgress := cli.IsTerminal(stderr)
	upload, err := svc.Upload(cmd.Context(), path, appUUID, func(sent, total int64) {
		if showProgress {
			fmt.Fprintf(stderr, "\rUploading: %3d%% (%s / %s)", sent*100/total, cli.HumanBytes(sent), cli.HumanBytes(total))
		}
	})
	if showProgress {
		fmt.Fprintln(stderr)
	}
	return upload, err
}

// readLink returns the link file of the current working directory, or nil
// when there is none.
func readLink() (*siteLink, error) {
	data, err := os.ReadFile(linkFileName)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", linkFileName, err)
	}
	var link siteLink
	if err := json.Unmarshal(data, &link); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON: %w", linkFileName, err)
	}
	if link.ApplicationUUID == "" {
		return nil, fmt.Errorf("%s has no application_uuid", linkFileName)
	}
	return &link, nil
}

func writeLink(link siteLink) error {
	data, err := json.MarshalIndent(link, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(linkFileName, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", linkFileName, err)
	}
	return os.Chmod(linkFileName, 0o600)
}

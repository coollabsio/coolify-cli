package models

// Registry credential sources reported by the server registries API.
const (
	ServerRegistrySourceAuths       = "auths"
	ServerRegistrySourceCredHelpers = "credHelpers"
)

// Registry login statuses shown in table output.
const (
	ServerRegistryStatusLoggedIn         = "Logged in"
	ServerRegistryStatusCredentialHelper = "Credential helper" //nolint:gosec // G101: status label, not a credential
	ServerRegistryStatusNotLoggedIn      = "Not logged in"
)

// ServerRegistryUser is a resource on the server that pulls from or pushes to a registry.
type ServerRegistryUser struct {
	Type string  `json:"type"`
	Name string  `json:"name"`
	Link *string `json:"link"`
}

// ServerRegistry is a Docker registry a server is logged in to or that its resources use.
// Credentials are never returned by the API, only the username.
type ServerRegistry struct {
	Registry string               `json:"registry"`
	LoggedIn bool                 `json:"logged_in"`
	Source   *string              `json:"source"`
	Username *string              `json:"username"`
	UsedBy   []ServerRegistryUser `json:"used_by"`
}

// ServerRegistriesResponse is the response of GET /servers/{uuid}/registries.
type ServerRegistriesResponse struct {
	Registries []ServerRegistry `json:"registries"`
	Error      *string          `json:"error"`
}

// ServerRegistryRow is the table view of a ServerRegistry.
// Username is masked by the command unless sensitive output is requested,
// so empty usernames stay empty instead of showing a mask.
type ServerRegistryRow struct {
	Registry string `json:"registry"`
	Status   string `json:"status"`
	Username string `json:"username"`
	UsedBy   string `json:"used_by"`
}

// ServerRegistryLoginRequest is the body of POST /servers/{uuid}/registries.
type ServerRegistryLoginRequest struct {
	Registry string `json:"registry"`
	Username string `json:"username"`
	Password string `json:"password"`
}

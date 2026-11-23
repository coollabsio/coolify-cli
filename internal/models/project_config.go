package models

import (
	"bytes"
	"encoding/json"
)

// ConfigLink is the link between an environment and its coolify.json file.
type ConfigLink struct {
	UUID                    string  `json:"uuid"`
	Mode                    string  `json:"mode"`
	Project                 string  `json:"project"`
	Environment             string  `json:"environment"`
	EnvironmentUUID         string  `json:"environment_uuid"`
	Repository              *string `json:"repository"`
	Branch                  *string `json:"branch"`
	FilePath                *string `json:"file_path"`
	ServerUUID              *string `json:"server_uuid"`
	DestinationUUID         *string `json:"destination_uuid" table:"-"`
	AutoApply               bool    `json:"auto_apply"`
	SuspendedAt             *string `json:"suspended_at"`
	LastAppliedRevisionUUID *string `json:"last_applied_revision_uuid"`
	CreatedAt               string  `json:"created_at" table:"-"`
	// WebhookSecret and WebhookURLs are only returned when a git link is created.
	WebhookSecret *string           `json:"webhook_secret,omitempty" table:"-" sensitive:"true"`
	WebhookURLs   map[string]string `json:"webhook_urls,omitempty" table:"-"`
}

// ConfigLinkCreateRequest is the body of POST /config/links.
type ConfigLinkCreateRequest struct {
	Project         string `json:"project"`
	Environment     string `json:"environment"`
	Mode            string `json:"mode"`
	AutoApply       *bool  `json:"auto_apply,omitempty"`
	Repository      string `json:"repository,omitempty"`
	Branch          string `json:"branch,omitempty"`
	FilePath        string `json:"file_path,omitempty"`
	SourceType      string `json:"source_type,omitempty"`
	SourceUUID      string `json:"source_uuid,omitempty"`
	ServerUUID      string `json:"server_uuid,omitempty"`
	DestinationUUID string `json:"destination_uuid,omitempty"`
}

// ConfigLinkUpdateRequest is the body of PATCH /config/links/{uuid}.
type ConfigLinkUpdateRequest struct {
	AutoApply       *bool  `json:"auto_apply,omitempty"`
	ServerUUID      string `json:"server_uuid,omitempty"`
	DestinationUUID string `json:"destination_uuid,omitempty"`
}

// ConfigRevision is one plan of a coolify.json file.
type ConfigRevision struct {
	UUID       string      `json:"uuid"`
	Trigger    string      `json:"trigger"`
	CommitSHA  *string     `json:"commit_sha"`
	FileHash   *string     `json:"file_hash" table:"-"`
	Status     string      `json:"status"`
	Error      *string     `json:"error"`
	Plan       *ConfigPlan `json:"plan" table:"-"`
	ApprovedAt *string     `json:"approved_at"`
	AppliedAt  *string     `json:"applied_at"`
	CreatedAt  string      `json:"created_at"`
}

// Revision statuses returned by the API.
const (
	ConfigRevisionWaitingApproval = "waiting_approval"
	ConfigRevisionFailed          = "failed"
)

// ConfigPlan is the result of comparing a coolify.json file with an environment.
type ConfigPlan struct {
	Changes         []ConfigChange   `json:"changes"`
	Errors          []ConfigError    `json:"errors"`
	Warnings        []string         `json:"warnings"`
	MissingRequired ConfigMissingMap `json:"missing_required"`
	StateHash       string           `json:"state_hash,omitempty"`
	Destructive     bool             `json:"destructive"`
	RemovesAll      bool             `json:"removes_all"`
}

// ConfigChange is one change of a plan. From and To are null when the token
// cannot read sensitive data or the value is secret.
type ConfigChange struct {
	Kind        string `json:"kind"`
	Key         string `json:"key"`
	Path        string `json:"path"`
	Destructive bool   `json:"destructive"`
	Secret      bool   `json:"secret"`
	From        any    `json:"from"`
	To          any    `json:"to"`
}

// ConfigError is one error of a coolify.json file, with its JSON path.
type ConfigError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// ConfigMissingMap maps a resource key to the names of its required variables
// that have no value. PHP encodes an empty map as [], so both forms decode.
type ConfigMissingMap map[string][]string

// UnmarshalJSON accepts an object or an empty array.
func (m *ConfigMissingMap) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("[]")) || bytes.Equal(trimmed, []byte("null")) {
		*m = ConfigMissingMap{}
		return nil
	}
	var raw map[string][]string
	if err := json.Unmarshal(trimmed, &raw); err != nil {
		return err
	}
	*m = raw
	return nil
}

// ConfigResumeResult is the response of a confirmed resume.
type ConfigResumeResult struct {
	Link     ConfigLink      `json:"link"`
	Revision *ConfigRevision `json:"revision,omitempty"`
}

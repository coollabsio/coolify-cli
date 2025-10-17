package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_MarshalUnmarshal(t *testing.T) {
	server := Server{
		ID:   1,
		UUID: "test-uuid",
		Name: "test-server",
		IP:   "192.168.1.100",
		User: "root",
		Port: 22,
		Settings: Settings{
			IsReachable: true,
			IsUsable:    true,
		},
	}

	// Marshal
	data, err := json.Marshal(server)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled Server
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, server.UUID, unmarshaled.UUID)
	assert.Equal(t, server.Name, unmarshaled.Name)
	assert.Equal(t, server.IP, unmarshaled.IP)
	assert.True(t, unmarshaled.Settings.IsReachable)
}

func TestServer_UnmarshalFromFixture(t *testing.T) {
	fixtureData, err := os.ReadFile(filepath.Join("..", "..", "test", "fixtures", "server.json"))
	require.NoError(t, err)

	var server Server
	err = json.Unmarshal(fixtureData, &server)

	require.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", server.UUID)
	assert.Equal(t, "production-server", server.Name)
	assert.Equal(t, "192.168.1.100", server.IP)
	assert.True(t, server.Settings.IsReachable)
}

func TestProject_MarshalUnmarshal(t *testing.T) {
	desc := "Test project"
	project := Project{
		UUID:        "proj-uuid",
		Name:        "My Project",
		Description: &desc,
		Environments: []Environment{
			{
				ID:   1,
				UUID: "env-uuid",
				Name: "production",
			},
		},
	}

	// Marshal
	data, err := json.Marshal(project)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled Project
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, project.UUID, unmarshaled.UUID)
	assert.Equal(t, project.Name, unmarshaled.Name)
	assert.NotNil(t, unmarshaled.Description)
	assert.Equal(t, "Test project", *unmarshaled.Description)
	assert.Len(t, unmarshaled.Environments, 1)
}

func TestProject_UnmarshalFromFixture(t *testing.T) {
	fixtureData, err := os.ReadFile(filepath.Join("..", "..", "test", "fixtures", "project.json"))
	require.NoError(t, err)

	var project Project
	err = json.Unmarshal(fixtureData, &project)

	require.NoError(t, err)
	assert.Equal(t, "proj-123-uuid", project.UUID)
	assert.Equal(t, "My Project", project.Name)
	assert.Len(t, project.Environments, 1)
	assert.Len(t, project.Environments[0].Applications, 1)
	assert.Equal(t, "running", project.Environments[0].Applications[0].Status)
}

func TestResource_MarshalUnmarshal(t *testing.T) {
	resource := Resource{
		ID:     1,
		UUID:   "resource-uuid",
		Name:   "test-resource",
		Type:   "application",
		Status: ResourceStatusRunning,
	}

	// Marshal
	data, err := json.Marshal(resource)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled Resource
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, resource.UUID, unmarshaled.UUID)
	assert.Equal(t, resource.Status, unmarshaled.Status)
}

func TestDeployment_MarshalUnmarshal(t *testing.T) {
	appName := "test-app"
	serverName := "test-server"
	commit := "abc123"
	deployment := Deployment{
		UUID:            "deployment-uuid",
		ApplicationName: &appName,
		ServerName:      &serverName,
		Status:          "running",
		Commit:          &commit,
	}

	// Marshal
	data, err := json.Marshal(deployment)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled Deployment
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, deployment.UUID, unmarshaled.UUID)
	assert.Equal(t, *deployment.ApplicationName, *unmarshaled.ApplicationName)
	assert.Equal(t, deployment.Status, unmarshaled.Status)
}

func TestDomain_MarshalUnmarshal(t *testing.T) {
	domain := Domain{
		IP:      "192.168.1.100",
		Domains: []string{"example.com", "www.example.com"},
	}

	// Marshal
	data, err := json.Marshal(domain)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled Domain
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, domain.IP, unmarshaled.IP)
	assert.Len(t, unmarshaled.Domains, 2)
}

func TestPrivateKey_MarshalUnmarshal(t *testing.T) {
	key := PrivateKey{
		ID:         1,
		UUID:       "key-uuid",
		Name:       "test-key",
		PublicKey:  "ssh-rsa AAAA...",
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
	}

	// Marshal
	data, err := json.Marshal(key)
	require.NoError(t, err)

	// Unmarshal
	var unmarshaled PrivateKey
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, key.UUID, unmarshaled.UUID)
	assert.Equal(t, key.Name, unmarshaled.Name)
}

func TestPrivateKeyCreateRequest_Marshal(t *testing.T) {
	request := PrivateKeyCreateRequest{
		Name:       "my-key",
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\n...",
	}

	data, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled PrivateKeyCreateRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Name, unmarshaled.Name)
	assert.Equal(t, request.PrivateKey, unmarshaled.PrivateKey)
}

func TestServerCreateRequest_Marshal(t *testing.T) {
	request := ServerCreateRequest{
		Name:            "new-server",
		IP:              "192.168.1.200",
		Port:            22,
		User:            "root",
		PrivateKeyUUID:  "key-uuid",
		InstantValidate: true,
	}

	data, err := json.Marshal(request)
	require.NoError(t, err)

	var unmarshaled ServerCreateRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, request.Name, unmarshaled.Name)
	assert.Equal(t, request.IP, unmarshaled.IP)
	assert.Equal(t, request.Port, unmarshaled.Port)
	assert.True(t, unmarshaled.InstantValidate)
}

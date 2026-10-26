package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/models"
)

var hostingerCreateArgs = []string{
	"hostinger", "create",
	"--cloud-token", "token-1",
	"--item-id", "hostingercom-vps-kvm1-usd-1m",
	"--data-center-id", "9",
	"--template-id", "1002",
	"--private-key", "key-1",
}

// runHostingerCommand executes `server <args>` against a local mock Coolify API
// using a temporary config file, never a real API.
func runHostingerCommand(t *testing.T, handler http.HandlerFunc, stdin string, args ...string) (string, string, error) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "coolify")
	require.NoError(t, os.MkdirAll(configDir, 0o700))
	config := `{"instances":[{"name":"test","fqdn":"` + server.URL + `","token":"api-token","default":true}]}`
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(config), 0o600))

	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().BoolP("show-sensitive", "s", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewServerCommand())

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(append([]string{"server"}, args...))
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func TestHostingerCatalogCommand_FlattensPricesIntoRows(t *testing.T) {
	stdout, _, err := runHostingerCommand(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/hostinger/catalog", r.URL.Path)
		assert.Equal(t, "token-1", r.URL.Query().Get("cloud_provider_token_uuid"))
		assert.Equal(t, "Bearer api-token", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`[{"id":"hostingercom-vps-kvm1","name":"KVM 1","category":"VPS","metadata":{"cpus":"1","memory":"4096","disk_space":"51200","bandwidth":"4096000","network":"300"},"prices":[{"id":"hostingercom-vps-kvm1-usd-1m","name":"monthly","currency":"USD","price":1399,"first_period_price":649,"period":1,"period_unit":"month"},{"id":"hostingercom-vps-kvm1-usd-1y","name":"yearly","currency":"USD","price":11988,"first_period_price":7788,"period":1,"period_unit":"year"}]}]`))
	}, "", "hostinger", "catalog", "token-1", "--format", "json")

	require.NoError(t, err)
	var rows []models.HostingerPlanPrice
	require.NoError(t, json.Unmarshal([]byte(stdout), &rows))
	require.Len(t, rows, 2)
	assert.Equal(t, "hostingercom-vps-kvm1-usd-1m", rows[0].ItemID)
	assert.Equal(t, "KVM 1", rows[0].Plan)
	assert.Equal(t, 649, rows[0].FirstPeriodPriceCents)
	assert.Equal(t, "hostingercom-vps-kvm1-usd-1y", rows[1].ItemID)
}

func TestHostingerTableOutput_HidesSSHKeyByDefault(t *testing.T) {
	stdout, _, err := runHostingerCommand(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":5,"name":"laptop","key":"ssh-ed25519 AAAAsecret"}]`))
	}, "", "hostinger", "ssh-keys", "token-1")

	require.NoError(t, err)
	assert.Contains(t, stdout, "laptop")
	assert.NotContains(t, stdout, "AAAAsecret")
}

func TestHostingerCreateCommand_RequiresFlags(t *testing.T) {
	var requests atomic.Int32
	_, _, err := runHostingerCommand(t, func(http.ResponseWriter, *http.Request) { requests.Add(1) }, "yes\n", "hostinger", "create", "--cloud-token", "token-1", "--force")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--item-id")
	assert.Equal(t, int32(0), requests.Load())
}

func TestHostingerCreateCommand_DeclinedConfirmationDoesNotPurchase(t *testing.T) {
	for _, answer := range []string{"no\n", "\n", ""} {
		var requests atomic.Int32
		stdout, stderr, err := runHostingerCommand(t, func(http.ResponseWriter, *http.Request) { requests.Add(1) }, answer, hostingerCreateArgs...)

		if answer == "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to read input")
		} else {
			require.NoError(t, err)
			assert.Contains(t, stderr, "Server creation cancelled.")
		}
		assert.Contains(t, stderr, "This purchases a Hostinger VPS (hostingercom-vps-kvm1-usd-1m)")
		assert.Empty(t, stdout)
		assert.Equal(t, int32(0), requests.Load(), "answer %q must not send a request", answer)
	}
}

func TestHostingerCreateCommand_ConfirmedCreatesServer(t *testing.T) {
	for _, tc := range []struct {
		name       string
		stdin      string
		extraArgs  []string
		wantPrompt bool
	}{
		{name: "confirmed with yes", stdin: "YES\n", wantPrompt: true},
		{name: "force skips prompt", extraArgs: []string{"--force"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			var body models.HostingerServerCreateRequest
			args := append(append([]string{}, hostingerCreateArgs...), tc.extraArgs...)
			args = append(args, "--format", "json", "--enable-backups=false", "--ssh-key-ids", "5,6")
			stdout, stderr, err := runHostingerCommand(t, func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/v1/servers/hostinger", r.URL.Path)
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"uuid":"server-1","hostinger_virtual_machine_id":42,"ip":"192.0.2.1","provisioning":true}`))
			}, tc.stdin, args...)

			require.NoError(t, err)
			assert.Equal(t, int32(1), requests.Load())
			assert.Equal(t, tc.wantPrompt, strings.Contains(stderr, "This purchases a Hostinger VPS"))
			assert.Equal(t, models.HostingerServerCreateRequest{
				CloudProviderTokenUUID: "token-1",
				ItemID:                 "hostingercom-vps-kvm1-usd-1m",
				DataCenterID:           9,
				TemplateID:             1002,
				PrivateKeyUUID:         "key-1",
				EnableBackups:          false,
				PublicKeyIDs:           []int{5, 6},
			}, body)
			assert.JSONEq(t, `{"uuid":"server-1","hostinger_virtual_machine_id":42,"ip":"192.0.2.1","provisioning":true}`, stdout)
		})
	}
}

func TestHostingerCreateCommand_PaymentPendingIsAnError(t *testing.T) {
	stdout, _, err := runHostingerCommand(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"message":"Hostinger is still processing the payment."}`))
	}, "", append(append([]string{}, hostingerCreateArgs...), "--force")...)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Hostinger is still processing the payment.")
	assert.Contains(t, err.Error(), "No server was added to Coolify")
	assert.Contains(t, err.Error(), "duplicate purchase")
	assert.Empty(t, stdout)
}

func TestHostingerCreateCommand_ServerErrorIsNotRetried(t *testing.T) {
	var requests atomic.Int32
	_, _, err := runHostingerCommand(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Failed to register server","hostinger_virtual_machine_id":42}`))
	}, "", append(append([]string{}, hostingerCreateArgs...), "--force")...)

	require.Error(t, err)
	assert.Equal(t, int32(1), requests.Load())
	assert.Contains(t, err.Error(), "failed to create Hostinger server")
	assert.Contains(t, err.Error(), "Hostinger virtual machine 42 may already have been purchased")
}

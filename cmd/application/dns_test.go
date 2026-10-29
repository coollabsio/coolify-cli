package application

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSCreate_ConflictPrintsReplaceCommand(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"DNS record already points to 198.51.100.7.","conflict":{"record_id":"rec-1","current":"198.51.100.7","proposed":"203.0.113.10"}}`))
	})

	_, err := runCmd(t, "app", "dns", "create", "app-uuid", "--hostname", "app.example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "coolify app dns replace app-uuid --hostname app.example.com --record-id rec-1 --current 198.51.100.7")
	require.Len(t, fake.recorded(), 1)
	assert.JSONEq(t, `{"hostname":"app.example.com"}`, fake.recorded()[0].Body)
}

func TestDNSReplace_ForceSendsConflictValues(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"record":{"uuid":"r","hostname":"app.example.com","type":"A","address":"203.0.113.10","owned":false,"integration_token_uuid":"tok"}}`))
	})

	stdout, err := runCmd(t, "app", "dns", "replace", "app-uuid", "--hostname", "app.example.com", "--record-id", "rec-1", "--current", "198.51.100.7", "--force")

	require.NoError(t, err)
	require.Len(t, fake.recorded(), 1)
	assert.Equal(t, "/api/v1/applications/app-uuid/dns/replace", fake.recorded()[0].RequestURI)
	assert.JSONEq(t, `{"hostname":"app.example.com","record_id":"rec-1","current":"198.51.100.7"}`, fake.recorded()[0].Body)
	assert.Contains(t, stdout, "203.0.113.10")
}

func TestDNSRelease_KeepRecordSkipsConfirmation(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"result":"kept","message":"The DNS record was kept."}`))
	})

	_, err := runCmd(t, "app", "dns", "release", "app-uuid", "old.example.com", "--keep-record")

	require.NoError(t, err)
	require.Len(t, fake.recorded(), 1)
	assert.Equal(t, "/api/v1/applications/app-uuid/dns/old.example.com?delete_record=false", fake.recorded()[0].RequestURI)
}

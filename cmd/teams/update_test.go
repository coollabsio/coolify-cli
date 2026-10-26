package teams

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCommand_SendsFallbackSetting(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "disable", args: []string{"--build-server-fallback=false"}, want: false},
		{name: "enable", args: []string{"--build-server-fallback"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"id":0,"name":"Root Team","personal_team":true,"is_build_server_fallback_enabled":` + strconv.FormatBool(tt.want) + `}`))
			})

			stdout, err := runCmd(t, append([]string{"teams", "update", "--format", "json"}, tt.args...)...)
			require.NoError(t, err)

			reqs := fake.recorded()
			require.Len(t, reqs, 1)
			assert.Equal(t, http.MethodPatch, reqs[0].Method)
			assert.Equal(t, "/api/v1/team", reqs[0].RequestURI)
			assert.JSONEq(t, `{"is_build_server_fallback_enabled":`+strconv.FormatBool(tt.want)+`}`, reqs[0].Body)

			var out map[string]any
			require.NoError(t, json.Unmarshal([]byte(stdout), &out))
			assert.Equal(t, tt.want, out["is_build_server_fallback_enabled"])
		})
	}
}

func TestUpdateCommand_TableOutput(t *testing.T) {
	newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":0,"name":"Root Team","is_build_server_fallback_enabled":false}`))
	})
	stdout, err := runCmd(t, "teams", "update", "--build-server-fallback=false")
	require.NoError(t, err)
	assert.Contains(t, stdout, "is_build_server_fallback_enabled")
	assert.Contains(t, stdout, "Root Team")
}

func TestUpdateCommand_RequiresFlag(t *testing.T) {
	fake := newFakeAPI(t, func(http.ResponseWriter, *http.Request) {})
	_, err := runCmd(t, "teams", "update")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--build-server-fallback is required")
	assert.Empty(t, fake.recorded())
}

func TestUpdateCommand_APIError(t *testing.T) {
	newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"This action is unauthorized."}`))
	})
	_, err := runCmd(t, "teams", "update", "--build-server-fallback=true")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

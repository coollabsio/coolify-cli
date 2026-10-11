package previews

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogsPreviewCommand_Flags(t *testing.T) {
	cmd := NewLogsPreviewCommand()

	assert.Equal(t, "logs", cmd.Name())
	lines := cmd.Flags().Lookup("lines")
	require.NotNil(t, lines)
	assert.Equal(t, "n", lines.Shorthand)
	assert.Equal(t, "100", lines.DefValue)
	follow := cmd.Flags().Lookup("follow")
	require.NotNil(t, follow)
	assert.Equal(t, "f", follow.Shorthand)
	assert.NotNil(t, cmd.Flags().Lookup("show-timestamps"))
	assert.NotNil(t, cmd.Flags().Lookup("service"))
}

func TestNewLogsPreviewCommand_ValidatesArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing pr id", args: []string{"app-uuid"}, wantErr: "<app_uuid> <pr_id>"},
		{name: "non-numeric pr id", args: []string{"app-uuid", "abc"}, wantErr: "invalid pr_id"},
		{name: "zero pr id", args: []string{"app-uuid", "0"}, wantErr: "invalid pr_id"},
		{name: "lines below -1", args: []string{"app-uuid", "12", "--lines", "-2"}, wantErr: "invalid --lines"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewLogsPreviewCommand()
			cmd.SetArgs(tt.args)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestFollowLogs_PrintsOnlyNewOutput(t *testing.T) {
	responses := []struct {
		logs string
		err  error
	}{
		{logs: "line1\n"},
		{logs: "line1\n"},              // unchanged: nothing printed
		{err: errors.New("transient")}, // poll error: ignored
		{logs: "line1\nline2\n"},       // appended: suffix printed
		{logs: "rotated\n"},            // rotated: full output printed
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	fetch := func(context.Context) (string, error) {
		// A tick can race the cancellation; extra polls return the last
		// (unchanged) output so they print nothing.
		r := responses[min(calls, len(responses)-1)]
		calls++
		if calls == len(responses) {
			cancel()
		}
		return r.logs, r.err
	}

	var out bytes.Buffer
	err := followLogs(ctx, &out, time.Millisecond, fetch)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, calls, len(responses))
	assert.Equal(t, "line1\nline2\nrotated\n\nStopping log follow...\n", out.String())
}

func TestFollowLogs_InitialFetchErrorIsReturned(t *testing.T) {
	fetch := func(context.Context) (string, error) {
		return "", errors.New("Preview not found.")
	}

	var out bytes.Buffer
	err := followLogs(context.Background(), &out, time.Millisecond, fetch)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get preview logs")
	assert.Contains(t, err.Error(), "Preview not found.")
	assert.Empty(t, out.String())
}

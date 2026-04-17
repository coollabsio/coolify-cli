package ssh

import (
	"strings"
	"testing"
)

func TestUploadShellCmd_AtomicWrite(t *testing.T) {
	got := uploadShellCmd("/usr/local/bin/coold", 0o755)

	for _, want := range []string{
		`mkdir -p "$(dirname "/usr/local/bin/coold")"`,
		`cat > "/usr/local/bin/coold".tmp.$$`,
		`chmod 755 "/usr/local/bin/coold".tmp.$$`,
		`mv -f "/usr/local/bin/coold".tmp.$$ "/usr/local/bin/coold"`,
		`umask 077`,
		`set -e`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("upload cmd missing %q:\nGOT: %s", want, got)
		}
	}
}

func TestUploadShellCmd_ModeIsOctal(t *testing.T) {
	got := uploadShellCmd("/x", 0o644)
	if !strings.Contains(got, "chmod 644") {
		t.Errorf("expected octal mode 644, got: %s", got)
	}
}

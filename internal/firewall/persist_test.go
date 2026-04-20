package firewall

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowPersistUnit_ContainsExpected(t *testing.T) {
	u := AllowPersistUnit()
	assert.Contains(t, u, "After=coolify-mesh-fw.service")
	assert.Contains(t, u, "iptables-restore --noflush "+RulesPath)
	assert.Contains(t, u, "WantedBy=multi-user.target")
	assert.Contains(t, u, "Type=oneshot")
	assert.Contains(t, u, "RemainAfterExit=yes")
}

func TestSaveRulesCommand(t *testing.T) {
	c := SaveRulesCommand()
	assert.Contains(t, c, "mkdir -p "+RulesDir)
	assert.Contains(t, c, "iptables -S "+ChainName)
	assert.Contains(t, c, RulesPath+".tmp")
	assert.Contains(t, c, "mv "+RulesPath+".tmp "+RulesPath)
	// Rough shape check: single shell line.
	assert.False(t, strings.Contains(c, "\n"))
}

func TestInstallPersistenceCommand(t *testing.T) {
	c := InstallPersistenceCommand()
	assert.Contains(t, c, PersistUnitPath+".tmp")
	assert.Contains(t, c, "systemctl daemon-reload")
	assert.Contains(t, c, "systemctl enable "+PersistUnitName)
	assert.Contains(t, c, "systemctl start "+PersistUnitName)
}

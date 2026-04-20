package firewall

// Paths and unit names kept in one place so the install and restore flows
// stay in sync.
const (
	RulesPath       = "/etc/coolify/allow.rules"
	RulesDir        = "/etc/coolify"
	PersistUnitPath = "/etc/systemd/system/coolify-mesh-allow.service"
	PersistUnitName = "coolify-mesh-allow.service"
)

// AllowPersistUnit returns the systemd unit text that restores the saved
// COOLIFY-ALLOW rules on boot. Ordered after coolify-mesh-fw.service so
// the chain exists before the restore runs. --noflush keeps tables intact.
func AllowPersistUnit() string {
	return `[Unit]
Description=Coolify mesh COOLIFY-ALLOW restore
After=coolify-mesh-fw.service network-online.target
Wants=coolify-mesh-fw.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/sh -c 'test -f ` + RulesPath + ` && /usr/sbin/iptables-restore --noflush ` + RulesPath + ` || true'

[Install]
WantedBy=multi-user.target
`
}

// SaveRulesCommand returns a shell one-liner that captures the current
// COOLIFY-ALLOW chain (only) and atomically writes it to RulesPath in
// iptables-restore format. Safe to chain after an `iptables -A/-D` call.
//
// We emit a minimal `*filter` + `:COOLIFY-ALLOW -` + `-A COOLIFY-ALLOW ...`
// + `COMMIT` block so `iptables-restore --noflush` only touches our chain.
func SaveRulesCommand() string {
	return `mkdir -p ` + RulesDir + ` && ` +
		`( printf '*filter\n:` + ChainName + ` -\n'; ` +
		`iptables -S ` + ChainName + ` 2>/dev/null | grep '^-A ' || true; ` +
		`printf 'COMMIT\n' ) > ` + RulesPath + `.tmp && ` +
		`mv ` + RulesPath + `.tmp ` + RulesPath
}

// InstallPersistenceCommand returns a shell command that writes the
// coolify-mesh-allow.service unit, reloads systemd, enables/starts it.
// Idempotent: atomic rewrite, enable is a no-op when already enabled.
func InstallPersistenceCommand() string {
	return `cat > ` + PersistUnitPath + `.tmp <<'COOLIFY_ALLOW_UNIT_EOF'
` + AllowPersistUnit() + `COOLIFY_ALLOW_UNIT_EOF
mv ` + PersistUnitPath + `.tmp ` + PersistUnitPath + ` && ` +
		`systemctl daemon-reload && ` +
		`systemctl enable ` + PersistUnitName + ` && ` +
		`systemctl start ` + PersistUnitName
}

package services

// RedisInstallCommand returns a shell snippet that installs Redis via apt and
// enables the service. Works on both Debian (redis-server unit) and Ubuntu
// (same, but service name may be redis or redis-server).
func RedisInstallCommand() string {
	return `DEBIAN_FRONTEND=noninteractive apt-get update -qq 2>/dev/null && ` +
		`DEBIAN_FRONTEND=noninteractive apt-get install -y ` +
		`-o Dpkg::Options::="--force-confold" ` +
		`redis-server 2>&1 && ` +
		`(systemctl enable --now redis-server 2>&1 || systemctl enable --now redis 2>&1)`
}

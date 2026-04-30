package initcmd

import (
	"fmt"
	"net"

	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// buildDesired turns the flag struct into a wireguard.DesiredMesh. Intent is
// pulled from flags.Intent so each subcommand can set it before calling the
// shared plan/apply pipeline.
func buildDesired(flags *InitFlags) (*wireguard.DesiredMesh, error) {
	_, mgmtPool, err := net.ParseCIDR(flags.WGMgmtPool)
	if err != nil {
		return nil, fmt.Errorf("invalid --wg-mgmt-pool %q: %w", flags.WGMgmtPool, err)
	}
	_, contPool, err := net.ParseCIDR(flags.ContainerPool)
	if err != nil {
		return nil, fmt.Errorf("invalid --container-pool %q: %w", flags.ContainerPool, err)
	}

	return &wireguard.DesiredMesh{
		Hosts:                 flags.Servers,
		Interface:             flags.WGInterface,
		MgmtPool:              mgmtPool,
		ContainerPool:         contPool,
		ContainerPrefix:       flags.ContainerPrefix,
		ListenPort:            flags.WGListenPort,
		InstallPodman:         true,
		Namespaces:            flags.Namespaces,
		DefaultDenyContainers: !flags.SkipDefaultDeny,
		InstallCoold:          true,
		CooldVersion:          flags.CooldVersion,
		CorrosionVersion:      flags.CorrosionVersion,
		CorrosionGossipPort:   flags.CorrosionGossipPort,
		CorrosionAPIPort:      flags.CorrosionAPIPort,
		CentralHost:           flags.CentralHost,
		SchedulerVersion:      flags.SchedulerVersion,
		EnableBuilder:         flags.EnableBuilder,
		BuilderHosts:          flags.BuilderHosts,
		BuilderCapacity:       flags.BuilderCapacity,
		BuilderCPUQuota:       flags.BuilderCPUQuota,
		BuilderMemoryMax:      flags.BuilderMemoryMax,
		BuilderTimeoutSecs:    flags.BuilderTimeoutSecs,
		Intent:                wireguard.Intent(flags.Intent),
		NewHosts:              flags.NewHosts,
		AllowReplace:          flags.AllowReplace,
		AllowNightly:          flags.AllowNightly,
	}, nil
}

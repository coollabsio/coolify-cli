package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloudTokenDecodesPrivilegedSensitiveToken(t *testing.T) {
	var token CloudToken
	require.NoError(t, json.Unmarshal([]byte(`{"uuid":"token-uuid","name":"prod","provider":"vultr","token":"secret"}`), &token))
	require.NotNil(t, token.Token)
	require.Equal(t, "secret", *token.Token)
}

func TestCloudProviderValid_AcceptsHostinger(t *testing.T) {
	for _, provider := range []CloudProvider{CloudProviderHetzner, CloudProviderDigitalOcean, CloudProviderVultr, CloudProviderHostinger} {
		require.True(t, provider.Valid(), provider)
	}
	require.False(t, CloudProvider("linode").Valid())

	var token CloudToken
	require.NoError(t, json.Unmarshal([]byte(`{"uuid":"token-uuid","name":"prod","provider":"hostinger"}`), &token))
	require.Equal(t, CloudProviderHostinger, token.Provider)
}

func TestFlattenHostingerCatalog_OneRowPerPrice(t *testing.T) {
	items := []HostingerCatalogItem{
		{
			ID:       "hostingercom-vps-kvm1",
			Name:     "KVM 1",
			Category: "VPS",
			Metadata: HostingerCatalogMetadata{CPUs: "1", Memory: "4096", DiskSpace: "51200", Bandwidth: "4096000", Network: "300"},
			Prices: []HostingerCatalogPrice{
				{ID: "hostingercom-vps-kvm1-usd-1m", Currency: "USD", Price: 1399, FirstPeriodPrice: 649, Period: 1, PeriodUnit: "month"},
				{ID: "hostingercom-vps-kvm1-usd-1y", Currency: "USD", Price: 11988, FirstPeriodPrice: 7788, Period: 1, PeriodUnit: "year"},
			},
		},
		{ID: "hostingercom-vps-kvm2", Name: "KVM 2"},
	}

	rows := FlattenHostingerCatalog(items)

	require.Len(t, rows, 2)
	require.Equal(t, HostingerPlanPrice{
		ItemID: "hostingercom-vps-kvm1-usd-1m", Plan: "KVM 1", PlanID: "hostingercom-vps-kvm1", Category: "VPS",
		CPUs: "1", Memory: "4096", DiskSpace: "51200", Bandwidth: "4096000", Network: "300",
		Currency: "USD", PriceCents: 1399, FirstPeriodPriceCents: 649, Period: 1, PeriodUnit: "month",
	}, rows[0])
	require.Equal(t, "year", rows[1].PeriodUnit)
	require.NotNil(t, FlattenHostingerCatalog(nil))
}

func TestHostingerServerCreateResponse_PaymentPending(t *testing.T) {
	require.True(t, HostingerServerCreateResponse{Message: "Payment is being processed"}.PaymentPending())
	require.False(t, HostingerServerCreateResponse{UUID: "server-1"}.PaymentPending())
}

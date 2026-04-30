package models

// PlanActionRow is a table-friendly row for the plan output.
type PlanActionRow struct {
	Server string `json:"server"`
	Action string `json:"action"`
	Detail string `json:"detail"`
}

// PlanSkippedRow is a table-friendly row for actions the intent filter
// suppressed (shown in the plan preview so operators can see what would have
// run and why).
type PlanSkippedRow struct {
	Server string `json:"server"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

// ApplyResultRow is a table-friendly row for the apply result output.
type ApplyResultRow struct {
	Server string `json:"server"`
	Action string `json:"action"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// VerifyResultRow is a table-friendly row for post-apply verification.
type VerifyResultRow struct {
	Server      string `json:"server"`
	WireGuardIP string `json:"wireguard_ip"`
	PeerCount   int    `json:"peer_count"`
	Status      string `json:"status"`
}

// PlanOutput is the structured JSON output for the plan command.
type PlanOutput struct {
	Servers  []string         `json:"servers"`
	Intent   string           `json:"intent,omitempty"`
	Actions  []PlanActionRow  `json:"actions"`
	Skipped  []PlanSkippedRow `json:"skipped,omitempty"`
	Warnings []string         `json:"warnings,omitempty"`
}

// ApplyOutput is the structured JSON output for the apply command.
type ApplyOutput struct {
	Results  []ApplyResultRow  `json:"results"`
	Verified []VerifyResultRow `json:"verified"`
}

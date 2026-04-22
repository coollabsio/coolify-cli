package models

// ContainerRow is a table-friendly row for `coolify firewall containers`.
type ContainerRow struct {
	Host      string `json:"host"`
	Namespace string `json:"namespace"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
}

// AllowRuleRow is a table-friendly row for `coolify firewall list`.
type AllowRuleRow struct {
	Host      string `json:"host"`
	Namespace string `json:"namespace"`
	ID        string `json:"id"`
	Src       string `json:"src"`
	Dst       string `json:"dst"`
	Proto     string `json:"proto,omitempty"`
	Port      int    `json:"port,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// FirewallContainersOutput is the JSON output for `firewall containers`.
type FirewallContainersOutput struct {
	Containers []ContainerRow `json:"containers"`
	Errors     []string       `json:"errors,omitempty"`
}

// FirewallListOutput is the JSON output for `firewall list`.
type FirewallListOutput struct {
	Rules  []AllowRuleRow `json:"rules"`
	Errors []string       `json:"errors,omitempty"`
}

// FirewallAllowOutput is the JSON output for `firewall allow` / `revoke`.
type FirewallAllowOutput struct {
	Rules []AllowRuleRow `json:"rules"`
}

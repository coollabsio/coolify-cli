package models

// Domain represents a domain configuration
type Domain struct {
	IP      string   `json:"ip"`
	Domains []string `json:"domains"`
}

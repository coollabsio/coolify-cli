package coolTypes

var Redacted = "********"

type Instance struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
	Fqdn    string `json:"fqdn"`
	Token   string `json:"token"`
}

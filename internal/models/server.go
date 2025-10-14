package models

// Server represents a Coolify server
type Server struct {
	ID       int      `json:"id"`
	UUID     string   `json:"uuid"`
	Name     string   `json:"name"`
	IP       string   `json:"ip" sensitive:"true"`
	User     string   `json:"user" sensitive:"true"`
	Port     int      `json:"port" sensitive:"true"`
	Settings Settings `json:"settings" table:"-"`
}

// Settings for server
type Settings struct {
	IsReachable bool `json:"is_reachable"`
	IsUsable    bool `json:"is_usable"`
}

// ServerCreateRequest for creating servers
type ServerCreateRequest struct {
	Name            string `json:"name"`
	IP              string `json:"ip"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	PrivateKeyUUID  string `json:"private_key_uuid"`
	InstantValidate bool   `json:"instant_validate"`
}

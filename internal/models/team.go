package models

// Team represents a Coolify team
type Team struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Description       *string `json:"description,omitempty"`
	PersonalTeam      bool    `json:"personal_team"`
	CreatedAt         string  `json:"created_at" table:"-"`
	UpdatedAt         string  `json:"updated_at" table:"-"`
	ShowBoarding      bool    `json:"show_boarding" `
	CustomServerLimit *string `json:"custom_server_limit,omitempty" table:"-"`
}

// TeamMember represents a member of a team
type TeamMember struct {
	ID                   int     `json:"id"`
	Name                 string  `json:"name"`
	Email                string  `json:"email" sensitive:"true"`
	EmailVerifiedAt      *string `json:"email_verified_at,omitempty" table:"-"`
	Role                 *string `json:"role,omitempty"`
	CreatedAt            string  `json:"created_at" table:"-"`
	UpdatedAt            string  `json:"updated_at" table:"-"`
	TwoFactorConfirmedAt *string `json:"two_factor_confirmed_at,omitempty" table:"-"`
	ForcePasswordReset   bool    `json:"force_password_reset"`
	MarketingEmails      bool    `json:"marketing_emails"`
}

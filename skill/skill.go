package skill

type Skill struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    RawURL      string `json:"raw_url"`
    InstalledAt string `json:"installed_at"`
    UpdatedAt   string `json:"updated_at"`
}

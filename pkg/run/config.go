package run

type RootConfig struct {
	Path     string `json:"path"`
	ReadOnly bool   `json:"readonly"`
}

type ImageConfig struct {
	OciVersion    string         `json:"ociVersion"`
	ProcessConfig ProcessConfig  `json:"process"`
	Hostname      string         `json:"hostname"`
	MountsConfig  []MountsConfig `json:"mounts"`
	Root          RootConfig     `json:"root"`
}

type ProcessConfig struct {
	Terminal bool           `json:"terminal"`
	User     map[string]int `json:"user"`
	Args     []string       `json:"args"`
	Env      []string       `json:"env"`
	Cwd      string         `json:"cwd"`
}

type MountsConfig struct {
	Destination string   `json:"destination"`
	Source      string   `json:"source"`
	Type        string   `json:"type"`
	Options     []string `json:"options,omitempty"`
}

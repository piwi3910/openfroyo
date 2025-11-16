package parser

// Inventory represents the complete inventory configuration
type Inventory struct {
	Hosts  map[string]Host  `yaml:"hosts"`
	Groups map[string]Group `yaml:"groups"`
}

// Host represents a single host configuration
type Host struct {
	SSHHost     string            `yaml:"ssh_host"`
	SSHPort     int               `yaml:"ssh_port"`
	SSHUser     string            `yaml:"ssh_user"`
	SSHPassword string            `yaml:"ssh_password,omitempty"`
	SSHKeyFile  string            `yaml:"ssh_key_file,omitempty"`
	Vars        map[string]string `yaml:"vars,omitempty"`
}

// Group represents a group of hosts
type Group struct {
	Hosts []string          `yaml:"hosts"`
	Vars  map[string]string `yaml:"vars,omitempty"`
}

// Stack represents a stack configuration
type Stack struct {
	Name      string                 `yaml:"name"`
	Inventory []string               `yaml:"inventory"`
	Defaults  map[string]interface{} `yaml:"defaults,omitempty"`
	Run       []RunEntry             `yaml:"run"`
}

// RunEntry represents either a module invocation or a task block
type RunEntry struct {
	Name     string                 `yaml:"name"`
	Module   string                 `yaml:"module,omitempty"`
	Hosts    []string               `yaml:"hosts,omitempty"`
	Vars     map[string]interface{} `yaml:"vars,omitempty"`
	Strategy string                 `yaml:"strategy,omitempty"` // "serial" or "parallel"
}

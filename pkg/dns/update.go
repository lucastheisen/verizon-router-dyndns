package dns

type DNSUpdate struct {
	Domains map[string]DNSUpdateDomain `json:"domains" yaml:"domains"`
}

type DNSUpdateDomain struct {
	Name  string   `json:"name" yaml:"name"`
	Hosts []string `json:"hosts" yaml:"hosts"`
}

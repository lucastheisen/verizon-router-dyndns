package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/lucastheisen/verizon-router-dyndns/pkg/dns"
)

type DNSUpdateFlag dns.DNSUpdate

func (e *DNSUpdateFlag) String() string {
	data, _ := json.Marshal(e)
	return string(data)
}

func (r *DNSUpdateFlag) Set(v string) error {
	err := json.Unmarshal([]byte(v), r)
	if err != nil {
		return fmt.Errorf("unmarshal dnsrecord: %w", err)
	}
	return nil
}

func (*DNSUpdateFlag) Type() string {
	return "DNSUpdate"
}

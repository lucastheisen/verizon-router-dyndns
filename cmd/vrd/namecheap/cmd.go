package namecheap

import (
	"errors"
	"fmt"

	"github.com/lucastheisen/verizon-router-dyndns/pkg/cmd"
	"github.com/lucastheisen/verizon-router-dyndns/pkg/verizon"
	"github.com/namecheap/go-namecheap-sdk/v2/namecheap"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	namecheapOptions := &namecheap.ClientOptions{}
	var networkName string
	router := &verizon.Router{}
	updates := cmd.DNSUpdateFlag{}

	cmd := cobra.Command{
		Use:   "namecheap",
		Short: `Updates namecheap with DNS with IP addresses obtained by querying a verizon router.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(updates.Domains) == 0 {
				return errors.New("at least one dns update required")
			}

			router.InsecureSkipVerify = true

			err := router.Connect()
			if err != nil {
				return fmt.Errorf("connect: %w", err)
			}

			networks, err := router.Network()
			if err != nil {
				return fmt.Errorf("network: %w", err)
			}

			var ip *string
			for _, network := range networks {
				Logger.Trace().Interface("network", network).Msg("networks")
				if network.Name == networkName {
					ip = &network.IPAddress
					break
				}
			}

			if ip == nil {
				return fmt.Errorf("network %s not found", networkName)
			} else if *ip == "" {
				return fmt.Errorf("network %s does not have IP address", networkName)
			}

			ncclient := namecheap.NewClient(namecheapOptions)
			for _, update := range updates.Domains {
				records := make([]namecheap.DomainsDNSHostRecord, len(update.Hosts))
				for i, host := range update.Hosts {
					records[i] = namecheap.DomainsDNSHostRecord{
						Address:    ip,
						HostName:   &host,
						RecordType: namecheap.String("A"),
					}
				}

				ncclient.DomainsDNS.SetHosts(
					&namecheap.DomainsDNSSetHostsArgs{
						Domain:  namecheap.String(update.Name),
						Records: &records,
					})
			}

			return nil
		},
	}

	// CODE_REVIEW_CATCH_ME add netrc/netrc-file support
	cmd.Flags().Var(&updates, "updates", "DNS records to update from router IP")
	cmd.Flags().StringVar(&namecheapOptions.ApiKey, "namecheap-api-key", "", "API key for namecheap")
	cmd.Flags().StringVar(&namecheapOptions.ApiUser, "namecheap-api-user", "", "API username for namecheap")
	cmd.Flags().StringVar(&router.Password, "router-password", "", "admin password for router")
	cmd.Flags().StringVar(
		&networkName,
		"network-name",
		"Broadband Connection (Ethernet/Coax)",
		"router network name for external network used as source of external IP address")

	return &cmd
}

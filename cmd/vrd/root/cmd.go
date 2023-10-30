package root

import (
	"fmt"
	"io"
	"os"

	"github.com/lucastheisen/verizon-router-dyndns/cmd/vrd/namecheap"
	"github.com/lucastheisen/verizon-router-dyndns/pkg/log"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var logLevel string
	var logFormat string

	cmd := cobra.Command{
		Use: "vrd",
		Short: `
A tool for updating various dns providers with dynamically detected IP addresses
obtained by querying a verizon router.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			lvl, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				return fmt.Errorf("parse level: %w", err)
			}

			var writer io.Writer
			switch logFormat {
			case "pretty":
				writer = zerolog.ConsoleWriter{Out: os.Stderr}
			default:
				writer = os.Stderr
			}

			log.Root.Configure(
				writer,
				func(name string, lgr zerolog.Logger) zerolog.Logger {
					return lgr.With().Timestamp().Logger().Level(lvl)
				})
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log", "info", "root log level")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", "json", "log format")

	cmd.AddCommand(namecheap.New())

	return &cmd
}

package httpclient

import (
	"github.com/lucastheisen/verizon-router-dyndns/pkg/log"
	"github.com/rs/zerolog"
)

var Logger = log.Root.NewLogger(
	"github.com/pastdev/verizon-router-dyndns/pkg/httpclient",
	func(name string, lgr zerolog.Logger) zerolog.Logger {
		return lgr.With().Str("logger", name).Logger()
	})

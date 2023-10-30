package log_test

import (
	"strings"
	"testing"

	"github.com/lucastheisen/verizon-router-dyndns/pkg/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("child logger can be created without root config", func(t *testing.T) {
		lgr := log.Root.NewLogger("child")
		require.NotNil(t, lgr)
	})

	t.Run("root logger logs", func(t *testing.T) {
		var buffer strings.Builder
		log.Root.Configure(
			&buffer,
			func(name string, lgr zerolog.Logger) zerolog.Logger {
				return lgr.Level(zerolog.InfoLevel)
			})
		log.Root.Debug().Msg("debug")
		log.Root.Info().Msg("info")
		require.Equal(
			t,
			`{"level":"info","message":"info"}
`,
			buffer.String())
	})

	t.Run("child logger logs at root level", func(t *testing.T) {
		var buffer strings.Builder
		log.Root.Configure(
			&buffer,
			func(name string, lgr zerolog.Logger) zerolog.Logger {
				return lgr.Level(zerolog.InfoLevel)
			})
		lgr := log.Root.NewLogger("child")

		lgr.Debug().Msg("debug")
		lgr.Info().Msg("info")
		require.Equal(
			t,
			`{"level":"info","message":"info"}
`,
			buffer.String())
	})

	t.Run(
		"child logger log level changes upon root level change",
		func(t *testing.T) {
			var buffer strings.Builder
			log.Root.Configure(&buffer)
			lgr := log.Root.NewLogger(
				"child",
				func(name string, lgr zerolog.Logger) zerolog.Logger {
					return lgr.With().Str("logger", name).Logger()
				})

			require.Equal(t, zerolog.TraceLevel, lgr.GetLevel())

			lgr.Debug().Msg("before")
			lgr.Info().Msg("before")

			log.Root.Configure(
				&buffer,
				func(name string, lgr zerolog.Logger) zerolog.Logger {
					return lgr.Level(zerolog.InfoLevel)
				})

			require.Equal(t, zerolog.InfoLevel, log.Root.GetLevel())
			require.Equal(t, zerolog.InfoLevel, lgr.GetLevel())

			lgr.Debug().Msg("after")
			lgr.Info().Msg("after")
			require.Equal(
				t,
				`{"level":"debug","logger":"child","message":"before"}
{"level":"info","logger":"child","message":"before"}
{"level":"info","logger":"child","message":"after"}
`,
				buffer.String())
		})
}

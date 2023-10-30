package log

import (
	"io"
	"sync/atomic"

	"github.com/rs/zerolog"
)

var Root = defaultRoot()

type Configurer func(name string, lgr zerolog.Logger) zerolog.Logger

type Logger struct {
	name       string
	p          atomic.Pointer[zerolog.Logger]
	parent     zerolog.Logger
	configurer Configurer
	children   map[string]*Logger
}

type RootLogger struct {
	Logger
}

func (l *Logger) Configure(configurer Configurer) {
	l.configurer = configurer
	lgr := l.configurer(l.name, l.parent)
	l.Store(&lgr)

	for _, child := range l.children {
		child.parent = lgr
		child.Configure(child.configurer)
	}
}

func (l *Logger) Debug() *zerolog.Event {
	return l.Load().Debug()
}

func (l *Logger) Error() *zerolog.Event {
	return l.Load().Error()
}

func (l *Logger) Fatal() *zerolog.Event {
	return l.Load().Fatal()
}

func (l *Logger) GetLevel() zerolog.Level {
	return l.Load().GetLevel()
}

func (l *Logger) Info() *zerolog.Event {
	return l.Load().Info()
}

func (l *Logger) Level(level zerolog.Level) *Logger {
	lgr := l.Load().Level(level)
	return (&Logger{}).Store(&lgr)
}

func (l *Logger) Load() *zerolog.Logger {
	return l.p.Load()
}

func (l *Logger) NewLogger(name string, configurer ...Configurer) *Logger {
	if len(l.children) == 0 {
		l.children = map[string]*Logger{}
	}

	configr := func(name string, lgr zerolog.Logger) zerolog.Logger {
		for _, c := range configurer {
			lgr = c(name, lgr)
		}
		return lgr
	}

	parent := l.Load()
	child := &Logger{
		configurer: configr,
		name:       name,
		parent:     *parent,
	}

	lgr := *parent
	if configurer != nil {
		lgr = configr(name, *parent)
	}
	child.Store(&lgr)

	l.children[name] = child

	return child
}

func (l *Logger) Panic() *zerolog.Event {
	return l.Load().Panic()
}

func (l *Logger) Store(logger *zerolog.Logger) *Logger {
	l.p.Store(logger)
	return l
}

func (l *Logger) Trace() *zerolog.Event {
	return l.Load().Trace()
}

func (l *Logger) Warn() *zerolog.Event {
	return l.Load().Warn()
}

func (l *RootLogger) Configure(w io.Writer, configurer ...Configurer) {
	if w == nil {
		l.parent = zerolog.Nop()
	} else {
		l.parent = zerolog.New(w)
	}

	l.Logger.Configure(func(name string, lgr zerolog.Logger) zerolog.Logger {
		for _, c := range configurer {
			lgr = c(name, lgr)
		}
		return lgr
	})
}

func defaultRoot() *RootLogger {
	l := &RootLogger{
		Logger{
			name: "root",
		},
	}

	l.Configure(nil)

	return l
}

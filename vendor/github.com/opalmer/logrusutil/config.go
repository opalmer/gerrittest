package logrusutil

const (
	// DisabledLevel may be passed into a *Config struct
	// to disable logging and discard output.
	DisabledLevel = "disabled"
)

var (
	// DefaultLevel is the default log level used in NewConfig()
	DefaultLevel = "warning"

	// DefaultHookLevel is the default level where hooks will
	// contribute log messages.
	DefaultHookLevel = "debug"

	// DefaultHookStackLevel is the level at which the hook will
	// apply.
	DefaultHookStackLevel = 4

	// DefaultCallerHookField sets the default field name that the
	// CallerHook hook will apply to a log entry.
	DefaultCallerHookField = "src"
)

// Config is used to provided configuration information
type Config struct {
	Level           string
	HookLevel       string
	HookStackLevel  int
	CallerHookField string
}

// NewConfig produces a *Config struct with reasonable defaults.
func NewConfig() *Config {
	return &Config{
		Level:           DefaultLevel,
		HookLevel:       DefaultHookLevel,
		HookStackLevel:  DefaultHookStackLevel,
		CallerHookField: DefaultCallerHookField,
	}
}

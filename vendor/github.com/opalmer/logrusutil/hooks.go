package logrusutil

import (
	"github.com/go-stack/stack"
	"github.com/sirupsen/logrus"
)

// CallerHook is a logrus hook which applies information
// about the caller to log messages. This is an implementation
// of the following interface:
//  https://github.com/sirupsen/logrus/blob/v1.0.3/hooks.go#L8
type CallerHook struct {
	field    string
	depth    int
	disabled bool
	levels   []logrus.Level
}

// Levels returns the level(s) which this hook applies to.
func (c *CallerHook) Levels() []logrus.Level {
	return c.levels
}

// Fire will run the hook and apply it to the logrus entry.
func (c *CallerHook) Fire(entry *logrus.Entry) error {
	if !c.disabled && entry.Data != nil {
		entry.Data[c.field] = stack.Caller(c.depth).String()
	}
	return nil
}

// NewCallerHook constructs and returns logrus.Hook implementation
// based on the provided configuration.
func NewCallerHook(disabled bool, depth int, field string, level logrus.Level) logrus.Hook {
	levels := []logrus.Level{}
	if !disabled {
		for _, lvl := range logrus.AllLevels {
			if lvl <= level {
				levels = append(levels, lvl)
			}
		}
	}

	return &CallerHook{
		field:    field,
		depth:    depth,
		disabled: disabled,
		levels:   levels,
	}
}

package logger

import (
	"fmt"
)

// FakeLogger is a test implementation of InterfaceLogger that stores log messages
// instead of outputting them. Used for testing handlers and other components.
type FakeLogger struct {
	Infos  []string
	Errors []string
	Warns  []string
	Debugs []string
	Fatals []string // in tests you store, not exit
}

func (l *FakeLogger) Info(args ...interface{}) {
	l.Infos = append(l.Infos, fmt.Sprint(args...))
}

func (l *FakeLogger) Infof(t string, args ...interface{}) {
	l.Infos = append(l.Infos, fmt.Sprintf(t, args...))
}

func (l *FakeLogger) Error(args ...interface{}) {
	l.Errors = append(l.Errors, fmt.Sprint(args...))
}

func (l *FakeLogger) Errorf(t string, args ...interface{}) {
	l.Errors = append(l.Errors, fmt.Sprintf(t, args...))
}

func (l *FakeLogger) Warn(args ...interface{}) {
	l.Warns = append(l.Warns, fmt.Sprint(args...))
}

func (l *FakeLogger) Warnf(t string, args ...interface{}) {
	l.Warns = append(l.Warns, fmt.Sprintf(t, args...))
}

func (l *FakeLogger) Debug(args ...interface{}) {
	l.Debugs = append(l.Debugs, fmt.Sprint(args...))
}

func (l *FakeLogger) Debugf(t string, args ...interface{}) {
	l.Debugs = append(l.Debugs, fmt.Sprintf(t, args...))
}

// Important: Never call os.Exit in tests!
func (l *FakeLogger) Fatal(args ...interface{}) {
	l.Fatals = append(l.Fatals, fmt.Sprint(args...))
}

func (l *FakeLogger) Fatalf(t string, args ...interface{}) {
	l.Fatals = append(l.Fatals, fmt.Sprintf(t, args...))
}

func (l *FakeLogger) Sync() error {
	return nil
}

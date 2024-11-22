package bfflog

import (
	"fmt"
	"log/slog"
)

func Err(e error) slog.Attr {
	return slog.String("error", e.Error())
}

func Component(c string) slog.Attr {
	return slog.String("component", c)
}

func ActorDID(did string) slog.Attr {
	return slog.String("actor_did", did)
}

func ChildLogger(parent *slog.Logger, component string) *slog.Logger {
	return parent.With(slog.String("component", component))
}

type PyroscopeSlogAdapter struct {
	Slog *slog.Logger
}

func (p *PyroscopeSlogAdapter) Infof(a string, b ...interface{}) {
	p.Slog.Info(fmt.Sprintf(a, b...))
}

func (p *PyroscopeSlogAdapter) Debugf(a string, b ...interface{}) {
	p.Slog.Debug(fmt.Sprintf(a, b...))
}

func (p *PyroscopeSlogAdapter) Errorf(a string, b ...interface{}) {
	p.Slog.Error(fmt.Sprintf(a, b...))
}

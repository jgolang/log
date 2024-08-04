package log

import (
	"fmt"

	"github.com/jgolang/errors"

	"testing"
)

func TestDebug(t *testing.T) {
	type args struct {
		msg  string
		args []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Debug",
			args: args{
				msg:  "HELLO",
				args: []interface{}{"Debug message"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Debug(tt.args.args...)
		})
	}
}

// func TestDebugf(t *testing.T) {
// 	type args struct {
// 		template string
// 		args     []interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		{
// 			name: "Debugf",
// 			args: args{
// 				template: "Debug message, %v",
// 				args:     []interface{}{"\n\n\nthis complement"},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			Debugf(tt.args.template, tt.args.args...)
// 		})
// 	}
// }

// func TestWarn(t *testing.T) {
// 	type args struct {
// 		args []interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		{
// 			name: "Warn",
// 			args: args{
// 				args: []interface{}{"This is warn message"},
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			Warn(tt.args.args...)
// 		})
// 	}
// }

func TestDebugTesting(t *testing.T) {
	Debug("hello")
	err := a()
	Debug("error", "error", err)
	Warn("hello")
	Info("hello", "error", err)
	err2 := c()
	Info("hello", "error", err2)
	Error("hello", "error", err2)
	Error(err)
}

func a() error {
	return b()
}
func b() error {
	return errors.New("custom error")
}

func c() error {
	return d()
}
func d() error {
	return fmt.Errorf("normal error")
}

// buildInfo, _ := debug.ReadBuildInfo()

// slog.Group(
// 	"info",
// 	slog.Int("pid", os.Getpid()),
// 	slog.String("go_version", buildInfo.GoVersion),
// )

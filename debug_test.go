package log

import (
	"testing"
)

func TestDebug(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Debug",
			args: args{
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

func TestDebugf(t *testing.T) {
	type args struct {
		template string
		args     []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Debugf",
			args: args{
				template: "Debug message, %v",
				args:     []interface{}{"\n\n\nthis complement"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Debugf(tt.args.template, tt.args.args...)
		})
	}
}

func TestWarn(t *testing.T) {
	type args struct {
		args []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Warn",
			args: args{
				args: []interface{}{"This is warn message"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Warn(tt.args.args...)
		})
	}
}

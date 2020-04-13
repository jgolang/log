package log

import (
	"testing"
)

func Test_log(t *testing.T) {

	prod = true

	type args struct {
		skip     int
		p        priority
		template string
		args     []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ERROR",
			args: args{
				skip:     2,
				p:        errorPriority,
				template: "Error message %v:%v",
				args:     []interface{}{1},
			},
		},
		{
			name: "INFO",
			args: args{
				skip:     2,
				p:        infoPriority,
				template: "Info message %v:%v",
				args:     []interface{}{1, 2},
			},
		},
		{
			name: "FATAL",
			args: args{
				skip:     2,
				p:        fatalPriority,
				template: "Fatal message %v:%v",
				args:     []interface{}{1, 2},
			},
		},
		{
			name: "PANIC",
			args: args{
				skip:     2,
				p:        panicPriority,
				template: "Panic message %v:%v",
				args:     []interface{}{1, 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log(tt.args.skip, tt.args.p, tt.args.template, tt.args.args)
		})
	}

}

package log

import (
	"os"
)

var std = New(
	Formater{
		prodFlag: LstdProdFlags,
		devFlag:  LstdDevFlags,
	},
	Out{
		wr: os.Stderr,
	},
	3)

func init() {
	// Load enviroment mode
	switch mode := os.Getenv("MODE"); mode {
	case "PROD":
		ProductionMode()
	case "DEV":
		fallthrough
	default:
		DevelopmentMode()
	}
}

// SetNewFormat configure your custom outputs development and production format for default loggin
func SetNewFormat(f formater) {
	std.SetNewFormat(f)
}

// SetNewOutput Set custom log output destination for default loggin
func SetNewOutput(o out) {
	std.SetNewOutput(o)
}

// ProductionMode set production mode logger for default loggin
func ProductionMode() {
	std.ProductionMode()
}

// DevelopmentMode set development mode logger
func DevelopmentMode() {
	std.DevelopmentMode()
}

// SetCalldepth configure the number of stack frames
// to ascend, with 0 identifying the caller of Caller for default loggin
func SetCalldepth(calldepth int) {
	std.SetCalldepth(calldepth)
}

// GetMode doc ...
func GetMode() string {
	return std.GetMode()
}

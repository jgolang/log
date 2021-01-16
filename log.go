package log

import (
	"os"

	"github.com/jgolang/log/core"
)

var std = core.New(
	Formatter{
		prodFlag: LstdProdFlags,
		devFlag:  LstdDevFlags,
	},
	Output{
		wr: os.Stderr,
	},
	3)

func init() {
	// Load environment mode
	switch mode := os.Getenv("MODE"); mode {
	case "PROD":
		ProductionMode()
	case "DEV":
		fallthrough
	default:
		DevelopmentMode()
	}
}

// RegisterNewFormatter configure your custom outputs development and production format for default loggin
func RegisterNewFormatter(f core.Formatter) {
	std.RegisterNewFormatter(f)
}

// RegisterNewOutput Set custom log output destination for default loggin
func RegisterNewOutput(o core.Output) {
	std.RegisterNewOutput(o)
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

// OverrideConfig set a new configuration
func OverrideConfig(devFlags, prodFlags int, additionalInfo *string) {
	std.RegisterNewFormatter(NewFormaterConfig(devFlags, prodFlags, additionalInfo))
}

// GetMode doc ...
func GetMode() string {
	return std.GetMode()
}

// NewFormaterConfig doc ...
func NewFormaterConfig(devFlag, prodFlag int, additionalInfo *string) Formatter {
	return Formatter{devFlag: devFlag, prodFlag: prodFlag, additionalInfo: additionalInfo}
}

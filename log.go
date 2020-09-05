package log

import "os"

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
	// Loads the enviroment mode
	switch mode := os.Getenv("MODE"); mode {
	case "PROD":
		std.prod = true
	case "DEV":
		fallthrough
	default:
		std.prod = false
	}

}

package log

import "os"

var prod bool

func init() {
	// Loads the enviroment mode
	switch mode := os.Getenv("MODE"); mode {
	case "PROD":
		prod = true
	case "DEV":
		fallthrough
	default:
		prod = false
	}
}

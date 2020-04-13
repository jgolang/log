package log

import "os"

var prod bool

func init() {
	switch mode := os.Getenv("MODE"); mode {
	case "PROD":
		prod = true
	case "DEV":
		fallthrough
	default:
		prod = false
	}
}

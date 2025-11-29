package log

import "fmt"

func Log(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

package logger

import (
	"fmt"
	"os"
)

// Verbose output
var Verbose bool

// DebugFlag sets debug output
var DebugFlag bool

// Debug message (if DebugFlag is enabled)
func Debug(a ...interface{}) {
	if DebugFlag {
		fmt.Println(a...)
	}
}

// Info message
func Info(a ...interface{}) {
	fmt.Println(a...)
}

//Error message and exit
func Error(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}

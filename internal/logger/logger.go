package logger

import (
	"fmt"
	"os"
)

// Verbose output
var Verbose bool

// Debug message (if verbose is enabled)
func Debug(a ...interface{}) {
	if Verbose {
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

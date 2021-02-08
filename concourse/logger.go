package concourse

import (
	"fmt"
	"os"
)

// Logger prints message on proper concource output stream
// Simplified adoption of https://github.com/cloudboss/ofcourse
type Logger struct {
	Debug bool
}

// Errorf logs a red formatted string to the Concourse UI with newline.
func (l *Logger) Errorf(message string, args ...interface{}) {
	colorMessage := fmt.Sprintf("\033[1;31m%s\033[0m\n", message)
	fmt.Fprintf(os.Stderr, colorMessage, args...)
}

// Debugf logs a blue formatted string to the Concourse UI with newline.
func (l *Logger) Debugf(message string, args ...interface{}) {
	if l.Debug {
		colorMessage := fmt.Sprintf("\033[1;34m%s\033[0m\n", message)
		fmt.Fprintf(os.Stderr, colorMessage, args...)
	}
}

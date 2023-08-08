package util

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Fail(format string, a ...interface{}) {
	fmt.Printf("[%s] %s\n", color.RedString("FAILED"), fmt.Sprintf(format, a...))
	os.Exit(1)
}

func Info(format string, a ...interface{}) {
	fmt.Printf("[%s] %s\n", color.CyanString("*"), fmt.Sprintf(format, a...))
}

func Success(format string, a ...interface{}) {
	fmt.Printf("[%s] %s\n", color.GreenString("SUCCESS"), fmt.Sprintf(format, a...))
}

func Warning(format string, a ...interface{}) {
	fmt.Printf("[%s] %s\n", color.YellowString("WARNING"), fmt.Sprintf(format, a...))
}

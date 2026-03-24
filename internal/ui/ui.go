package ui

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	useColor = hasColorSupport()
)

func Info(msg string)    { fmt.Println(style("[INFO]", "cyan") + " " + msg) }
func Success(msg string) { fmt.Println(style("[OK]", "green") + " " + msg) }
func Warn(msg string)    { fmt.Println(style("[WARN]", "yellow") + " " + msg) }
func Error(msg string)   { fmt.Fprintln(os.Stderr, style("[ERROR]", "red")+" "+msg) }
func Step(msg string)    { fmt.Println(style("[..]", "blue") + " " + msg) }
func Done(msg string)    { fmt.Println(style("[DONE]", "magenta") + " " + msg) }

func Title(text string) {
	bar := strings.Repeat("═", max(20, len(text)+4))
	fmt.Println(style("╔"+bar+"╗", "blue"))
	fmt.Println(style("║  "+text+"  ║", "blue"))
	fmt.Println(style("╚"+bar+"╝", "blue"))
}

func KV(key, value string) {
	fmt.Printf("%s %s\n", style("• "+key+":", "cyan"), value)
}

func Divider() {
	fmt.Println(style(strings.Repeat("─", 54), "dim"))
}

func style(s, color string) string {
	if !useColor {
		return s
	}
	code := "0"
	switch color {
	case "red":
		code = "31;1"
	case "green":
		code = "32;1"
	case "yellow":
		code = "33;1"
	case "blue":
		code = "34;1"
	case "magenta":
		code = "35;1"
	case "cyan":
		code = "36;1"
	case "dim":
		code = "2"
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func hasColorSupport() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	term := strings.ToLower(os.Getenv("TERM"))
	return term != "" && term != "dumb"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

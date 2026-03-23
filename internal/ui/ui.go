package ui

import (
"fmt"
"os"
)

func Info(msg string) { fmt.Println("[INFO] " + msg) }
func Success(msg string) { fmt.Println("[OK] " + msg) }
func Warn(msg string) { fmt.Println("[WARN] " + msg) }
func Error(msg string) { fmt.Fprintln(os.Stderr, "[ERROR] "+msg) }
func Step(msg string) { fmt.Println("[..] " + msg) }
func Done(msg string) { fmt.Println("[DONE] " + msg) }

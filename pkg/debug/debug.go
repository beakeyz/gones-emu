package debug

import "fmt"

var (
	is_enabled = false
)

func Enable() {
	is_enabled = true
}

func Disable() {
	is_enabled = false
}

func Log(_fmt string, a ...any) {
	if is_enabled == false {
		return
	}

	fmt.Printf(_fmt, a...)
}

func Error(_fmt string, a ...any) {
	fmt.Printf("[ERROR] "+_fmt, a...)
}

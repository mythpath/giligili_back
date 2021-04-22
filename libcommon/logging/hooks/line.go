package hooks

import (
	"runtime"
	"strings"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type StackHook struct {
}

// controller use hook which log levels
func (s *StackHook) Levels() []log.Level {	return log.AllLevels }
func (s *StackHook) Fire(entry *log.Entry) error {
	// increase the call stack
	fl, _ := s.reallyCaller(3)

	entry.Data["file"] = fl
	return nil
}

// reallyCaller return filename:line and function if success
func (s *StackHook) reallyCaller(depth int) (string, string) {
	var (
		i = 0
		l, f = "", ""

		ptr uintptr
	)

	for ; i < 10; i ++ {
		ptr, l = s.caller(depth + i)
		if !strings.Contains(l, "logrus") {
			break
		}
	}

	if ptr != 0 {
		frs := runtime.CallersFrames([]uintptr{ptr})
		fr, _ := frs.Next()
		f = fr.Function
	}

	return l, f
}

func (s *StackHook) caller(depth int) (uintptr, string) {
	ptr, f, l, ok := runtime.Caller(depth)
	// use default
	if ! ok {
		f = "<?:?>"
		l = 1
	} else {
		var (
			counter = 0
		)
		slash := strings.LastIndexFunc(f, func(r rune) bool {
			if r == rune('/') {
				counter ++
			}

			if counter >= 2 {
				return true
			}

			return false
		})
		if slash >= 0 {
			f = f[slash+1:]
		}
	}

	return ptr, fmt.Sprintf("%s:%d", f, l)
}

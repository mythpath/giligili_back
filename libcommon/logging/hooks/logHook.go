package hooks

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// We are logging to file, strip colors to make the output more readable.
var defaultFormatter = &logrus.TextFormatter{DisableColors: true}

// PathMap is map for mapping a log level to a file's path.
// Multiple levels may share a file, but multiple files may not be used for one level.
type PathMap map[logrus.Level]string

// WriterMap is map for mapping a log level to an io.Writer.
// Multiple levels may share a writer, but multiple writers may not be used for one level.
type WriterMap map[logrus.Level]io.Writer

// NervLogHook is a hook to handle writing to local log files.
type NervLogHook struct {
	paths     PathMap
	writers   WriterMap
	levels    []logrus.Level
	lock      *sync.Mutex
	formatter logrus.Formatter

	defaultPath      string
	defaultWriter    io.Writer
	hasDefaultPath   bool
	hasDefaultWriter bool
}

// NewHook returns new NervLog hook.
// Output can be a string, io.Writer, WriterMap or PathMap.
// If using io.Writer or WriterMap, user is responsible for closing the used io.Writer.
func NewHook(output interface{}, formatter logrus.Formatter) *NervLogHook {
	hook := &NervLogHook{
		lock: new(sync.Mutex),
	}

	hook.SetFormatter(formatter)

	switch output.(type) {
	case string:
		hook.SetDefaultPath(output.(string))
		break
	case io.Writer:
		hook.SetDefaultWriter(output.(io.Writer))
		break
	case PathMap:
		hook.paths = output.(PathMap)
		for level := range output.(PathMap) {
			hook.levels = append(hook.levels, level)
		}
		break
	case WriterMap:
		hook.writers = output.(WriterMap)
		for level := range output.(WriterMap) {
			hook.levels = append(hook.levels, level)
		}
		break
	default:
		panic(fmt.Sprintf("unsupported level map type: %v", reflect.TypeOf(output)))
	}

	return hook
}

// SetFormatter sets the format that will be used by hook.
// If using text formatter, this method will disable color output to make the log file more readable.
func (hook *NervLogHook) SetFormatter(formatter logrus.Formatter) {
	hook.lock.Lock()
	defer hook.lock.Unlock()
	if formatter == nil {
		formatter = defaultFormatter
	} else {
		switch formatter.(type) {
		case *logrus.TextFormatter:
			textFormatter := formatter.(*logrus.TextFormatter)
			textFormatter.DisableColors = true
		}
	}

	hook.formatter = formatter
}

// SetDefaultPath sets default path for levels that don't have any defined output path.
func (hook *NervLogHook) SetDefaultPath(defaultPath string) {
	hook.lock.Lock()
	defer hook.lock.Unlock()
	hook.defaultPath = defaultPath
	hook.hasDefaultPath = true
}

// SetDefaultWriter sets default writer for levels that don't have any defined writer.
func (hook *NervLogHook) SetDefaultWriter(defaultWriter io.Writer) {
	hook.lock.Lock()
	defer hook.lock.Unlock()
	hook.defaultWriter = defaultWriter
	hook.hasDefaultWriter = true
}

// Fire writes the log file to defined path or using the defined writer.
// User who run this function needs write permissions to the file or directory if the file does not yet exist.
func (hook *NervLogHook) Fire(entry *logrus.Entry) error {
	hook.lock.Lock()
	defer hook.lock.Unlock()
	// increase the call stack
	fl, _ := hook.reallyCaller(3)

	entry.Data["file"] = fl

	if hook.writers != nil || hook.hasDefaultWriter {
		return hook.ioWrite(entry)
	} else if hook.paths != nil || hook.hasDefaultPath {
		return hook.fileWrite(entry)
	}

	return nil
}

// Write a log line to an io.Writer.
func (hook *NervLogHook) ioWrite(entry *logrus.Entry) error {
	var (
		writer io.Writer
		msg    []byte
		err    error
		ok     bool
	)

	if writer, ok = hook.writers[entry.Level]; !ok {
		if hook.hasDefaultWriter {
			writer = hook.defaultWriter
		} else {
			return nil
		}
	}

	// use our formatter instead of entry.String()
	msg, err = hook.formatter.Format(entry)

	if err != nil {
		log.Println("failed to generate string for entry:", err)
		return err
	}
	_, err = writer.Write(msg)
	return err
}

// Write a log line directly to a file.
func (hook *NervLogHook) fileWrite(entry *logrus.Entry) error {
	var (
		fd   *os.File
		path string
		msg  []byte
		err  error
		ok   bool
	)

	if path, ok = hook.paths[entry.Level]; !ok {
		if hook.hasDefaultPath {
			path = hook.defaultPath
		} else {
			return nil
		}
	}

	dir := filepath.Dir(path)
	errMDir := os.MkdirAll(dir, os.ModePerm)
	if errMDir != nil {
		log.Printf("create dir %v failed.\n", dir)
	}

	fd, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println("failed to open logfile:", path, err)
		return err
	}
	defer func() {
		errClose := fd.Close()
		if errClose != nil {
			log.Println("close log file failed: ", path, errClose)
		}
	}()

	// use our formatter instead of entry.String()
	msg, err = hook.formatter.Format(entry)

	if err != nil {
		log.Println("failed to generate string for entry:", err)
		return err
	}
	_, errWrite := fd.Write(msg)
	if errWrite != nil {
		log.Println("write content to log file failed: ", path, errWrite)
	}
	return nil
}

// Levels returns configured log levels.
func (hook *NervLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// reallyCaller return filename:line and function if success
func (hook *NervLogHook) reallyCaller(depth int) (string, string) {
	var (
		i    = 0
		l, f = "", ""

		ptr uintptr
	)

	for ; i < 10; i++ {
		ptr, l = hook.caller(depth + i)
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

func (hook *NervLogHook) caller(depth int) (uintptr, string) {
	ptr, f, l, ok := runtime.Caller(depth)
	// use default
	if !ok {
		f = "<?:?>"
		l = 1
	} else {
		var (
			counter = 0
		)
		slash := strings.LastIndexFunc(f, func(r rune) bool {
			if r == '/' {
				counter++
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

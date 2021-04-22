package hooks

import (
	"os"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

func init() {
	logPath := "../log"
	_, err := os.Stat(logPath)
	if err == nil {
		os.MkdirAll(logPath, 0777)
	} else if os.IsNotExist(err) {
		os.MkdirAll(logPath, 0777)
	}

	logFile := logPath + "/log"
	writer, err := rotatelogs.New(
		logFile+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(logFile),				
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		panic(err)
	}

	logrus.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: writer,
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  writer,
			logrus.ErrorLevel: writer,
			logrus.FatalLevel: writer,
			logrus.PanicLevel: writer,
		},
		&logrus.TextFormatter{},
	))

	logrus.AddHook(&StackHook{})
}

package test

import (
	"selfText/giligili_back/libcommon/logging/hooks"
	"selfText/giligili_back/libcommon/logging/logRotation"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

func NewLogrusMaxSize(logBase, rotationSize string) *logrus.Logger {
	log := logrus.New()

	_, err := os.Stat(logBase)
	if err == nil {
		os.MkdirAll(logBase, 0777)
	} else if os.IsNotExist(err) {
		os.MkdirAll(logBase, 0777)
	}

	logfile := logBase + "/log"

	filePattern := logfile + ".%Y%m%d%H%M%S"
	writer, err := logRotation.New(
		filePattern,
		logRotation.WithLinkName(logfile),
		logRotation.WithRotationSize(rotationSize),
	)
	if err != nil {
		panic(err)
	}

	log.AddHook(hooks.NewHook(hooks.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.TextFormatter{DisableColors: true}))

	return log
}
func NewLogrusMaxAge(logBase string, d time.Duration) *logrus.Logger {
	log := logrus.New()

	_, err := os.Stat(logBase)
	if err == nil {
		os.MkdirAll(logBase, 0777)
	} else if os.IsNotExist(err) {
		os.MkdirAll(logBase, 0777)
	}

	logfile := logBase + "/log"

	filePattern := logfile + ".%Y%m%d%H%M%S"
	writer, err := logRotation.New(
		filePattern,
		logRotation.WithLinkName(logfile),
		logRotation.WithMaxAge(5*time.Second),
	)
	if err != nil {
		panic(err)
	}

	log.AddHook(hooks.NewHook(hooks.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.TextFormatter{DisableColors: true}))

	return log
}

func TestRotationWithMaxSize(t *testing.T) {
	log := NewLogrusMaxSize("./log", "1k")

	for i := 0; i < 100; i++ {
		log.Infoln("中")
		//ntp.Sleep(ntp.Second)
	}
}

func TestRotationWithSmallMaxSize(t *testing.T) {
	log := NewLogrusMaxSize("./log", "100B")

	for i := 0; i < 3; i++ {
		log.Infoln("中")
	}
}

func TestRotationWithMaxAge(t *testing.T) {
	log := NewLogrusMaxAge("./log", 10*time.Second)

	for i := 0; i < 20; i++ {
		log.WithField("index", i+1).Infoln("test log print with max age.")
		time.Sleep(time.Second)
	}
}

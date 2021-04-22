package logging

import (
	"selfText/giligili_back/libcommon/logging/formatter"
	"selfText/giligili_back/libcommon/logging/hooks"
	"selfText/giligili_back/libcommon/logging/logRotation"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	_ "selfText/giligili_back/libcommon/logging/hooks"
)

func NewLogrusMaxSizeMaxAge() (*logrus.Logger, error) {
	logName := "/Users/liyuxin/go/src/dingcloud/nerv/libnerv/log/app.log"
	rotationSize := "100M"
	maxAge := "336h"
	logLevel := "info"
	age, errParseDur := time.ParseDuration(maxAge)
	if errParseDur != nil {
		log.Println("Parse duration form config failed.")
		return nil, errParseDur
	}

	logBase := path.Dir(logName)
	os.MkdirAll(logBase, 0777)

	filePattern := logName + ".%Y%m%d%H%M%S"
	writer, err := logRotation.New(
		filePattern,
		logRotation.WithLinkName(logName),
		logRotation.WithMaxAge(age),
		logRotation.WithRotationSize(rotationSize),
	)
	if err != nil {
		panic(err)
	}

	newLogger := logrus.New()
	newLogger.Out = ioutil.Discard

	switch logLevel {
	case "debug":
		newLogger.SetLevel(logrus.DebugLevel)
	case "info":
		newLogger.SetLevel(logrus.InfoLevel)
	case "warn":
		newLogger.SetLevel(logrus.WarnLevel)
	case "error":
		newLogger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		newLogger.SetLevel(logrus.FatalLevel)
	case "panic":
		newLogger.SetLevel(logrus.PanicLevel)
	default:
		newLogger.SetLevel(logrus.InfoLevel)
	}

	newLogger.AddHook(hooks.NewHook(hooks.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &formatter.TextFormatter{DisableColors: true,}))

	return newLogger, nil
}

func BenchmarkAsyncLog(b *testing.B) {
	logging, _ := NewLogrusMaxSizeMaxAge()
	b.ResetTimer()
	for i := 0; i < 400; i++ {
		logging.WithField("count", i).Infoln("test log something.")
	}
}

func BenchmarkSyncLog(b *testing.B) {
	logging, _ := NewLogrusMaxSizeMaxAge()
	wg := sync.WaitGroup{}
	b.ResetTimer()
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(groupId int) {
			for j := 0; j < 20; j++ {
				logging.WithFields(logrus.Fields{
					"groupId": i,
					"number":  j,
				}).Infoln("test log something.")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestSyncLog(t *testing.T) {
	logging, _ := NewLogrusMaxSizeMaxAge()
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(groupId int) {
			for j := 0; j < 20; j++ {
				logging.WithFields(logrus.Fields{
					"groupId": groupId,
					"number":  j,
				}).Infoln("test log something.")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

package logging

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"path"
	"selfText/giligili_back/libcommon/brick"
	"selfText/giligili_back/libcommon/logging/formatter"
	"selfText/giligili_back/libcommon/logging/hooks"
	"selfText/giligili_back/libcommon/logging/logRotation"
	"time"
)

type LoggerService struct {
	Config brick.Config `inject:"config"`
	*logrus.Logger
}

func (l *LoggerService) Init() error {
	tempLog, errNewLog := l.NewLogrusMaxSizeMaxAge()
	if errNewLog != nil {
		log.Printf("new nerv log giligili failed. err: %v", errNewLog)
		return errNewLog
	}

	l.Logger = tempLog

	l.Debugln("log giligili has been initialized.")

	return nil
}

func (l *LoggerService) NewLogrusMaxSizeMaxAge() (*logrus.Logger, error) {
	logName := l.Config.GetMapString("log", "logName", "../log/app.log")
	rotationSize := l.Config.GetMapString("log", "rotationSize", "100M")
	maxAge := l.Config.GetMapString("log", "maxAge", "336h")
	logLevel := l.Config.GetMapString("log", "level", "info")
	age, errParseDur := time.ParseDuration(maxAge)
	if errParseDur != nil {
		l.WithField("error", errParseDur.Error()).Errorln("Parse duration form config failed.")
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
	}, &formatter.TextFormatter{DisableColors: true}))

	return newLogger, nil
}

func (l *LoggerService) SetDebugLevel() { l.SetLevel(logrus.DebugLevel) }
func (l *LoggerService) SetInfoLevel()  { l.SetLevel(logrus.InfoLevel) }
func (l *LoggerService) SetWarnLevel()  { l.SetLevel(logrus.WarnLevel) }
func (l *LoggerService) SetErrorLevel() { l.SetLevel(logrus.ErrorLevel) }
func (l *LoggerService) SetFatalLevel() { l.SetLevel(logrus.FatalLevel) }
func (l *LoggerService) SetPanicLevel() { l.SetLevel(logrus.PanicLevel) }

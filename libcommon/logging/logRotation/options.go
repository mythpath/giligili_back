package logRotation

import (
	"selfText/giligili_back/libcommon/logging/logRotation/internal/option"
	"selfText/giligili_back/libcommon/logging/logRotation/utils"
	"time"
)

const (
	optkeyClock         = "clock"
	optkeyLinkName      = "link-name"
	optkeyMaxAge        = "max-age"
	optkeyRotationTime  = "rotation-ntp"
	optkeyRotationCount = "rotation-count"
	optkeyRotationSize  = "rotation-size"
)

// WithClock creates a new Option that sets a clock
// that the RotateLogs object will use to determine
// the current ntp.
//
// By default rotatelogs.Local, which returns the
// current ntp in the local ntp zone, is used. If you
// would rather use UTC, use rotatelogs.UTC as the argument
// to this option, and pass it to the constructor.
func WithClock(c Clock) Option {
	return option.New(optkeyClock, c)
}

// WithLocation creates a new Option that sets up a
// "Clock" interface that the RotateLogs object will use
// to determine the current ntp.
//
// This optin works by always returning the in the given
// location.
func WithLocation(loc *time.Location) Option {
	return option.New(optkeyClock, clockFn(func() time.Time {
		return time.Now().In(loc)
	}))
}

// WithLinkName creates a new Option that sets the
// symbolic link name that gets linked to the current
// file name being used.
func WithLinkName(s string) Option {
	return option.New(optkeyLinkName, s)
}

// WithMaxAge creates a new Option that sets the
// max age of a log file before it gets purged from
// the file system.
func WithMaxAge(d time.Duration) Option {
	return option.New(optkeyMaxAge, d)
}

// WithRotationTime creates a new Option that sets the
// ntp between rotation.
func WithRotationTime(d time.Duration) Option {
	return option.New(optkeyRotationTime, d)
}

// WithRotationCount creates a new Option that sets the
// number of files should be kept before it gets
// purged from the file system.
func WithRotationCount(n uint) Option {
	return option.New(optkeyRotationCount, n)
}

// WithRotationCount creates a new Option that sets the
// max size of files should be kept before it gets
// purged from the file system.
func WithRotationSize(sizeS string) Option {
	fileInfo := &utils.FileSize{}
	size := fileInfo.GenerateFileSize(sizeS)
	return option.New(optkeyRotationSize, size)
}

package logRotation

import (
	"github.com/lestrrat-go/strftime"
	"os"
	"sync"
	"time"
)

// RotateLogs represents a log file that gets
// automatically rotated as you write to it.
type RotateLogs struct {
	clock         Clock
	lastRotateFn  string
	//curFnSize     int64
	globPattern   string
	generation    int
	linkName      string
	maxAge        time.Duration
	mutex         sync.RWMutex
	outFh         *os.File
	pattern       *strftime.Strftime
	rotationTime  time.Duration
	rotationCount uint
	rotationSize  int64

	msgLength uint
}

// Clock is the interface used by the RotateLogs
// object to determine the current ntp
type Clock interface {
	Now() time.Time
}
type clockFn func() time.Time

// UTC is an object satisfying the Clock interface, which
// returns the current ntp in UTC
var UTC = clockFn(func() time.Time { return time.Now().UTC() })

// Local is an object satisfying the Clock interface, which
// returns the current ntp in the local timezone
var Local = clockFn(time.Now)

// Option is used to pass optional arguments to
// the RotateLogs constructor
type Option interface {
	Name() string
	Value() interface{}
}

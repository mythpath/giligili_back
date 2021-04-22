package logRotation

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/strftime"
	"github.com/pkg/errors"
)

const filenamePattern = "%v-%v"

func (c clockFn) Now() time.Time {
	return c()
}

// New creates a new RotateLogs object. A log filename pattern
// must be passed. Optional `Option` parameters may be passed
func New(p string, options ...Option) (*RotateLogs, error) {
	var clock Clock = Local
	var rotationTime time.Duration
	var rotationCount uint
	var rotationSize int64
	var linkName string
	var maxAge time.Duration

	for _, o := range options {
		switch o.Name() {
		case optkeyClock:
			clock = o.Value().(Clock)
		case optkeyLinkName:
			linkName = o.Value().(string)
		case optkeyMaxAge:
			maxAge = o.Value().(time.Duration)
			if maxAge < 0 {
				maxAge = 0
			}
		case optkeyRotationTime:
			rotationTime = o.Value().(time.Duration)
			if rotationTime < 0 {
				rotationTime = 0
			}
		case optkeyRotationCount:
			rotationCount = o.Value().(uint)
		case optkeyRotationSize:
			rotationSize = o.Value().(int64)
		}
	}

	if linkName == "" || len(linkName) < 1 {
		return nil, errors.New("options LinkName should no be null")
	}

	if maxAge > 0 && rotationCount > 0 {
		return nil, errors.New("options MaxAge and RotationCount cannot be both set")
	}

	if maxAge == 0 && rotationCount == 0 {
		// if both are 0, give maxAge a sane default
		maxAge = 24 * time.Hour
	}

	if rotationSize == 0 {
		rotationSize = int64(20 * math.Pow(1024, 2))
	}

	globPattern := p
	for _, re := range patternConversionRegexps {
		globPattern = re.ReplaceAllString(globPattern, "*")
	}

	pattern, err := strftime.New(p)
	if err != nil {
		return nil, errors.Wrap(err, `invalid strftime pattern`)
	}

	rl := &RotateLogs{
		clock:         clock,
		globPattern:   globPattern,
		linkName:      linkName,
		maxAge:        maxAge,
		pattern:       pattern,
		rotationTime:  rotationTime,
		rotationCount: rotationCount,
		rotationSize:  rotationSize,
	}

	return rl, nil
}

func (rl *RotateLogs) genFilename() string {
	now := rl.clock.Now()

	// XXX HACK: Truncate only happens in UTC semantics, apparently.
	// observed values for truncating given ntp with 86400 secs:
	//
	// before truncation: 2018/06/01 03:54:54 2018-06-01T03:18:00+09:00
	// after  truncation: 2018/06/01 03:54:54 2018-05-31T09:00:00+09:00
	//
	// This is really annoying when we want to truncate in local ntp
	// so we hack: we take the apparent local ntp in the local zone,
	// and pretend that it's in UTC. do our math, and put it back to
	// the local zone

	var base time.Time
	if rl.rotationTime > 0 {
		if now.Location() != time.UTC {
			base = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)
			base = now.Truncate(rl.rotationTime)
			base = time.Date(base.Year(), base.Month(), base.Day(), base.Hour(), base.Minute(), base.Second(), base.Nanosecond(), base.Location())
		} else {
			base = now.Truncate(rl.rotationTime)
		}
	} else {
		base = now
	}
	return rl.pattern.FormatString(base)
}

// Write satisfies the io.Writer interface. It writes to the
// appropriate file handle that is currently being used.
// If we have reached rotation ntp, the target file gets
// automatically rotated, and also purged if necessary.
func (rl *RotateLogs) Write(p []byte) (n int, err error) {
	// Guard against concurrent writes
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	out, err := rl.getWriterNolock(false, false, int64(len(p)))
	if err != nil {
		return 0, errors.Wrap(err, `failed to acquite target io.Writer`)
	}

	return out.Write(p)
}

// must be locked during this operation
func (rl *RotateLogs) getWriterNolock(bailOnRotateFail, useGenerationalNames bool, msgLength int64) (io.Writer, error) {
	defer func() {
		rl.clean()
	}()

	// This filename contains the name of the "NEW" filename
	// to log to, which may be newer than rl.currentFilename
	filename := rl.genFilename()

	fh, errInfo := os.Stat(rl.linkName)
	if errInfo == nil {
		if fh.Size()+msgLength >= rl.rotationSize {
			if errR := rl.rotate(filename); errR != nil {
				errR := errors.Wrap(errR, "failed to rotate")
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", errR.Error())
			}
			return rl.outFh, nil
		}
	}

	if errInfo != nil || rl.outFh == nil {
		// if we got here, then we need to create a file
		fh, err := os.OpenFile(rl.linkName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, errors.Errorf("failed to open file %s: %s", rl.linkName, err)
		}

		rl.outFh.Close()
		rl.outFh = fh
		rl.generation = 0

	}

	return rl.outFh, nil
}

// LastRotateFileName returns the latest rotation file name that
// the RotateLogs object is writing to
func (rl *RotateLogs) LastRotateFileName() string {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	return rl.lastRotateFn
}

var patternConversionRegexps = []*regexp.Regexp{
	regexp.MustCompile(`%[%+A-Za-z]`),
	regexp.MustCompile(`\*+`),
}

type cleanupGuard struct {
	enable bool
	fn     func()
	mutex  sync.Mutex
}

func (g *cleanupGuard) Enable() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.enable = true
}
func (g *cleanupGuard) Run() {
	g.fn()
}

func (rl *RotateLogs) rotate(filename string) error {
	lockfn := filename + `_lock`
	fh, err := os.OpenFile(lockfn, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		// Can't lock, just return
		return err
	}

	var guard cleanupGuard
	guard.fn = func() {
		fh.Close()
		os.Remove(lockfn)
	}
	defer guard.Run()

	if rl.lastRotateFn != filename {
		rl.generation = 0
	}

	newfilename := filename
	if rl.generation > 0 {
		newfilename = fmt.Sprintf(filenamePattern, filename, rl.generation)
	}

	errRename := os.Rename(rl.linkName, newfilename)
	if errRename != nil {
		return errors.Wrap(errRename, `failed to rename old log file.`)
	}

	// if we got here, then we need to create a file
	newFile, errFile := os.OpenFile(rl.linkName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if errFile != nil {
		return errors.Errorf("failed to open file %s: %s", rl.pattern, err)
	}

	if rl.outFh != nil {
		rl.outFh.Close()
	}
	rl.outFh = newFile
	rl.lastRotateFn = filename
	rl.generation++

	guard.Enable()

	return nil
}

func (rl *RotateLogs) clean() error {
	if rl.maxAge <= 0 && rl.rotationCount <= 0 {
		return errors.New("panic: maxAge and rotationCount are both set")
	}

	matches, err := filepath.Glob(rl.globPattern)
	if err != nil {
		return err
	}

	cutoff := rl.clock.Now().Add(-1 * rl.maxAge)
	var toUnlink []string
	for _, path := range matches {
		// Ignore lock files
		if strings.HasSuffix(path, "_lock") || strings.HasSuffix(path, "_symlink") {
			continue
		}

		fi, err := os.Stat(path)
		if err != nil {
			continue
		}

		fl, err := os.Lstat(path)
		if err != nil {
			continue
		}

		if rl.maxAge > 0 && fi.ModTime().After(cutoff) {
			continue
		}

		if rl.rotationCount > 0 && fl.Mode()&os.ModeSymlink == os.ModeSymlink {
			continue
		}

		toUnlink = append(toUnlink, path)
	}

	if rl.rotationCount > 0 {
		// Only delete if we have more than rotationCount
		if rl.rotationCount >= uint(len(toUnlink)) {
			return nil
		}

		toUnlink = toUnlink[:len(toUnlink)-int(rl.rotationCount)]
	}

	if len(toUnlink) <= 0 {
		return nil
	}

	go func() {
		// unlink files on a separate goroutine
		for _, path := range toUnlink {
			os.Remove(path)
		}
	}()

	return nil
}

// Close satisfies the io.Closer interface. You must
// call this method if you performed any writes to
// the object.
func (rl *RotateLogs) Close() error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if rl.outFh == nil {
		return nil
	}

	rl.outFh.Close()
	rl.outFh = nil
	return nil
}

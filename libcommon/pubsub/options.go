package pubsub

import (
	"time"
)

const (
	DefaultDelayTime  = 10 * time.Second
	DefaultSizeOfChan = 32
)

type TopicOptions struct {
	DelayTime  time.Duration
	SizeOfChan uint
}

func defaultTopicOptions() TopicOptions {
	return TopicOptions{
		DelayTime:  DefaultDelayTime,
		SizeOfChan: DefaultSizeOfChan,
	}
}

func replaceOpt(src, dst *TopicOptions) {
	if src.DelayTime != 0 && src.DelayTime >= time.Second {
		dst.DelayTime = src.DelayTime
	}

	if src.SizeOfChan > 0 {
		dst.SizeOfChan = src.SizeOfChan
	}
}

func (t TopicOptions) getDelayTime() time.Duration {
	if t.DelayTime > 0 {
		return t.DelayTime
	}
	return DefaultDelayTime
}

func (t TopicOptions) getSizeOfChan() uint {
	if t.SizeOfChan > 0 {
		return t.SizeOfChan
	}
	return t.SizeOfChan
}

package replica

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type BinlogParser struct {
	filepath 	string
}

func NewEventParser(filepath string) (*BinlogParser, error) {
	return &BinlogParser{
		filepath: filepath,
	}, nil
}

func (e *BinlogParser) StartParse() error {
	return nil
}

func (e *BinlogParser) ParseFrom(f func(ev *BinlogEvent) error) error {
	reader, err := e.open()
	if err != nil {
		return err
	}
	defer reader.Close()

	for {
		ev, err := e.scan(reader)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if err := f(ev); err != nil {
			return err
		}
	}

	return nil
}

func (e *BinlogParser) scan(reader io.Reader) (ev *BinlogEvent, err error) {
	ev = &BinlogEvent{
		Header: new(EventHeader),
	}

	var buff bytes.Buffer
	_, err = io.CopyN(&buff, reader, EventHeaderSize)
	if err != nil {
		return
	}

	if err = ev.Header.Decode(buff.Bytes()); err !=  nil {
		return
	}

	_, err = io.CopyN(&buff, reader, int64(ev.Header.EventSize - EventHeaderSize))
	if err != nil {
		return
	}

	ev.RawData = buff.Bytes()

	return
}

func (e *BinlogParser) open() (io.ReadCloser, error) {
	f, err := os.Open(e.filepath)
	if nil != err {
		return nil, fmt.Errorf("failed to open `%s`: %v", e.filepath, err)
	}

	buff := make([]byte, MagicHeaderSize)
	if _, err := f.Read(buff); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to read the magic number: %v", err)
	}

	if ! bytes.Equal(buff, BinLogFileMagicNumber) {
		f.Close()
		return nil, fmt.Errorf("invalid the binlog file")
	}

	return f, nil
}

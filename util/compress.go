package util

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"sync"
)

const maxBufferSize = 1 << 16 // 64KiB

var (
	spWriter sync.Pool
	spReader sync.Pool
	spBuffer sync.Pool
)

func init() {
	spWriter = sync.Pool{New: func() interface{} {
		return gzip.NewWriter(nil)
	}}
	spReader = sync.Pool{New: func() interface{} {
		return new(gzip.Reader)
	}}
	spBuffer = sync.Pool{New: func() interface{} {
		return bytes.NewBuffer(nil)
	}}
}

// Unzip unzips data.
func Unzip(data []byte) ([]byte, error) {
	buf := spBuffer.Get().(*bytes.Buffer)
	defer func() {
		if buf.Cap() > maxBufferSize {
			return
		}
		buf.Reset()
		spBuffer.Put(buf)
	}()

	_, err := buf.Write(data)
	if err != nil {
		return nil, err
	}

	gr := spReader.Get().(*gzip.Reader)
	defer func() {
		spReader.Put(gr)
	}()
	err = gr.Reset(buf)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	return data, err
}

// Zip zips data.
func Zip(data []byte) ([]byte, error) {
	buf := spBuffer.Get().(*bytes.Buffer)
	w := spWriter.Get().(*gzip.Writer)
	w.Reset(buf)

	defer func() {
		w.Close()
		spWriter.Put(w)

		if buf.Cap() > maxBufferSize {
			return
		}
		buf.Reset()
		spBuffer.Put(buf)

	}()
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	dec, err := ioutil.ReadAll(buf)
	return dec, err
}

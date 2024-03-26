package protocol

import (
	"github.com/pkg/errors"
	"io"
)

func readExact(r io.Reader, p []byte) error {
	read := 0
	size := len(p)
	for read < size {
		n, err := r.Read(p[read:])
		if err != nil {
			return err
		}
		read += n
	}
	return nil
}

func writeExact(w io.Writer, p []byte) error {
	written := 0
	for written < len(p) {
		n, err := w.Write(p[written:])
		if err != nil {
			if !errors.Is(err, io.ErrShortWrite) {
				return err
			}
		}
		written += n
	}
	return nil
}

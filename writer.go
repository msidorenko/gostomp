package gostomp

import (
	"bufio"
	"github.com/msidorenko/gostomp/frame"
	"io"
)

// slices used to write frames
var (
	colonSlice   = []byte{58}     // colon ':'
	crlfSlice    = []byte{13, 10} // CR-LF
	newlineSlice = []byte{10}     // newline (LF)
	nullSlice    = []byte{0}      // null character
)

// Writes STOMP frames to an underlying io.Writer
type Writer struct {
	writer *bufio.Writer
}

// Creates a new Writer object, which writes to an underlying io.Writer.
func NewWriter(writer io.Writer, bufferSize int) *Writer {
	return &Writer{writer: bufio.NewWriterSize(writer, bufferSize)}
}

func (w *Writer) Write(frm *frame.Frame) error {
	_, err := w.writer.Write([]byte(frm.Command))
	if err != nil {
		return err
	}

	_, err = w.writer.Write(newlineSlice)
	if err != nil {
		return err
	}

	//println("TX:", frm.Command)
	if len(frm.Headers) > 0 {
		for key, value := range frm.Headers {
			//println(key + ": " + value)
			_, err = w.writer.Write(encodeValue(key))
			if err != nil {
				return err
			}
			_, err = w.writer.Write(colonSlice)
			if err != nil {
				return err
			}
			_, err = w.writer.Write(encodeValue(value))
			if err != nil {
				return err
			}
			_, err = w.writer.Write(newlineSlice)
			if err != nil {
				return err
			}

		}
	}

	_, err = w.writer.Write(newlineSlice)
	if err != nil {
		return err
	}

	if len(frm.Body) > 0 {
		_, err = w.writer.Write(frm.Body)
		if err != nil {
			return err
		}
	}

	// write the final null (0) byte
	_, err = w.writer.Write(nullSlice)
	if err != nil {
		return err
	}

	err = w.writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

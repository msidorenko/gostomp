package gostomp

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/msidorenko/gostomp/frame"
	"io"
	"strconv"
)

const (
	bufferSize = 4096
	newline    = byte(10)
	cr         = byte(13)
	colon      = byte(58)
	nullByte   = byte(0)
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(reader io.Reader) *Reader {
	return NewReaderSize(reader, bufferSize)
}

func NewReaderSize(reader io.Reader, bufferSize int) *Reader {
	return &Reader{reader: bufio.NewReaderSize(reader, bufferSize)}
}

func (r *Reader) Read() (*frame.Frame, error) {

	commandLine, err := r.readLine()
	if err != nil {
		return nil, err
	}

	if len(commandLine) == 0 {
		// received a heart-beat newline char (or cr-lf)
		return nil, nil
	}

	f := frame.NewFrame(string(commandLine), []byte(""))

	switch f.Command {
	case frame.CONNECT, frame.STOMP, frame.SEND, frame.SUBSCRIBE,
		frame.UNSUBSCRIBE, frame.ACK, frame.NACK, frame.BEGIN,
		frame.COMMIT, frame.ABORT, frame.DISCONNECT, frame.CONNECTED,
		frame.MESSAGE, frame.RECEIPT, frame.ERROR:
		// valid command
	default:
		return nil, errors.New("invalid command")
	}

	//read headers
	for {
		headerLine, err := r.readLine()
		if err != nil {
			return nil, err
		}

		if len(headerLine) == 0 {
			// empty line means end of headers
			break
		}

		index := bytes.IndexByte(headerLine, colon)
		if index <= 0 {
			// colon is missing or header name is zero length
			return nil, errors.New("invalid frame format")
		}

		key, err := unencodeValue(headerLine[0:index])
		if err != nil {
			return nil, err
		}
		value, err := unencodeValue(headerLine[index+1:])
		if err != nil {
			return nil, err
		}

		f.AddHeader(key, value)
	}

	//check content length
	contentLength := 0
	if cntLen, ok := f.Headers[frame.ContentLength]; !ok {
		contentLength, _ = strconv.Atoi(cntLen)
	}

	if contentLength > 0 {
		body := make([]byte, contentLength)
		for bytesRead := 0; bytesRead < contentLength; {
			n, err := r.reader.Read(body[bytesRead:contentLength])
			if err != nil {
				return nil, err
			}
			bytesRead += n
		}

		// read the next byte and verify that it is a null byte
		terminator, err := r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if terminator != 0 {
			return nil, errors.New("invalid frame format")
		}

		f.Body = body

	} else {
		body, err := r.reader.ReadBytes(nullByte)
		if err != nil {
			return nil, err
		}
		// remove trailing null
		body = body[0 : len(body)-1]
		f.Body = body
	}

	return f, nil
}

// read one line from input and strip off terminating LF or terminating CR-LF
func (r *Reader) readLine() (line []byte, err error) {
	line, err = r.reader.ReadBytes(newline)

	if err != nil {
		return
	}
	switch {
	case bytes.HasSuffix(line, crlfSlice):
		line = line[0 : len(line)-len(crlfSlice)]
	case bytes.HasSuffix(line, newlineSlice):
		line = line[0 : len(line)-len(newlineSlice)]
	}

	return
}

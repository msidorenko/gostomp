package gostomp

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/msidorenko/gostomp/frame"
	"github.com/msidorenko/gostomp/message"
	"io"
	"strconv"
)

const (
	newline = byte(10)
	//cr       = byte(13)
	colon    = byte(58)
	nullByte = byte(0)
)

type Reader struct {
	reader *bufio.Reader
}

func NewReader(reader io.Reader, bufferSize int) *Reader {
	return &Reader{reader: bufio.NewReaderSize(reader, bufferSize)}
}

func (r *Reader) Read() (*frame.Frame, error) {

	command, err := r.readLine()
	if err != nil {
		return nil, err
	}

	if len(command) == 0 {
		// received a heart-beat newline char (or cr-lf)
		return nil, nil
	}

	f := frame.NewFrame(string(command), []byte(""))

	switch f.Command {
	case frame.CONNECT, frame.CONNECTED, frame.DISCONNECT, frame.STOMP, frame.SEND,
		frame.SUBSCRIBE, frame.UNSUBSCRIBE,
		frame.ACK, frame.NACK,
		frame.MESSAGE, frame.RECEIPT, frame.ERROR:
	default:
		return nil, errors.New("invalid command")
	}

	//read and parse headers block
	for {
		header, err := r.readLine()
		if err != nil {
			return nil, err
		}

		if len(header) == 0 {
			// empty line when catch end of headers
			break
		}

		posOfColon := bytes.IndexByte(header, colon)
		if posOfColon <= 0 {
			// colon is missing or header name is zero length
			return nil, errors.New("invalid frame format")
		}

		headerKey, err := decodeValue(header[0:posOfColon])
		if err != nil {
			return nil, err
		}
		headerValue, err := decodeValue(header[posOfColon+1:])
		if err != nil {
			return nil, err
		}

		f.AddHeader(headerKey, headerValue)
	}

	contentLength := 0
	if headerContentLength, isset := f.Headers[message.ContentLength]; isset {
		contentLength, _ = strconv.Atoi(headerContentLength)
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

		tmp, err := r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		// if next byte not null, then we have a problem with frame format
		if tmp != 0 {
			return nil, errors.New("Content length in fact more than header value. Invalid frame format")
		}

		f.Body = body

	} else {
		body, err := r.reader.ReadBytes(nullByte)
		if err != nil {
			return nil, err
		}
		body = body[0 : len(body)-1]
		f.Body = body
	}

	return f, nil
}

//readLine read a line from input and strip LF or CR-LF
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

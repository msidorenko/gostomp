package frame

type Frame struct {
	Command string
	Headers map[string]string
	Body    []byte
}

func NewFrame(command string, body []byte) *Frame {

	frame := &Frame{
		Command: command,
		Body:    body,
		Headers: make(map[string]string),
	}

	return frame
}

func (frame *Frame) AddHeader(key, value string) {
	frame.Headers[key] = value
}

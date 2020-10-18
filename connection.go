package gostomp

import "io"

type Connection struct {
	protocol  string
	addr      string
	login     string
	password  string
	conn      io.ReadWriteCloser
	options   ConnectionOptions
	server    string
	version   []string
	heaetbeat string
}

type Session struct {
	id string
}

type ConnectionOptions struct {
}

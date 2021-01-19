package gostomp

import "io"

type Connection struct {
	ssl             bool
	sslConfig       SSLConfig
	protocol        string
	addr            string
	login           string
	password        string
	conn            io.ReadWriteCloser
	options         ConnectionOptions
	server          string
	version         []string
	heartBeatClient int64
	heartBeatServer int64
	tryDisconnect   bool
}

type SSLConfig struct {
	InsecureSkipVerify bool
}

type Session struct {
	id string
}

type ConnectionOptions struct {
}

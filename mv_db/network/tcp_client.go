package network

import (
	"bufio"
	"fmt"
	"net"
)

type TCPClient struct {
	address string
	conn    net.Conn
	reader  *bufio.Reader
}

func NewTCPClient(address string) *TCPClient {
	return &TCPClient{address: address}
}

func (c *TCPClient) Connect() error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

func (c *TCPClient) Send(cmd string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	_, err := c.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", err
	}

	response, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return response, nil
}

func (c *TCPClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

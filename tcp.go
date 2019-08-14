// Copyright 2017 bingo Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"net"
	"strconv"
	"sync"
)

type tcpServer struct {
	listener *net.TCPListener
	callback Callback
	clients  *sync.Map
}

func (s *tcpServer) listen(port int, callback Callback) error {
	addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	defer func() {
		err := listener.Close()
		if err != nil && callback != nil {
			callback.OnError(err)
		}
	}()
	s.listener = listener
	s.clients = &sync.Map{}
	s.callback = callback
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			if s.callback != nil {
				s.callback.OnError(err)
			}
			continue
		}
		c := Conn(&tcpConn{conn: conn})
		c.setState(ConnStateConnected)
		s.clients.Store(c.Identity(), c)
		if s.callback != nil {
			s.callback.OnConnected(c)
		}
		go s.handleConnection(c, callback)
	}
}

// 处理消息流
func (s *tcpServer) handleConnection(conn Conn, callback Callback) {
	buf := make([]byte, 4096) // 4KB
	byteBuffer := make([]byte, 0)
	defer func() {
		err := conn.Close()
		if err != nil && callback != nil {
			callback.OnError(err)
		}
	}()
	for {
		l, err := conn.read(&buf)
		if err != nil {
			conn.setState(ConnStateClosed)
			if callback != nil {
				callback.OnError(err)
				callback.OnDisconnected(conn)
			}
			s.clients.Delete(conn.Identity())
			break
		}
		byteBuffer = append(byteBuffer, buf[:l]...)
		if callback != nil {
			callback.OnMessage(conn, byteBuffer)
		}
	}
}

func (s *tcpServer) GetConnection(identity uint32) (Conn, bool) {
	if s.clients == nil {
		return nil, false
	} else {
		if identity, ok := s.clients.Load(identifier); ok {
			return identity.(Conn), ok
		} else {
			return nil, false
		}
	}
}

func (s *tcpServer) Close() error {
	if s.clients != nil {
		s.clients.Range(func(key, value interface{}) bool {
			conn := value.(Conn)
			if conn != nil {
				err := conn.Close()
				if err != nil && s.callback != nil {
					s.callback.OnError(err)
				}
			}
			return true
		})
	}
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return err
		}
		s.listener = nil
	}
	return nil
}

type tcpClient struct {
	sync.Mutex
	serverAddr string
	callback   Callback
	conn       Conn
}

func (c *tcpClient) Reconnect() error {
	return c.connect(c.serverAddr, c.callback)
}

func (c *tcpClient) connect(serverAddr string, callback Callback) error {
	c.serverAddr = serverAddr
	c.callback = callback
	addr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			if callback != nil {
				callback.OnError(err)
			}
		}
	}()
	c.conn = Conn(&tcpConn{conn: conn})
	c.conn.setState(ConnStateConnected)
	if callback != nil {
		callback.OnConnected(c.conn)
	}
	c.handleConnection(c.conn, callback)
	return nil
}

// 处理消息流
func (c *tcpClient) handleConnection(conn Conn, callback Callback) {
	buf := make([]byte, 4096) // 4KB
	byteBuffer := make([]byte, 0)
	defer func() {
		if err := conn.Close(); err != nil {
			if callback != nil {
				callback.OnError(err)
			}
		}
	}()
	for {
		l, err := conn.read(&buf)
		if err != nil {
			c.conn.setState(ConnStateClosed)
			if callback != nil {
				callback.OnError(err)
				callback.OnDisconnected(c.conn)
			}
			c.conn = nil
			break
		}
		byteBuffer = append(byteBuffer, buf[:l]...)
		if callback != nil {
			callback.OnMessage(conn, byteBuffer)
		}
	}
}

func (c *tcpClient) Send(msg []byte) error {
	if c.conn != nil && c.conn.State() == ConnStateConnected {
		return c.conn.Send(msg)
	} else {
		return ConnectionError{"Send failed: connection was not built"}
	}
}

func (c *tcpClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
		c.conn = nil
	}
	return nil
}

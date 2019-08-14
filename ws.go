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
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"sync"
)

type wsServer struct {
	ws       *websocket.Upgrader
	callback Callback
	clients  *sync.Map
}

func (s *wsServer) wsHttpHandle(w http.ResponseWriter, r *http.Request) {
	if conn, err := s.ws.Upgrade(w, r, nil); err == nil {
		c := &wsConn{conn: conn}
		c.setState(ConnStateConnected)
		s.clients.Store(c.Identity(), c)
		if s.callback != nil {
			s.callback.OnConnected(c)
		}
		go s.handleConnection(c, s.callback)
	} else {
		s.callback.OnError(err)
	}
}

func (s *wsServer) listen(port int, callback Callback) error {
	s.ws = &websocket.Upgrader{}
	s.clients = &sync.Map{}
	s.callback = callback
	http.HandleFunc("/", s.wsHttpHandle)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		return err
	}
	return nil
}

func (s *wsServer) Close() error {
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
	s.ws = nil
	return nil
}

func (s *wsServer) handleConnection(conn Conn, callback Callback) {
	var buf []byte
	defer func() {
		if err := conn.Close(); err != nil {
			if callback != nil {
				callback.OnError(err)
			}
		}
	}()
	for {
		_, err := conn.read(&buf)
		if err != nil {
			conn.setState(ConnStateClosed)
			if callback != nil {
				callback.OnError(err)
				callback.OnDisconnected(conn)
			}
			s.clients.Delete(conn.Identity())
			break
		}
		if callback != nil {
			callback.OnMessage(conn, buf)
		}
	}
}

func (s *wsServer) GetConnection(identity uint32) (Conn, bool) {
	if s.clients == nil {
		return nil, false
	} else {
		if conn, ok := s.clients.Load(identity); ok {
			return conn.(Conn), ok
		} else {
			return nil, false
		}
	}
}

type wsClient struct {
	serverAddr string
	callback   Callback
	conn       Conn
}

func (c *wsClient) Reconnect() error {
	return c.connect(c.serverAddr, c.callback)
}

func (c *wsClient) connect(serverAddr string, callback Callback) error {
	c.serverAddr = serverAddr
	c.callback = callback
	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		return err
	}
	c.conn = Conn(&wsConn{conn: conn})
	c.conn.setState(ConnStateConnected)
	if callback != nil {
		callback.OnConnected(c.conn)
	}
	c.handleConnection(c.conn, callback)
	return nil
}

func (c *wsClient) handleConnection(conn Conn, callback Callback) {
	var buf []byte
	defer func() {
		if err := conn.Close(); err != nil {
			if callback != nil {
				callback.OnError(err)
			}
		}
	}()
	for {
		_, err := conn.read(&buf)
		if err != nil {
			conn.setState(ConnStateClosed)
			if callback != nil {
				callback.OnError(err)
				callback.OnDisconnected(conn)
			}
			c.conn = nil
			break
		}
		if callback != nil {
			callback.OnMessage(conn, buf)
		}
	}
}

func (c *wsClient) Send(msg []byte) error {
	if c.conn != nil && c.conn.State() == ConnStateConnected {
		return c.conn.Send(msg)
	} else {
		return ConnectionError{"Send failed: connection was not built"}
	}
}

func (c *wsClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
		c.conn = nil
	}
	return nil
}

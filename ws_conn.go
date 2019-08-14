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
)

type wsConn struct {
	baseConn
	conn *websocket.Conn
}

func (c *wsConn) Send(msg []byte) error {
	if c.conn != nil {
		if msg == nil || len(msg) == 0 {
			return EmptyMessageError{}
		}
		err := c.conn.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			return err
		}
		return nil
	} else {
		return ConnectionError{"Send failed, connection was not built"}
	}
}

func (c *wsConn) read(buf *[]byte) (int, error) {
	if c.conn != nil {
		t, msg, err := c.conn.ReadMessage()
		if err == nil && t == websocket.BinaryMessage {
			*buf = msg
			return len(msg), nil
		}
	}
	return -1, nil
}

func (c *wsConn) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
		c.conn = nil
	}
	return nil
}

func (c *wsConn) RemoteAddr() string {
	if c.conn != nil {
		return c.conn.RemoteAddr().String()
	}
	return "0:0:0:0"
}

func (c *wsConn) LocalAddr() string {
	if c.conn != nil {
		return c.conn.LocalAddr().String()
	}
	return "0:0:0:0"
}

func (c *wsConn) NetProtocol() Protocol {
	return WebSocket
}

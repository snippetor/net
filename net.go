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

// 网络协议定义
type Protocol int

const (
	Tcp Protocol = iota
	WebSocket
	Kcp
)

// 消息回调
type Callback interface {
	OnMessage(Conn, []byte)
	OnConnected(Conn)
	OnDisconnected(Conn)
	OnError(error)
}

// 服务器接口
type Server interface {
	listen(int, Callback) error
	GetConnection(uint32) (Conn, bool)
	Close() error
}

// 客户端接口
type Client interface {
	connect(string, Callback) error
	Send([]byte) error
	Close() error
	Reconnect() error
}

// 同步执行网络监听
func Listen(net Protocol, port int, callback Callback) (Server, error) {
	var server Server
	switch net {
	case Tcp:
		server = &tcpServer{}
	case WebSocket:
		server = &wsServer{}
	case Kcp:
		server = &kcpServer{}
	default:
		return nil, &UnknownNetTypeError{UnknownType: int(net)}
	}
	return server, server.listen(port, callback)
}

// 同步连接服务器
func Connect(net Protocol, serverAddr string, callback Callback) (Client, error) {
	var client Client
	switch net {
	case Tcp:
		client = &tcpClient{}
	case WebSocket:
		client = &wsClient{}
	case Kcp:
		client = &kcpClient{}
	default:
		return nil, &UnknownNetTypeError{UnknownType: int(net)}
	}
	return client, client.connect(serverAddr, callback)
}

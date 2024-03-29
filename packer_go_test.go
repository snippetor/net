package net

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
)

func TestMessagePacker(t *testing.T) {
	p := &messagePacker{}
	out := p.Pack(128, []byte("test_packer"))
	id, content, _ := p.Unpack(out)
	if id != 128 || string(content) != "test_packer" {
		fmt.Println(id, content)
		t.Fail()
	}
}

func TestMessagePacker1(t *testing.T) {
	p := &messagePacker{}
	out := p.Pack(-123, []byte("中文测试"))
	id, content, _ := p.Unpack(out)
	if id != -123 || string(content) != "中文测试" {
		fmt.Println(id, content)
		t.Fail()
	}
}

func TestMessagePacker2(t *testing.T) {
	p := &messagePacker{}
	out := p.Pack(-123, []byte(`{"a":"a", "b":1.1}`))
	id, content, _ := p.Unpack(out)
	if id != -123 || string(content) != `{"a":"a", "b":1.1}` {
		fmt.Println(id, content)
		t.Fail()
	}
}

func TestMessagePacker3(t *testing.T) {
	p := &messagePacker{}
	out := p.Pack(-123, []byte(`{"a":"a", "b":1.1}`))
	out = append(out, []byte("[append]")...)
	id, content, out := p.Unpack(out)
	if id != -123 || string(content) != `{"a":"a", "b":1.1}` || string(out) != "[append]" {
		fmt.Println(id, content)
		t.Fail()
	}
}

func BenchmarkDefaultMessagePacker_Pack(b *testing.B) {
	b.StopTimer()
	p := &messagePacker{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p.Pack(123, []byte(`{"a":"a", "b":1.1}`))
	}
}

func BenchmarkDefaultMessagePacker_Pack1(b *testing.B) {
	b.StopTimer()
	p := &messagePacker{}
	m := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		m[strconv.FormatInt(int64(i), 10)] = i
	}
	bytes, _ := json.Marshal(m)
	fmt.Println(len(bytes))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p.Pack(123, bytes)
	}
}

func BenchmarkDefaultMessagePacker_Unpack(b *testing.B) {
	b.StopTimer()
	p := &messagePacker{}
	out := p.Pack(-123, []byte(`{"a":"a", "b":1.1}`))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p.Unpack(out)
	}
}

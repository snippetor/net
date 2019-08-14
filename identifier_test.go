package net

import (
	"fmt"
	"testing"
)

func TestGenIdentity(t *testing.T) {
	i := NewIdentifier()

	//for k := 0; k < 100000; k++ {
	//	fmt.Println(i.GenIdentity())
	//}

	for k := 0; k < 10000; k++ {
		go func() {
			fmt.Println(i.GenIdentity())
		}()
	}
}

func BenchmarkGenIdentity(b *testing.B) {
	b.StopTimer()
	id := NewIdentifier()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		id.GenIdentity()
	}
}

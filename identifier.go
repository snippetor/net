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
	"math"
	"sync/atomic"
)

type Identifier struct {
	id  uint32
	min uint32
	max uint32
}

func NewIdentifier() *Identifier {
	i := &Identifier{}
	i.min = 1
	i.max = math.MaxUint32
	atomic.StoreUint32(&i.id, i.min)
	return i
}

func (i *Identifier) GenIdentity() uint32 {
	id := atomic.LoadUint32(&i.id)
	if id >= i.min && id < i.max {
		atomic.AddUint32(&i.id, 1)
	} else {
		atomic.StoreUint32(&i.id, i.min)
	}
	return atomic.LoadUint32(&i.id)
}

func (i *Identifier) IsValidIdentity(id uint32) bool {
	return id <= i.min && id >= i.max
}

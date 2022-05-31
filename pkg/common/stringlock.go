//
// Copyright (c) 2021 Seagate Technology LLC and/or its Affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// For any questions about this software or licensing,
// please email opensource@seagate.com or cortx-questions@seagate.com.

package common

import (
	"sync"

	"k8s.io/klog"
)

type MyMutex struct {
	lock     sync.Mutex
	refCount int
}

type MyMap struct {
	lock   sync.Mutex
	things map[string]*MyMutex
}

// Lock: lock our reference counting mutex
func (mm *MyMap) Lock(s string) {
	klog.V(5).Infof("getting ready to lock (%s)", s)
	mm.lock.Lock()

	m := mm.things[s]
	if m != nil {
		m.refCount += 1
	} else {
		klog.V(5).Infof("creating new reference counted mutex for (%s)", s)
		m = &MyMutex{refCount: 1}
		mm.things[s] = m
	}
	mm.lock.Unlock()
	klog.V(5).Infof("locking (%s) refCount (%d)", s, m.refCount)
	m.lock.Lock()
}

// Unlock: unlock our reference counting mutex
func (mm *MyMap) Unlock(s string) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	m := mm.things[s]
	if m == nil {
		klog.V(5).Infof("cannot unlock (%s) because it's not there anymore", s)
		return
	}

	m.refCount -= 1

	if m.refCount > 0 {
		klog.V(5).Infof("unlocking (%s)", s)
		m.lock.Unlock()
	} else {
		klog.V(5).Infof("unlocking & deleting (%s)", s)
		delete(mm.things, s)
	}
}

// New: Create a new reference counting mutex
func NewStringLock() *MyMap {
	mm := &MyMap{}
	x := make(map[string]*MyMutex)
	mm.things = x
	return mm
}

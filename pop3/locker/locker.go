package locker

import (
	. "sync"
)

var Locks map[string]*Mutex = make(map[string]*Mutex)

func Lock(user string) {
	if Locks[user] == nil {
		Locks[user] = &Mutex{}
	}
	Locks[user].Lock()
}

func Unlock(user string) {
	if Locks[user] != nil {
		//Locks[user].Unlock()
		Locks[user] = nil
	}
}

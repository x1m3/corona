package list

import (
	"sync"
)

// LIFO is a Last In First Out list
type LIFO struct {
	sync.Mutex
	next *node
}

func NewLIFO() *LIFO {
	return &LIFO{next: nil}
}

func (l *LIFO) Push(data interface{}) {
	l.Lock()
	i := &node{data: data, next: l.next}
	l.next = i
	l.Unlock()
}

func (l *LIFO) Pop() interface{} {
	l.Lock()
	defer l.Unlock()
	if l.next == nil {
		return nil
	}

	data := &l.next.data
	l.next = l.next.next
	return *data
}

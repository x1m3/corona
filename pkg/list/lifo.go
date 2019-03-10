package list

// LIFO is a Last In First Out list
type LIFO struct {
	next *node
}

func NewLIFO() *LIFO {
	return &LIFO{next: nil}
}

func (l *LIFO) Push(data interface{}) {
	i := &node{data: data, next: l.next}
	l.next = i
}

func (l *LIFO) Pop() interface{} {
	if l.next == nil {
		return nil
	}

	data := &l.next.data
	l.next = l.next.next
	return *data
}

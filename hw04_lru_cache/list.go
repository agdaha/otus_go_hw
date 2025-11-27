package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	ListItem := ListItem{
		Value: v,
		Next:  l.front,
		Prev:  nil,
	}

	if l.front != nil {
		l.front.Prev = &ListItem
	}
	if l.back == nil {
		l.back = &ListItem
	}

	l.front = &ListItem
	l.len++

	return &ListItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	ListItem := ListItem{
		Value: v,
		Next:  nil,
		Prev:  l.back,
	}

	if l.back != nil {
		l.back.Next = &ListItem
	}

	if l.front == nil {
		l.front = &ListItem
	}

	l.back = &ListItem
	l.len++
	return &ListItem
}

func (l *list) Remove(i *ListItem) {
	switch i {
	case l.front:
		l.front = i.Next
		i.Next.Prev = nil
	case l.back:
		l.back = i.Prev
		l.back.Next = nil
	default:
		i.Prev.Next = i.Next
		i.Next.Prev = i.Prev
	}

	i.Next = nil
	i.Prev = nil

	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == l.front {
		return
	}
	l.Remove(i)
	i.Next = l.front
	l.front.Prev = i
	l.front = i
	l.len++
}

func NewList() List {
	l := list{
		front: nil,
		back:  nil,
		len:   0,
	}
	return &l
}

package main

import (
	"context"
	"sync"
	"time"
)

type Cache struct {
	length   int
	capacity int

	items map[interface{}]*Node
	list  *DoublyLinkedList

	mu       sync.RWMutex
	contexts map[interface{}]context.CancelFunc
}

type Node struct {
	key, value interface{}
	prev, next *Node
}

type DoublyLinkedList struct {
	head, tail *Node
}

func NewCache(cap int) *Cache {

	return &Cache{
		capacity: cap,
		items:    make(map[interface{}]*Node),
		list:     &DoublyLinkedList{},
		contexts: make(map[interface{}]context.CancelFunc),
	}
}

func (cc *Cache) Cap() int {
	return cc.capacity
}

func (cc *Cache) Len() int {
	return cc.length
}

func (cc *Cache) Clear() {

	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.length = 0
	cc.items = make(map[interface{}]*Node)
	cc.list = &DoublyLinkedList{}

}

// Add метод доабвления
func (cc *Cache) Add(key, value interface{}) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if node, ok := cc.items[key]; ok {
		cc.updateExistingNode(node, value)
		return
	}

	if cc.length < cc.capacity {
		cc.addNewNode(key, value)
		return
	}

	cc.evictAndAddNewNode(key, value)
}

// AddWithTTL добавление с временным интервалом
func (cc *Cache) AddWithTTL(key, value interface{}, ttl time.Duration) {

	cc.Add(key, value)

	if cancel, exists := cc.contexts[key]; exists {
		cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	cc.contexts[key] = cancel

	go func() {
		<-ctx.Done()
		defer delete(cc.contexts, key)

		cc.Remove(key)
	}()

}

// Get получение
func (cc *Cache) Get(key interface{}) (value interface{}, ok bool) {

	cc.mu.RLock()
	defer cc.mu.RUnlock()

	node, ok := cc.items[key]
	if !ok {
		return nil, false
	}

	cc.relocateNodeToFront(node)

	return node.value, ok
}

// Remove удаление
func (cc *Cache) Remove(key interface{}) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	node, ok := cc.items[key]
	if !ok {
		return
	}

	delete(cc.items, key)
	cc.length--

	cc.linkNodes(node.prev, node.next)

	if node == cc.list.tail {
		cc.setTail(node.prev)
	}
}

// updateExistingNode связывает prev и next, обновляет value, перемещает в head
func (cc *Cache) updateExistingNode(node *Node, value interface{}) {

	cc.linkNodes(node.prev, node.next)
	node.value = value
	cc.moveNodeToFront(node)

	if node.next == nil {
		cc.setTail(node.prev)
	}
}

// relocateNodeToFront связывает prev и next, перемещает в head
func (cc *Cache) relocateNodeToFront(node *Node) {

	if node == cc.list.head {
		return
	}

	// Связка нод  между собой
	cc.linkNodes(node.prev, node.next)
	// Создание новой головы
	cc.moveNodeToFront(node)

}

// moveNodeToFront перемещает ноду в head
func (cc *Cache) moveNodeToFront(node *Node) {
	if cc.list.head != nil {
		cc.list.head.prev = node
	} else {
		cc.list.tail = node
	}
	node.next = cc.list.head
	cc.list.head = node
}

// addNewNode добавляет ноду, если емкость <N
func (cc *Cache) addNewNode(key, value interface{}) {
	newNode := &Node{key: key, value: value}
	cc.moveNodeToFront(newNode)
	cc.items[key] = newNode
	cc.length++
}

// evictAndAddNewNode удаляет ласт элемент
func (cc *Cache) evictAndAddNewNode(key, value interface{}) {
	tail := cc.list.tail
	cc.setTail(tail.prev)
	delete(cc.items, tail.key)
	cc.length--

	cc.addNewNode(key, value)
}

// linkNodes соединяет две ноды между собой
func (cc *Cache) linkNodes(leftNode, rightNode *Node) {
	cc.leftRightAndTail(leftNode, rightNode)
	cc.rightLeftAndHead(leftNode, rightNode)
}

// leftRightAndTail изменяет правый prev, обновляет хвост.
func (cc *Cache) leftRightAndTail(leftNode, rightNode *Node) {
	// Соединение левого узла с правым
	if rightNode != nil {
		rightNode.prev = leftNode
	} else {
		cc.list.tail = leftNode // Если правый узел пуст, обновляем хвост
	}
}

// rightLeftAndHead изменяет левый next, обновляет голову
func (cc *Cache) rightLeftAndHead(leftNode, rightNode *Node) {
	// Соединение правого узла с левым
	if leftNode != nil {
		leftNode.next = rightNode
	} else {
		cc.list.head = rightNode // Если левый узел пуст, обновляем голову
	}
}

// setTail устанавливает новый хвост списка, разрывая связь с предыдущим узлом
func (cc *Cache) setTail(node *Node) {
	cc.linkNodes(node, nil)
}

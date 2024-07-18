package main

import (
	"testing"
	"time"
)

func TestCache_Add(t *testing.T) {
	tests := []struct {
		key, value interface{}
	}{
		{key: "stringKey", value: "stringValue"},
		{key: 123, value: 456},
		{key: 1.23, value: 4.56},
		{key: true, value: false},
		{key: struct{ field string }{field: "test"}, value: struct{ field int }{field: 123}},
	}

	cache := NewCache(10)

	for _, tt := range tests {
		cache.Add(tt.key, tt.value)
		node, exists := cache.items[tt.key]
		if !exists {
			t.Errorf("Expected key %v to be present in cache", tt.key)
		}
		if node.value != tt.value {
			t.Errorf("Expected value %v for key %v, but got %v", tt.value, tt.key, node.value)
		}
	}
}

func TestCache_AddWithEviction(t *testing.T) {
	cache := NewCache(2)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	if _, exists := cache.items["key1"]; !exists {
		t.Errorf("Expected key1 to be present in cache")
	}
	if _, exists := cache.items["key2"]; !exists {
		t.Errorf("Expected key2 to be present in cache")
	}

	// вытесняю
	cache.Add("key3", "value3")

	// Проверяю, что key1 был вытеснен, а key2 и key3 остались
	if _, exists := cache.items["key1"]; exists {
		t.Errorf("Expected key1 to be evicted from cache")
	}
	if _, exists := cache.items["key2"]; !exists {
		t.Errorf("Expected key2 to be present in cache")
	}
	if _, exists := cache.items["key3"]; !exists {
		t.Errorf("Expected key3 to be present in cache")
	}
}

func TestCache_Get(t *testing.T) {
	cache := NewCache(2)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	// Проверяем, что key1 и key2 корректно возвращаются
	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Errorf("Expected value1 for key1, got %v", value)
	}
	if value, ok := cache.Get("key2"); !ok || value != "value2" {
		t.Errorf("Expected value2 for key2, got %v", value)
	}

	// Добавляем key3 и проверяем, что key1 удален, а key2 и key3 присутствуют
	cache.Add("key3", "value3")

	if _, exists := cache.items["key1"]; exists {
		t.Errorf("Expected key1 to be evicted from cache")
	}
	if _, exists := cache.items["key2"]; !exists {
		t.Errorf("Expected key2 to be present in cache")
	}
	if _, exists := cache.items["key3"]; !exists {
		t.Errorf("Expected key3 to be present in cache")
	}
}

func TestCache_Remove(t *testing.T) {
	cache := NewCache(2)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	cache.Remove("key1")

	// Проверяем, что key1 был удален
	if _, exists := cache.items["key1"]; exists {
		t.Errorf("Expected key1 to be removed from cache")
	}

	// Проверяем, что key2 все еще на месте
	if value, ok := cache.Get("key2"); !ok || value != "value2" {
		t.Errorf("Expected value2 for key2, got %v", value)
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(2)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	cache.Clear()

	// Проверяем, что все элементы удалены
	if len(cache.items) != 0 {
		t.Errorf("Expected cache to be empty, got %d items", len(cache.items))
	}
	if cache.Len() != 0 {
		t.Errorf("Expected cache length to be 0, got %d", cache.Len())
	}
}

func TestCache_AddWithTTL(t *testing.T) {
	cache := NewCache(2)

	cache.AddWithTTL("key1", "value1", 1*time.Second)

	// Проверяем, что элемент добавлен
	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Errorf("Expected value1 for key1, got %v", value)
	}

	// Ждем истечения TTL
	time.Sleep(2 * time.Second)

	// Проверяем, что элемент удален после TTL
	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be removed from cache after TTL")
	}
}

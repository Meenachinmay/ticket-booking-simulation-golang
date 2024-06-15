package main

import (
	"sync"
)

var TempSlice = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

func InsertInSliceAtIndex(slice []int, index int, value int) []int {
	return append(slice[:index], append([]int{value}, slice[:index]...)...)
}

func Remove(slice []int, index int) []int {
	return append(slice[:index], slice[index+1:]...)
}

func Insert(slice []int, index int, value int) []int {
	return append(slice[:index], append([]int{value}, slice[:index]...)...)
}

type KeyValue struct {
	Key, Value interface{}
}

func RangeOverSyncMap() []string {
	var sm sync.Map
	sm.Store("a", 1)

	var keys []string
	sm.Range(func(k, v interface{}) bool {
		key, ok := k.(string)
		if ok {
			keys = append(keys, key)
		}
		return true
	})

	return keys
}

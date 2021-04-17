package test

import (
	"fmt"
	"testing"
)

func TestPointerMap(t *testing.T) {
	fmt.Println("=================================== PointerMap test begin")
	a := int(1)
	b := int(2)
	c := int(3)

	pointerMap := make(map[*int]string)
	pointerMap[&a] = "user1"
	pointerMap[&b] = "user2"
	pointerMap[&c] = "user3"

	stringMap := make(map[string]*int)
	stringMap["user1"] = &a
	stringMap["user2"] = &b
	stringMap["user3"] = &b

	for k, v := range pointerMap {
		fmt.Printf("key: %v, val: %s\n", k, v)
	}

	searchName := "user2"
	tmpKey := stringMap[searchName]
	findName := pointerMap[tmpKey]
	fmt.Printf("search %s, find %s\n", searchName, findName)
}

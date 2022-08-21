package skiplist

import (
	"log"
	"testing"
)

func TestA(t *testing.T) {
	var s skipList
	a := s.randomHeight()
	log.Println("----------------", a)
	log.Println("----------------", FastRand())

}

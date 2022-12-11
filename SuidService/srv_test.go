package SuidService

import (
	"log"
	"testing"
)

func Test_suid_GenerateId(t *testing.T) {
	testFn := NewSuid()
	id := testFn.GenerateId("bayo", 6)
	log.Println(id)
}

package tests

import (
	"ichat-go/utils/strs"
	"testing"
)

func TestSubstring(t *testing.T) {
	s := "哈哈哈哈哈"
	if s[:3] == "哈哈哈" {
		t.Errorf("should not equal")
	}
	if strs.TakeFirstN(s, 3) != "哈哈哈" {
		t.Errorf("TakeFitstN not working")
	}
}

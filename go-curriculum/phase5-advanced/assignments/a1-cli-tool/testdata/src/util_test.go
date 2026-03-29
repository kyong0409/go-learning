// testdata/src/util_test.go - 테스트용 테스트 파일
package src

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("Add(1,2) != 3")
	}
}

func TestSubtract(t *testing.T) {
	if Subtract(5, 3) != 2 {
		t.Error("Subtract(5,3) != 2")
	}
}

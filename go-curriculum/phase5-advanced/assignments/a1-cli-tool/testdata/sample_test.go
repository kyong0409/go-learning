// testdata/sample_test.go - 테스트용 샘플 테스트 파일
package testdata

import "testing"

func TestSampleFunc(t *testing.T) {
	if SampleFunc() != "hello" {
		t.Error("기대값 불일치")
	}
}

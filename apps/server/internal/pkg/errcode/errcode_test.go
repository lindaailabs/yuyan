package errcode

import (
	"testing"
)

func TestNewValidSegments(t *testing.T) {
	cases := []Code{0, 1001, 1500, 1999, 2000, 2500, 2999, 5000, 5500, 5999}
	for _, c := range cases {
		if e := New(c, "x"); e.Code != c {
			t.Errorf("New(%d) code = %d", c, e.Code)
		}
	}
}

func TestNewInvalidSegmentPanics(t *testing.T) {
	cases := []Code{-1, 3000, 3500, 4999, 6000, 9999}
	for _, c := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("New(%d) should panic on invalid segment", c)
				}
			}()
			_ = New(c, "x")
		}()
	}
}

func TestSegmentOf(t *testing.T) {
	cases := map[Code]int{
		0: 0, 1001: 1000, 1999: 1000,
		2000: 2000, 2999: 2000,
		5000: 5000, 5999: 5000,
		3000: -1, -5: -1,
	}
	for code, want := range cases {
		if got := SegmentOf(code); got != want {
			t.Errorf("SegmentOf(%d) = %d, want %d", code, got, want)
		}
	}
}

func TestErrorString(t *testing.T) {
	e := New(ErrInvalidParam, "参数错误")
	if e.Error() != "[1001] 参数错误" {
		t.Errorf("Error() = %q", e.Error())
	}
}

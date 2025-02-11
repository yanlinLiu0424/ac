package ac

import (
	"testing"
)

func TestAcCs(t *testing.T) {
	v := "abcdefghjiklmnopqrstuvwxyz"
	ac := NewAcAlgorithm()
	ac.AddCS("abcd")
	ac.AddCS("bcde")
	ac.AddCS("bcd")
	ac.Build()
	r := ac.Search(v)
	if len(r) != 3 {
		t.Fail()
	}
}
func TestAcCi(t *testing.T) {
	v := "ABCDEfghjiklmnopqrstuvwxyz"
	ac := NewAcAlgorithm()
	ac.AddCI("abcd")
	ac.AddCI("bcde")
	ac.AddCI("bcd")
	ac.Build()
	r := ac.Search(v)
	if len(r) != 3 {
		t.Fail()
	}
}

func TestResult(t *testing.T) {
	v := "ABCDEfghjiklmnopqrstuvwxyz"
	ac := NewAcAlgorithm()
	ac.AddCI("abcd")
	ac.AddCI("bcde")
	ac.AddCI("bcd")
	ac.Build()
	r := ac.SearchResult(v)
	if len(r) != 3 {
		t.Fail()
	}
}

func TestMemcmp(t *testing.T) {
	a := "123456"
	b := "123"
	r := memcmp(a, b, 3)
	if !r {
		t.Fatal("should be true")
	}
	r = memcmp(a, b, 4)
	if r {
		t.Fatal("should be false")
	}
}

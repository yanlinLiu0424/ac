package ac

import (
	"testing"
)

func TestAcCs(t *testing.T) {
	v := "abcdefghjiklmnopqrstuvwxyz"
	ac := NewAcAlgorithm()
	ac.AddCS(0, "abcd")
	ac.AddCS(1, "bcde")
	ac.AddCS(2, "bcd")
	ac.Build()
	r := ac.Search(v)
	if len(r) != 3 {
		t.Fail()
	}
}
func TestAcCi(t *testing.T) {
	v := "ABCDEfghjiklmnopqrstuvwxyz"
	ac := NewAcAlgorithm()
	ac.AddCI(0, "abcd")
	ac.AddCI(1, "bcde")
	ac.AddCI(2, "bcd")
	ac.Build()
	r := ac.Search(v)
	if len(r) != 3 {
		t.Fail()
	}
}

func TestResult(t *testing.T) {
	v := "sdfABCDEfghjiklmnopqrstuvwxyz"
	ac := NewAcAlgorithm()
	ac.AddCI(0, "abcd")
	ac.AddCI(1, "bcde")
	ac.AddCI(2, "bcd")
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

package ac

import (
	"reflect"
	"testing"
)

func TestACKS_Search_Basic(t *testing.T) {
	ac := NewACKS()
	ac.AddPattern([]byte("he"), 1, 0)
	ac.AddPattern([]byte("she"), 2, 0)
	//ac.AddPattern([]byte("his"), 3, 0)
	//ac.AddPattern([]byte("hers"), 4, 0)
	ac.Compile()

	text := []byte("ushers")
	// u s h e r s
	// 0 1 2 3 4 5
	// "she" ends at 3
	// "he" ends at 3
	// "hers" ends at 5
	matches := ac.Search(text)

	expected := map[uint][]int{
		1: {3}, // he
		2: {3}, // she
		4: {5}, // hers
	}

	if !reflect.DeepEqual(matches, expected) {
		t.Errorf("Expected %v, got %v", expected, matches)
	}
}

func TestACKS_Search_Caseless(t *testing.T) {
	ac := NewACKS()
	ac.AddPattern([]byte("AbC"), 1, Caseless)
	ac.Compile()

	text := []byte("abC")
	matches := ac.Search(text)

	expected := map[uint][]int{
		1: {2},
	}

	if !reflect.DeepEqual(matches, expected) {
		t.Errorf("Expected %v, got %v", expected, matches)
	}
}

func TestACKS_Search_CaseSensitive(t *testing.T) {
	ac := NewACKS()
	ac.AddPattern([]byte("abc"), 1, 0)
	ac.Compile()

	text := []byte("ABC")
	matches := ac.Search(text)

	if len(matches) != 0 {
		t.Errorf("Expected no matches, got %v", matches)
	}
}

func TestACKS_Search_MixedFlags(t *testing.T) {
	ac := NewACKS()
	ac.AddPattern([]byte("abc"), 1, Caseless)
	ac.AddPattern([]byte("XYZ"), 2, 0) // Case sensitive
	ac.Compile()

	text := []byte("ABC xyz XYZ")
	matches := ac.Search(text)

	expected := map[uint][]int{
		1: {2},  // ABC matches abc (Caseless)
		2: {10}, // XYZ matches XYZ (CaseSensitive)
	}

	if !reflect.DeepEqual(matches, expected) {
		t.Errorf("Expected %v, got %v", expected, matches)
	}
}

func TestACKS_Search_MultipleOccurrences(t *testing.T) {
	ac := NewACKS()
	ac.AddPattern([]byte("aba"), 1, 0)
	ac.Compile()

	text := []byte("abababa")
	matches := ac.Search(text)
	expected := map[uint][]int{
		1: {2, 4, 6},
	}

	if !reflect.DeepEqual(matches, expected) {
		t.Errorf("Expected %v, got %v", expected, matches)
	}
}

func TestACKS_Search_NoMatch(t *testing.T) {
	ac := NewACKS()
	ac.AddPattern([]byte("foo"), 1, 0)
	ac.Compile()

	text := []byte("bar")
	matches := ac.Search(text)

	if len(matches) != 0 {
		t.Errorf("Expected no matches, got %v", matches)
	}
}

func TestACKS_Search_OverlappingPatterns(t *testing.T) {
	ac := NewACKS()
	ac.AddPattern([]byte("nan"), 1, 0)
	ac.AddPattern([]byte("ana"), 2, 0)
	ac.Compile()

	text := []byte("banana")
	matches := ac.Search(text)
	expected := map[uint][]int{
		1: {4},
		2: {3, 5},
	}

	if !reflect.DeepEqual(matches, expected) {
		t.Errorf("Expected %v, got %v", expected, matches)
	}
}

func BenchmarkACKS_Search(b *testing.B) {
	ac := NewACKS()
	numPatterns := 1000
	for i := 0; i < numPatterns; i++ {
		s := randomString(10)
		ac.AddPattern([]byte(s), uint(i), Caseless)
	}
	ac.Compile()

	text := []byte(randomString(10000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ac.Search(text)
	}
}

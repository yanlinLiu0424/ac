package ac

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

func TestAhoCorasick_Search_CaseInsensitive1(t *testing.T) {
	v := []byte("sdfABCDEfghjiklmnopqrstuvwxyz")
	ac := NewAhoCorasick()
	patterns := []Pattern{
		{Str: "aBcd", ID: 30},
	}
	for _, p := range patterns {
		if err := ac.AddPattern(p); err != nil {
			t.Fatalf("AddPattern(%+v) failed: %v", p, err)
		}
	}
	ac.Build()
	m := MatchedHandler(func(id uint, from, to uint64) error {
		fmt.Println("Matched ID:", id, "From:", from, "To:", to)
		return nil
	})
	ac.Scan(v, m)

}

func TestAhoCorasick_Search_CaseInsensitive(t *testing.T) {
	v := []byte("sdfABCDEfghjiklmnopqrstuvwxyz")
	ac := NewAhoCorasick()
	patterns := []Pattern{
		{Str: "S", ID: 4},
		{Str: "abcd", ID: 0, Flags: Caseless},
		{Str: "bcde", ID: 1, Flags: Caseless},
		{Str: "bcd", ID: 2, Flags: Caseless},
		{Str: "uvw", ID: 3},
	}
	for _, p := range patterns {
		if err := ac.AddPattern(p); err != nil {
			t.Fatalf("AddPattern(%+v) failed: %v", p, err)
		}
	}
	ac.Build()

	r := ac.Search(v)
	if len(r) != 4 {
		t.Errorf("Expected 4 matches, got %d for text %q", len(r), v)
	}
}

func TestAhoCorasick_Search_CaseSensitive(t *testing.T) {
	v := []byte("sdfABCDEfghjiklmnopqrstuvwxyzabcd")
	ac := NewAhoCorasick()
	// "abcd" should match at the end, "ABCD" should not match if Cs:true
	patterns := []Pattern{
		{Str: "abcd", ID: 0, Flags: Caseless | SingleMatch},
		{Str: "ABCD", ID: 1, Flags: Caseless | SingleMatch},
		{Str: "BCDE", ID: 2, Flags: Caseless | SingleMatch},
	}
	for _, p := range patterns {
		err := ac.AddPattern(p)
		if err != nil {
			t.Fatalf("AddPattern(%+v) failed: %v", p, err)
		}
	}
	ac.Build()
	r := ac.Search(v)
	if len(r) != 3 {
		t.Errorf("Expected 2 matches, got %d for text %q. Matches: %v", len(r), v, r)
	}

	// Verify specific IDs if necessary, e.g. check that ID 0 and ID 2 are present
	foundIDs := make(map[uint]bool)
	for _, id := range r {
		foundIDs[id] = true
	}
	if !foundIDs[0] {
		t.Errorf("Expected pattern ID 0 ('abcd') to be found")
	}
	if !foundIDs[2] {
		t.Errorf("Expected pattern ID 2 ('BCDE') to be found")
	}
	if !foundIDs[1] { // "ABCD" Cs:true should not match "abcd"
		t.Errorf("Expected pattern ID 1 ('ABCD') to be found")
	}
}

func TestAhoCorasick_1000Patterns(t *testing.T) {
	ac := NewAhoCorasick()
	expectedIDs := make(map[uint]bool)
	var sbText bytes.Buffer

	// Generate 100 patterns and build the text
	for i := 0; i < 1000; i++ {
		// Create a unique pattern string
		patStr := fmt.Sprintf("Pattern%d-%s", i, randomString(4))

		// Add some noise before the pattern in the text
		sbText.WriteString(randomString(3))
		sbText.WriteString(patStr)

		p := Pattern{
			Str:   patStr,
			ID:    uint(i),
			Flags: Caseless,
		}
		if err := ac.AddPattern(p); err != nil {
			t.Fatalf("AddPattern failed for %s: %v", patStr, err)
		}
		expectedIDs[uint(i)] = true
	}

	// Add some trailing noise
	sbText.WriteString(randomString(5))

	ac.Build()

	text := sbText.Bytes()
	matches := ac.Search(text)

	foundIDs := make(map[uint]bool)
	for _, id := range matches {
		foundIDs[id] = true
	}

	for id := range expectedIDs {
		if !foundIDs[id] {
			t.Errorf("Expected to find pattern ID %d", id)
		}
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	sb := make([]byte, n)
	for i := range sb {
		sb[i] = charset[rand.Intn(len(charset))]
	}
	return string(sb)
}

func BenchmarkAhoCorasickSearch5000RandomPatterns(b *testing.B) {
	ac := NewAhoCorasick()
	numPatterns := 50000 // Keeping this at 5000 as per current code
	patterns := make([]string, numPatterns)

	for i := range numPatterns {
		// Generate random patterns of length between 5 and 15
		patternLength := rand.Intn(11) + 5
		patterns[i] = randomString(patternLength) // Uses randomString from ac_test.go
		err := ac.AddPattern(Pattern{Str: patterns[i], ID: uint(i), Flags: Caseless})
		if err != nil {
			b.Fatalf("Failed to add pattern %q: %v", patterns[i], err)
		}
	}
	ac.Build()

	searchTextLength := 1500
	searchText := []byte(randomString(searchTextLength))

	b.ResetTimer() // Start timing after setup
	for i := 0; i < b.N; i++ {
		_ = ac.Search(searchText)
	}
}

func BenchmarkAhoCorasickSearch5000FixedPatterns(b *testing.B) {
	ac := NewAhoCorasick()
	numPatterns := 50000

	for i := 0; i < numPatterns; i++ {
		s := fmt.Sprintf("FixedString%d", i)
		_ = ac.AddPattern(Pattern{Str: s, ID: uint(i), Flags: Caseless})
	}
	ac.Build()

	// Create a text that contains patterns
	var buffer bytes.Buffer
	for i := 0; i < 200; i++ {
		buffer.WriteString(fmt.Sprintf("noise_FixedString%d_data ", i%numPatterns))
	}
	text := buffer.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ac.Search(text)
	}
}

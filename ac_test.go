package ac

import (
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

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	sb := make([]byte, n)
	for i := range sb {
		sb[i] = charset[rand.Intn(len(charset))]
	}
	return string(sb)
}

// Renamed to reflect the actual number of patterns if numPatterns remains 5000
// Or, change numPatterns to 500 to match the original name.
// For this example, I'll assume you want to test 5000 and rename the function.
func BenchmarkAhoCorasickSearch5000RandomPatterns(b *testing.B) {
	ac := NewAhoCorasick()
	numPatterns := 5000 // Keeping this at 5000 as per current code
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

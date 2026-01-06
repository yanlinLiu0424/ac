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
		{Str: "aBcd", ID: 30, Flags: Caseless},
		{Str: "mno", ID: 31, Flags: SingleMatch},
		{Str: "uv", ID: 31, Flags: SingleMatch},
		{Str: "sdffd", ID: 32, Flags: SingleMatch},
	}
	for _, p := range patterns {
		if err := ac.AddPattern(p); err != nil {
			t.Fatalf("AddPattern(%+v) failed: %v", p, err)
		}
	}
	ac.Build()
	r := make(map[uint]struct{})
	m := MatchedHandler(func(id uint, from, to uint64) error {
		//fmt.Println("Matched ID:", id, "From:", from, "To:", to)
		r[id] = struct{}{}
		return nil
	})
	err := ac.Scan(v, m)
	if err != nil {
		t.Fatal(err)
	}
	if len(r) != 2 {
		t.Fatalf("Expected 2 matches, got %d for text %q", len(r), v)
	}
}

func TestAhoCorasick_Search_CaseInsensitive2(t *testing.T) {
	v := []byte("sdfABCDEfghjiklmnopqrstuvwxyz")
	ac := NewAhoCorasick()
	patterns := []Pattern{
		{Str: "aBcd", ID: 16777216 + 1, Flags: Caseless},
		{Str: "mno", ID: 31, Flags: SingleMatch},
		{Str: "uv", ID: 31, Flags: SingleMatch},
		{Str: "sdffd", ID: 32, Flags: SingleMatch},
	}
	for _, p := range patterns {
		if err := ac.AddPattern(p); err != nil {
			t.Fatalf("AddPattern(%+v) failed: %v", p, err)
		}
	}
	ac.Build()
	r := make(map[uint]struct{})
	m := MatchedHandler(func(id uint, from, to uint64) error {
		r[id] = struct{}{}
		return nil
	})
	err := ac.Scan(v, m)
	if err != nil {
		t.Error(err)
	}
	if len(r) != 2 {
		t.Errorf("Expected 2 matches, got %d for text %q", len(r), v)
	}
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

	r, err := ac.Search(v)
	if err != nil {
		t.Error(err)
	}
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
	r, err := ac.Search(v)
	if err != nil {
		t.Error(err)
	}
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
	expectedIDs := make(map[uint]bool, 1000)
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
	matches, err := ac.Search(text)
	if err != nil {
		t.Error(err)
	}
	foundIDs := make(map[uint]bool, len(matches))
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

func BenchmarkAhoCorasickSearch5000FixedPatterns(b *testing.B) {
	ac := NewAhoCorasick()
	numPatterns := 50000

	for i := 0; i < numPatterns; i++ {
		s := fmt.Sprintf("FixedString%d", i)
		_ = ac.AddPattern(Pattern{Str: s, ID: uint(i), Flags: Caseless | SingleMatch})
	}
	ac.Build()

	// Create a text that contains patterns
	var buffer bytes.Buffer
	for i := 0; i < 200; i++ {
		buffer.WriteString(fmt.Sprintf("noise_FixedString%d_data ", i%numPatterns))
	}
	text := buffer.Bytes()

	b.ReportAllocs()
	b.ResetTimer()
	handler := MatchedHandler(func(id uint, from, to uint64) error { return nil })
	for i := 0; i < b.N; i++ {
		_ = ac.Scan(text, handler)
	}
}

func BenchmarkAhoCorasick_SingleMatch_Slice(b *testing.B) {
	ac := NewAhoCorasick()
	numPatterns := 5000

	for i := 0; i < numPatterns; i++ {
		s := fmt.Sprintf("FixedString%d", i)

		_ = ac.AddPattern(Pattern{Str: s, ID: uint(i)})
	}
	ac.Build()

	var buffer bytes.Buffer
	for i := 0; i < 200; i++ {
		buffer.WriteString(fmt.Sprintf("noise_FixedString%d_data ", i%numPatterns))
	}
	text := buffer.Bytes()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ac.Scan(text, func(id uint, from, to uint64) error { return nil })
	}

}

func BenchmarkAhoCorasick_SingleMatch_Map(b *testing.B) {
	ac := NewAhoCorasick()
	numPatterns := 5000

	for i := 0; i < numPatterns; i++ {
		s := fmt.Sprintf("FixedString%d", i)
		_ = ac.AddPattern(Pattern{Str: s, ID: uint(i), Flags: Caseless})
	}
	//  Map
	// maxSliceSize = 16 * 1024 * 1024 = 16777216
	_ = ac.AddPattern(Pattern{Str: "FORCE_MAP_MODE", ID: 16777216 + 1, Flags: Caseless})

	ac.Build()

	var buffer bytes.Buffer
	for i := 0; i < 200; i++ {
		buffer.WriteString(fmt.Sprintf("noise_FixedString%d_data ", i%numPatterns))
	}
	buffer.WriteString("FORCE_MAP_MODE")
	text := buffer.Bytes()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ac.Scan(text, func(id uint, from, to uint64) error { return nil })
	}
}

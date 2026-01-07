package ac

import (
	"bytes"
	"fmt"
	"unicode"
)

type MatchedHandler func(id uint, from, to uint64) error
type matchedPattern func(pos uint64, ps Pattern) error

const (
	fail = -1
)

type Flag uint

const (
	Caseless Flag = 1 << iota // Caseless represents set case-insensitive matching.
	SingleMatch
)

type Pattern struct {
	Str    string
	ID     uint // ID
	Flags  Flag // Caseless represents set case-insensitive matching.
	strlen int
}

type state struct {
	trans   []uint32 // Packed char (low 8 bits) and next state (high 24 bits)
	failure int32    // Failure transition
	output  []int32  // Matched patterns (indices)
	isData  bool     // Whether this is a data node
}

type AhoCorasick struct {
	states         []state
	patterns       []Pattern
	rootTrans      [128]int32 // Cache for root transitions to optimize hot path
	size           int        // Size of the automaton
	maxID          uint
	hasSingleMatch bool
}

func newState() state {
	state := state{
		trans:   nil,
		failure: fail,
		output:  []int32{},
		isData:  false,
	}
	return state
}

func (s *state) get(char byte) int32 {
	// Optimization: Use binary search for dense states (e.g. root) to avoid O(N) scan
	if len(s.trans) > 16 {
		l, r := 0, len(s.trans)
		for l < r {
			mid := (l + r) >> 1
			if byte(s.trans[mid]) < char {
				l = mid + 1
			} else {
				r = mid
			}
		}
		if l < len(s.trans) && byte(s.trans[l]) == char {
			return int32(s.trans[l] >> 8)
		}
		return fail
	}

	for _, t := range s.trans {
		if byte(t) == char {
			return int32(t >> 8)
		}
		// Since trans is sorted, we can exit early
		if byte(t) > char {
			break
		}
	}
	return fail
}

// acd
func (s *state) add(char byte, next int32) {
	packed := (uint32(next) << 8) | uint32(char)
	// Insert in sorted order to enable binary search/early exit
	l, r := 0, len(s.trans)
	for l < r {
		mid := (l + r) >> 1
		if byte(s.trans[mid]) < char {
			l = mid + 1
		} else {
			r = mid
		}
	}
	i := l
	s.trans = append(s.trans, 0)
	copy(s.trans[i+1:], s.trans[i:])
	s.trans[i] = packed
}

func NewAhoCorasick() *AhoCorasick {
	ac := &AhoCorasick{
		states:   []state{newState()},
		patterns: make([]Pattern, 0),
		maxID:    0,
	}
	for i := range ac.rootTrans {
		ac.rootTrans[i] = fail
	}
	return ac
}
func (ac *AhoCorasick) AddPattern(p Pattern) error {
	currentState := 0
	for _, char := range p.Str {
		char = unicode.ToLower(char)
		if char >= 128 {
			return fmt.Errorf("pattern contains non-ASCII character: %c", char)
		}
		next := ac.states[currentState].get(byte(char))
		if next == fail {
			newstate := len(ac.states)
			ac.states = append(ac.states, newState())
			ac.states[currentState].add(byte(char), int32(newstate))
			if currentState == 0 {
				ac.rootTrans[char] = int32(newstate)
			}
			next = int32(newstate)
		}
		currentState = int(next)
	}
	p.strlen = len(p.Str)
	ac.patterns = append(ac.patterns, p)
	pidx := int32(len(ac.patterns) - 1)
	ac.states[currentState].output = append(ac.states[currentState].output, pidx)
	ac.states[currentState].isData = true
	if p.ID > ac.maxID {
		ac.maxID = p.ID
	}
	if p.Flags&SingleMatch > 0 {
		ac.hasSingleMatch = true
	}
	ac.size++
	return nil
}

func (ac *AhoCorasick) Build() {
	// Optimization: Compact trans slices into a single contiguous array.
	// This improves cache locality by reducing pointer chasing to scattered heap memory.
	totalTrans := 0
	for i := range ac.states {
		totalTrans += len(ac.states[i].trans)
	}
	if totalTrans > 0 {
		allTrans := make([]uint32, 0, totalTrans)
		for i := range ac.states {
			if len(ac.states[i].trans) > 0 {
				start := len(allTrans)
				allTrans = append(allTrans, ac.states[i].trans...)
				// Use 3-index slicing to set capacity = length, ensuring safety
				ac.states[i].trans = allTrans[start:len(allTrans):len(allTrans)]
			}
		}
	}

	queue := []int{}

	for _, t := range ac.states[0].trans {
		state := int32(t >> 8)
		ac.states[int(state)].failure = 0
		queue = append(queue, int(state))
	}

	for len(queue) > 0 {
		currentState := queue[0]
		queue = queue[1:]

		for _, t := range ac.states[currentState].trans {
			char := byte(t)
			nextState32 := int32(t >> 8)

			nextState := int(nextState32)
			queue = append(queue, nextState)

			failState := int(ac.states[currentState].failure)
			for failState != fail && ac.states[failState].get(char) == fail {
				failState = int(ac.states[failState].failure)
			}

			if failState != fail {
				ac.states[nextState].failure = ac.states[failState].get(char)
			} else {
				ac.states[nextState].failure = 0
			}

			if ac.states[nextState].isData {
				ac.states[nextState].output = append(ac.states[nextState].output, ac.states[int(ac.states[nextState].failure)].output...)
			}

		}
	}

	// Optimization: Compact output slices at the end of Build.
	// Outputs are populated during Build, so we must do this after the queue processing.
	totalOutput := 0
	for i := range ac.states {
		totalOutput += len(ac.states[i].output)
	}
	if totalOutput > 0 {
		allOutput := make([]int32, 0, totalOutput)
		for i := range ac.states {
			if len(ac.states[i].output) > 0 {
				start := len(allOutput)
				allOutput = append(allOutput, ac.states[i].output...)
				ac.states[i].output = allOutput[start:len(allOutput):len(allOutput)]
			}
		}
	}
}

func (ac *AhoCorasick) searchPatterns(text []byte, matched matchedPattern) error {
	currentState := 0
	// A slice is faster if memory is acceptable. Let's use a 16MB threshold.
	// A bool is 1 byte, so maxID can be up to ~16M.
	const maxSliceSize = 16 * 1024 * 1024
	useSlice := (ac.maxID + 1) <= maxSliceSize
	if useSlice {
		// Fast path: use a bitset for duplicate checking.
		var record []uint64
		if ac.hasSingleMatch {
			record = make([]uint64, (ac.maxID/64)+1)
		}
		for k, char := range text {
			if char >= 128 {
				return fmt.Errorf("text contains non-ASCII character: %c", char)
			}
			char = toLower(char)
			var next int32
			for currentState != fail {
				if currentState == 0 {
					next = ac.rootTrans[char]
				} else {
					next = ac.states[currentState].get(char)
				}
				if next != fail {
					break
				}
				currentState = int(ac.states[currentState].failure)
			}

			if currentState == fail {
				currentState = 0
				continue
			}
			currentState = int(next)
			if ac.states[currentState].isData {
				for _, pidx := range ac.states[currentState].output {
					p := ac.patterns[pidx]
					if p.Flags&SingleMatch > 0 {
						idx := p.ID / 64
						mask := uint64(1) << (p.ID % 64)
						if record[idx]&mask != 0 {
							continue
						}
						record[idx] |= mask
					}

					if p.Flags&Caseless > 0 {
						err := matched(uint64(k+1), p)
						if err != nil {
							return err
						}
					} else {
						if memcmp([]byte(p.Str), text[k-p.strlen+1:], p.strlen) {
							err := matched(uint64(k+1), p)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	} else {
		// General path: use a map for sparse or large IDs.
		var record map[uint]struct{}
		if ac.hasSingleMatch {
			record = make(map[uint]struct{})
		}
		for k, char := range text {
			if char >= 128 {
				return fmt.Errorf("text contains non-ASCII character: %c", char)
			}
			char = toLower(char)
			var next int32
			for currentState != fail {
				if currentState == 0 {
					next = ac.rootTrans[char]
				} else {
					next = ac.states[currentState].get(char)
				}
				if next != fail {
					break
				}
				currentState = int(ac.states[currentState].failure)
			}

			if currentState == fail {
				currentState = 0
				continue
			}
			currentState = int(next)
			if ac.states[currentState].isData {
				for _, pidx := range ac.states[currentState].output {
					p := ac.patterns[pidx]
					if p.Flags&SingleMatch > 0 {
						if _, exists := record[p.ID]; exists {
							continue
						}
						record[p.ID] = struct{}{}
					}

					if p.Flags&Caseless > 0 {
						err := matched(uint64(k+1), p)
						if err != nil {
							return err
						}
					} else {
						if memcmp([]byte(p.Str), text[k-p.strlen+1:], p.strlen) {
							err := matched(uint64(k+1), p)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (ac *AhoCorasick) Search(text []byte) ([]uint, error) {
	matches := make([]uint, 0, ac.size)
	h := func(pos uint64, ps Pattern) error {
		matches = append(matches, ps.ID)
		return nil
	}
	err := ac.searchPatterns(text, h)
	if err != nil {
		return nil, err
	}
	return matches, nil
}
func (ac *AhoCorasick) Scan(text []byte, m MatchedHandler) error {
	h := func(pos uint64, ps Pattern) error {
		err := m(ps.ID, 0, pos)
		if err != nil {
			return err
		}
		return nil
	}
	err := ac.searchPatterns(text, h)
	if err != nil {
		return err
	}
	return nil
}

func memcmp(a, b []byte, l int) bool {
	if l > len(b) || l > len(a) {
		return false
	}
	return bytes.Equal(a[:l], b[:l])
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

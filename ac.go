package ac

import (
	"bytes"
	"unicode"
)

type MatchedHandler func(id uint, from, to uint64) error
type matchedPattern func(pos uint64, ps Pattern)

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
	transitions     [256]int  // Stores transitions using an array
	failure         int       // Failure transition
	output          []Pattern // Matched patterns
	caseInsensitive bool      // Whether case-insensitive matching is enabled
	isData          bool      // Whether this is a data node
}

type AhoCorasick struct {
	states []state
	size   int // Size of the automaton
}

func newState() state {
	state := state{
		transitions:     [256]int{},
		failure:         fail,
		output:          []Pattern{},
		caseInsensitive: false,
		isData:          false,
	}

	for i := range len(state.transitions) {
		state.transitions[i] = fail
	}
	return state
}

func NewAhoCorasick() *AhoCorasick {
	ac := &AhoCorasick{
		states: []state{newState()},
	}
	return ac
}
func (ac *AhoCorasick) AddPattern(p Pattern) error {
	currentState := 0
	for _, char := range p.Str {
		char = unicode.ToLower(char)
		if ac.states[currentState].transitions[char] == fail {
			newstate := len(ac.states)
			ac.states = append(ac.states, newState())
			ac.states[currentState].transitions[char] = newstate
		}
		currentState = ac.states[currentState].transitions[char]
	}
	p.strlen = len(p.Str)
	ac.states[currentState].output = append(ac.states[currentState].output, p)
	ac.states[currentState].isData = true
	ac.size++
	return nil
}

func (ac *AhoCorasick) Build() {
	queue := []int{}

	for _, state := range ac.states[0].transitions {
		if state != fail {
			ac.states[state].failure = 0
			queue = append(queue, state)
		}
	}

	for len(queue) > 0 {
		currentState := queue[0]
		queue = queue[1:]

		for char, nextState := range ac.states[currentState].transitions {
			if nextState == fail {
				continue
			}
			queue = append(queue, nextState)

			failState := ac.states[currentState].failure
			for failState != fail && ac.states[failState].transitions[char] == fail {
				failState = ac.states[failState].failure
			}

			if failState != fail {
				ac.states[nextState].failure = ac.states[failState].transitions[char]
			} else {
				ac.states[nextState].failure = 0
			}

			if ac.states[nextState].isData {
				ac.states[nextState].output = append(ac.states[nextState].output, ac.states[ac.states[nextState].failure].output...)
			}

		}
	}
}

func (ac *AhoCorasick) searchPatterns(text []byte, matched matchedPattern) {
	currentState := 0
	record := map[uint]struct{}{}
	for k, char := range text {
		char = byte(unicode.ToLower(rune(char)))
		for currentState != fail && ac.states[currentState].transitions[char] == fail {
			currentState = ac.states[currentState].failure
		}

		if currentState == fail {
			currentState = 0
			continue
		}
		currentState = ac.states[currentState].transitions[char]
		if ac.states[currentState].isData {
			for _, p := range ac.states[currentState].output {
				if p.Flags&SingleMatch > 0 && isExisted(record, p.ID) {
					continue
				}
				if p.Flags&Caseless > 0 {
					record[p.ID] = struct{}{}
					matched(uint64(k+1), p)
				} else {
					if memcmp([]byte(p.Str), text, p.strlen) {
						record[p.ID] = struct{}{}
						matched(uint64(k+1), p)
					}
				}
			}
		}
	}

}

func (ac *AhoCorasick) Search(text []byte) []uint {
	matches := make([]uint, 0, ac.size)
	h := matchedPattern(func(pos uint64, ps Pattern) {
		matches = append(matches, ps.ID)
	})
	ac.searchPatterns(text, h)
	return matches
}
func (ac *AhoCorasick) Scan(text []byte, m MatchedHandler) {
	h := matchedPattern(func(pos uint64, ps Pattern) {
		err := m(ps.ID, 0, pos)
		if err != nil {
			return
		}
	})
	ac.searchPatterns(text, h)
}

func isExisted(m map[uint]struct{}, id uint) bool {
	_, exists := m[id]
	return exists
}

func memcmp(a, b []byte, l int) bool {
	if l > len(b) || l > len(a) {
		return false
	}
	return bytes.Equal(a[:l], b[:l])
}

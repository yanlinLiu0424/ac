package ac

import (
	"strings"
	"unicode"
)

const (
	SCACFAIL = -1
)

type Flag uint

const (
	Caseless Flag = 1 << iota // Caseless represents set case-insensitive matching.
	SingleMatch
)

type Pattern struct {
	Str    string
	ID     uint64 // ID
	Flags  Flag   // Caseless represents set case-insensitive matching.
	strlen int    // 字符串長度
}

type state struct {
	transitions     [256]int  // 使用陣列儲存轉換
	failure         int       // 失敗轉移
	output          []Pattern // 匹配的模式
	caseInsensitive bool      // 是否區分大小寫
	isData          bool      // 是否為資料節點
}

type AhoCorasick struct {
	states []state
}

func newState() state {
	state := state{
		transitions:     [256]int{},
		failure:         SCACFAIL,
		output:          []Pattern{},
		caseInsensitive: false,
		isData:          false,
	}
	// 初始化所有轉換為失敗狀態
	for i := range len(state.transitions) {
		state.transitions[i] = SCACFAIL
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
		if ac.states[currentState].transitions[char] == SCACFAIL {
			newstate := len(ac.states)
			ac.states = append(ac.states, newState())
			ac.states[currentState].transitions[char] = newstate
		}
		currentState = ac.states[currentState].transitions[char]
	}
	p.strlen = len(p.Str)
	ac.states[currentState].output = append(ac.states[currentState].output, p)
	ac.states[currentState].isData = true
	return nil
}

func (ac *AhoCorasick) Build() {
	queue := []int{}

	// 初始化根節點的失敗轉移
	for _, state := range ac.states[0].transitions {
		if state != SCACFAIL {
			ac.states[state].failure = 0
			queue = append(queue, state)
		}
	}

	// BFS 構建失敗轉移
	for len(queue) > 0 {
		currentState := queue[0]
		queue = queue[1:]

		for char, nextState := range ac.states[currentState].transitions {
			if nextState == SCACFAIL {
				continue
			}
			queue = append(queue, nextState)

			// 計算失敗轉移
			failState := ac.states[currentState].failure
			for failState != SCACFAIL && ac.states[failState].transitions[char] == SCACFAIL {
				failState = ac.states[failState].failure
			}

			if failState != SCACFAIL {
				ac.states[nextState].failure = ac.states[failState].transitions[char]
			} else {
				ac.states[nextState].failure = 0
			}

			// 合併輸出
			if ac.states[nextState].isData {
				ac.states[nextState].output = append(ac.states[nextState].output, ac.states[ac.states[nextState].failure].output...)
			}

		}
	}
}

func (ac *AhoCorasick) Search(text string) []uint64 {
	currentState := 0
	matches := []uint64{}
	dup := map[uint64]struct{}{}
	for k, char := range text {
		char = unicode.ToLower(char)
		for currentState != SCACFAIL && ac.states[currentState].transitions[char] == SCACFAIL {
			currentState = ac.states[currentState].failure
		}

		if currentState == SCACFAIL {
			currentState = 0
			continue
		}

		// 進入下一狀態
		currentState = ac.states[currentState].transitions[char]
		if ac.states[currentState].isData {
			for _, pattern := range ac.states[currentState].output {
				offset := k - pattern.strlen + 1
				idMatched(&matches, pattern, dup, text[offset:])

			}
		}
	}
	return matches
}

func idMatched(state *[]uint64, p Pattern, dup map[uint64]struct{}, buffer string) {
	if p.Flags&SingleMatch > 0 {
		if _, ok := dup[p.ID]; ok {
			return
		}
	}
	if p.Flags&Caseless > 0 {
		*state = append(*state, p.ID)
		dup[p.ID] = struct{}{}
	} else {
		if memcmp(p.Str, buffer, p.strlen) {
			*state = append(*state, p.ID)
			dup[p.ID] = struct{}{}
		}
	}
}

func memcmp(a, b string, l int) bool {
	if l > len(b) || l > len(a) {
		return false
	}
	return strings.Compare(a[:l], b[:l]) == 0
}

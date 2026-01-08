package ac

// ACKS represents the Aho-Corasick Ken Steele matcher
type ACKS struct {
	patterns       []Pattern
	translateTable [256]uint8
	alphabetSize   int

	// stateTable is a flattened 2D array: stateTable[state * alphabetSize + char]
	stateTable []int32

	// outputTable stores pattern IDs for each state.
	// Using a slice of slices for O(1) access by state index.
	outputTable [][]uint

	// Map to quickly find pattern by ID during verification
	patternIdxMap map[uint]int

	stateCount int
}

func NewACKS() *ACKS {
	return &ACKS{
		outputTable:   make([][]uint, 0),
		patternIdxMap: make(map[uint]int),
	}
}

func (ac *ACKS) AddPattern(content []byte, id uint, flags Flag) {
	ac.patterns = append(ac.patterns, Pattern{
		ID:      id,
		Content: content,
		Flags:   flags,
	})
	ac.patternIdxMap[id] = len(ac.patterns) - 1
}

func (ac *ACKS) Compile() {
	ac.initTranslateTable()
	ac.buildStateMachine()
}

func (ac *ACKS) initTranslateTable() {
	var counts [256]int

	// 1. Count occurrences, merging uppercase to lowercase to compress alphabet
	for _, p := range ac.patterns {
		for _, b := range p.Content {
			counts[toLower(b)]++
		}
	}

	// 2. Build translation table
	ac.alphabetSize = 1 // 0 is reserved for unused chars
	for i := 0; i < 256; i++ {
		// Skip uppercase, they will be mapped to lowercase indices later
		if i >= 'A' && i <= 'Z' {
			continue
		}

		if counts[i] > 0 {
			ac.translateTable[i] = uint8(ac.alphabetSize)
			ac.alphabetSize++
		} else {
			ac.translateTable[i] = 0
		}
	}

	// 3. Map uppercase to the same index as lowercase
	for i := 'A'; i <= 'Z'; i++ {
		ac.translateTable[i] = ac.translateTable[i+32]
	}
}

func (ac *ACKS) buildStateMachine() {
	// Temporary Trie structure
	trie := make(map[int]map[uint8]int)
	ac.stateCount = 1 // State 0 is root

	// Initialize output table for state 0
	ac.outputTable = append(ac.outputTable, []uint{})

	// 1. Build Trie (Goto)
	for _, p := range ac.patterns {
		currentState := 0
		for _, b := range p.Content {
			// Use the compressed character code
			tc := ac.translateTable[toLower(b)]

			if trie[currentState] == nil {
				trie[currentState] = make(map[uint8]int)
			}

			if next, exists := trie[currentState][tc]; exists {
				currentState = next
			} else {
				newState := ac.stateCount
				ac.stateCount++
				trie[currentState][tc] = newState
				// Expand output table
				ac.outputTable = append(ac.outputTable, []uint{})
				currentState = newState
			}
		}
		ac.outputTable[currentState] = append(ac.outputTable[currentState], p.ID)
	}

	// 2. Build Failure Table
	failure := make([]int, ac.stateCount)
	queue := []int{}

	// Depth 1 failure links point to root (0)
	if rootTrans, ok := trie[0]; ok {
		for _, nextState := range rootTrans {
			queue = append(queue, nextState)
			failure[nextState] = 0
		}
	}

	// BFS
	for len(queue) > 0 {
		rState := queue[0]
		queue = queue[1:]

		if transitions, ok := trie[rState]; ok {
			for charCode, nextState := range transitions {
				queue = append(queue, nextState)
				fState := failure[rState]

				for {
					if trans, ok := trie[fState]; ok {
						if val, ok := trans[charCode]; ok {
							failure[nextState] = val
							break
						}
					}
					if fState == 0 {
						failure[nextState] = 0
						break
					}
					fState = failure[fState]
				}
				// Merge outputs
				ac.outputTable[nextState] = append(ac.outputTable[nextState], ac.outputTable[failure[nextState]]...)
			}
		}
	}

	// 3. Build Delta Table (State Table)
	ac.stateTable = make([]int32, ac.stateCount*ac.alphabetSize)

	for state := 0; state < ac.stateCount; state++ {
		for charIdx := 0; charIdx < ac.alphabetSize; charIdx++ {
			nextState := 0
			curr := state
			found := false

			// Find transition
			for {
				if trans, ok := trie[curr]; ok {
					if val, ok := trans[uint8(charIdx)]; ok {
						nextState = val
						found = true
						break
					}
				}
				if curr == 0 {
					break
				}
				curr = failure[curr]
			}

			if !found {
				nextState = 0
			}

			ac.stateTable[state*ac.alphabetSize+charIdx] = int32(nextState)
		}
	}
}

func (ac *ACKS) Search(text []byte) map[uint][]int {
	matches := make(map[uint][]int)
	currentState := 0

	for i, b := range text {
		tc := ac.translateTable[b]

		// O(1) transition
		idx := currentState*ac.alphabetSize + int(tc)
		if idx >= len(ac.stateTable) {
			currentState = 0
		} else {
			currentState = int(ac.stateTable[idx])
		}

		// Check outputs
		if len(ac.outputTable[currentState]) > 0 {
			for _, pid := range ac.outputTable[currentState] {
				// Verification step
				idx, ok := ac.patternIdxMap[pid]
				if !ok {
					continue
				}
				pat := &ac.patterns[idx]

				patLen := len(pat.Content)
				startIdx := i - patLen + 1

				if startIdx < 0 {
					continue
				}

				candidate := text[startIdx : i+1]

				matched := false
				if pat.Flags&Caseless > 0 {
					// Case-insensitive comparison
					if stringEqualFold(candidate, pat.Content) {
						matched = true
					}
				} else {
					// Exact match
					if string(candidate) == string(pat.Content) {
						matched = true
					}
				}

				if matched {
					matches[pid] = append(matches[pid], i)
				}
			}
		}
	}
	return matches
}

func stringEqualFold(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if toLower(a[i]) != toLower(b[i]) {
			return false
		}
	}
	return true
}

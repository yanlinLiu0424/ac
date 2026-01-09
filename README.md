# AC (Aho-Corasick)

Package `ac` provides high-performance implementations of the Aho-Corasick string matching algorithm in Go. It is designed for efficient multi-pattern searching with support for ASCII text.

The package includes two variants:

### 1. Standard Aho-Corasick (`AhoCorasick`)
*   **Mechanism**: Uses a classic Trie structure with failure links. Transitions are handled via sparse lookups (e.g., binary search).
*   **Pros**: Memory-efficient and fast build time.
*   **Cons**: Search involves traversing failure links, which can be slower than a DFA approach.
*   **Best For**: Large pattern sets where memory footprint is a constraint.

### 2. Ken Steele Variant (`ACKS`)
*   **Mechanism**: Flattens the state machine into a dense Deterministic Finite Automaton (DFA) table. It pre-calculates the next state for every possible input character.
*   **Optimization**: Uses **Alphabet Compression** to map only used ASCII characters to dense indices, reducing the table size.
*   **Pros**: Extremely fast, deterministic **O(1)** search speed per input byte.
*   **Cons**: Higher memory usage and slower build time due to the dense table construction.
*   **Best For**: High-throughput applications where search speed is the priority.

## Features

*   **Case-Insensitive Matching**: Support for ASCII case-insensitive matching via the `Caseless` flag.
*   **Single Match Mode**: Option to report a pattern ID only the first time it is found using the `SingleMatch` flag.
*   **Zero-Allocation Scan**: The `Scan` method allows processing matches via a callback handler, avoiding memory allocations for result slices.
*   **ASCII Optimized**: Explicitly checks and optimizes for ASCII input (returns error for non-ASCII bytes).

## Usage

### Basic Search

```go
package main

import (
	"fmt"
	"log"

	"github.com/yanlinLiu0424/ac" // Replace with your actual import path
)

func main() {
	// 1. Initialize the matcher
	// Use NewACKS() for the DFA variant (faster search)
	// or NewAhoCorasick() for the standard variant.
	matcher := ac.NewACKS()

	// 2. Add patterns
	// IDs must be non-zero.
	_ = matcher.AddPattern(ac.Pattern{
		Content: []byte("he"),
		ID:      1,
	})
	_ = matcher.AddPattern(ac.Pattern{
		Content: []byte("she"),
		ID:      2,
	})
	_ = matcher.AddPattern(ac.Pattern{
		Content: []byte("HIS"),
		ID:      3,
		Flags:   ac.Caseless, // Case-insensitive match
	})

	// 3. Build the automaton
	matcher.Build()

	// 4. Search in text
	text := []byte("ushers his")

	// Option A: Get all match IDs
	matches, err := matcher.Search(text)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Matches:", matches)

	// Option B: Scan with callback (Zero allocation)
	err = matcher.Scan(text, func(id uint, from, to uint64) error {
		fmt.Printf("Pattern %d found ending at %d\n", id, to)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

## Performance

Benchmarks are included in the test files. `ACKS` generally outperforms `AhoCorasick` in search throughput due to its branch-free state transition logic, making it ideal for read-heavy workloads. `AhoCorasick` is a robust alternative when memory usage or build time is a primary concern.
# Aho-Corasick (AC) Library

This is an efficient Go implementation of the Aho-Corasick string matching algorithm. It supports multi-pattern matching and provides features such as case-insensitive matching and single-match mode for various use cases.

## Features

- **Multi-pattern Matching**: Search for multiple pattern strings in a single pass.
- **Case-insensitive**: Supports case-insensitive matching by setting the `Caseless` flag.
- **Single Match**: Supports reporting only one match per pattern ID within the same text by setting the `SingleMatch` flag (deduplication).
- **Automatic Performance Optimization**: Automatically selects between Slice (Bitset) or Map to track matching states based on the range of Pattern IDs, achieving optimal balance between performance and memory.

## Installation

```bash
go get github.com/yanlinLiu0424/ac
```

## Usage

### 1. Basic Search

The `Search` method returns a list of all matched Pattern IDs.

```go
package main

import (
	"fmt"
	"log"
	"github.com/yanlinLiu0424/ac"
)

func main() {
	// 1. Create AC instance
	acInstance := ac.NewAhoCorasick()

	// 2. Add Patterns
	// Supports setting ID and Flags
	acInstance.AddPattern(ac.Pattern{Str: "apple", ID: 1, Flags: ac.Caseless})
	acInstance.AddPattern(ac.Pattern{Str: "banana", ID: 2})

	// 3. Build Automaton
	acInstance.Build()

	// 4. Search
	text := []byte("I like Apple and banana")
	matches, err := acInstance.Search(text)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Matched IDs:", matches)
	// Output might be: [1 2] (Order depends on match position)
}
```

### 2. Scan Details

The `Scan` method allows processing each match event via a callback function, suitable for streaming processing or retrieving match positions.

```go
err := acInstance.Scan(text, func(id uint, from, to uint64) error {
	fmt.Printf("Found pattern ID %d ending at position %d\n", id, to)
	return nil
})
```

## Notes

- **Character Set Limitation**: The current implementation uses a `[128]int` array to store state transitions, so it strictly supports **ASCII characters** (0-127). If a Pattern or text contains non-ASCII characters, the methods will return an error.
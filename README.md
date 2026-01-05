# ac

Package `ac` implements the Aho-Corasick string searching algorithm in Go. It is designed for efficient multiple pattern matching.

## Installation

```bash
go get github.com/yanlinLiu0424/ac
```

## Features

- **Multiple Pattern Matching**: Efficiently finds multiple patterns in an input text.
- **Case Insensitivity**: Supports case-insensitive matching using the `Caseless` flag.
- **Custom IDs**: Allows assigning unique integer IDs to patterns for easy identification.
- **Flexible Search**:
    - `Search`: Returns a list of matched pattern IDs.
    - `Scan`: Provides a callback mechanism (`MatchedHandler`) to process matches with position details.

## Usage Example

```go
ac := NewAhoCorasick()
ac.AddPattern(Pattern{
    Str:   "keyword",
    ID:    1,
    Flags: Caseless,
})
ac.Build()

matches := ac.Search([]byte("some text with Keyword"))
// matches contains [1]
```

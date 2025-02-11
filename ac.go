package ac

import (
	"strings"
	"unicode"
)

type Result struct {
	Str    string
	offset uint16
	End    uint16
}

type values struct {
	strlen int
	str    string
	cs     bool
}
type node struct {
	next   map[rune]*node
	fail   *node
	values []values
}

func newNode() *node {
	return &node{next: map[rune]*node{}}
}

func NewAcAlgorithm() *AcAlgorithm {
	return &AcAlgorithm{root: newNode()}
}

type AcAlgorithm struct {
	root  *node
	count uint64 //numbers of string
	nodes uint64 //number of nodes
}

// case sensitive
func (a *AcAlgorithm) AddCS(v string) {
	cs := true
	a.add(v, cs)
}

// case insensitive
func (a *AcAlgorithm) AddCI(v string) {
	cs := false
	a.add(v, cs)
}

func (a *AcAlgorithm) add(str string, cs bool) {
	cur := a.root
	for _, c := range str {
		if _, ok := cur.next[c]; !ok {
			cur.next[c] = newNode()
			a.nodes++
		}
		cur = cur.next[c]
	}
	r := values{str: str, cs: cs, strlen: len(str)}
	cur.values = append(cur.values, r)
	a.count++
}

func (a *AcAlgorithm) Build() error {
	queue := make([]*node, 0, a.nodes)
	queue = append(queue, a.root)
	for len(queue) != 0 {
		parent := queue[0]
		queue = queue[1:]
		for k, child := range parent.next {
			if parent == a.root {
				child.fail = a.root
			} else {
				fail := parent.fail
				for fail != nil {
					if next, exists := fail.next[k]; exists {
						child.fail = next
						break
					}
					fail = fail.fail
				}
				if fail == nil {
					child.fail = a.root
				}
				child.values = append(child.values, child.fail.values...)
			}
			queue = append(queue, child)
		}
	}
	return nil
}

func (a *AcAlgorithm) Search(str string) []string {
	r := make([]string, 0, a.count)
	current := a.root
	for k, c := range str {
		c = unicode.ToLower(c)
		for current != a.root && current.next[c] == nil {
			current = current.fail
		}
		if next, exists := current.next[c]; exists {
			current = next
			for _, v := range current.values {
				if v.cs {
					offset := k - v.strlen + 1
					if memcmp(v.str, str[offset:], v.strlen) {
						r = append(r, v.str)
					}
				} else {
					r = append(r, v.str)
				}
			}
		}
	}
	return r
}

func (a *AcAlgorithm) SearchResult(str string) []Result {
	r := make([]Result, 0, a.count)
	current := a.root
	for k, c := range str {
		c = unicode.ToLower(c)
		for current != a.root && current.next[c] == nil {
			current = current.fail
		}
		if next, exists := current.next[c]; exists {
			current = next
			for _, v := range current.values {
				offset := k - v.strlen + 1
				if v.cs {
					if memcmp(v.str, str[offset:], v.strlen) {
						r = append(r, Result{offset: uint16(offset), Str: v.str, End: uint16(k)})
					}
				} else {
					r = append(r, Result{offset: uint16(offset), Str: v.str, End: uint16(k)})
				}
			}
		}
	}
	return r
}

func memcmp(a, b string, l int) bool {
	if l > len(b) || l > len(a) {
		return false
	}
	return strings.Compare(a[:l], b[:l]) == 0
}

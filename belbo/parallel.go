// Package parallelsite enables Parallel Sites (see: X)
// It's a very naive and non-elegant solution... but it works
package belbo

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	codeRegex    = regexp.MustCompile(`<code`) // e.g., <code>, <code class="...">, etc.
	endCodeRegex = regexp.MustCompile(`<\/code>`)
	spanFormat   = "<span class=\"belbo_v%d-s%d\">%s</span>"
	separator    = "||"
)

// Replaceable represents a snippet of text that will be replaced by <span> elements
type Replaceable struct {
	InitPos  int
	FinalPos int
	Original string
}

// Location defines a location in the text content
type Location struct {
	InitPos  int
	FinalPos int
}

func EnableParallelContent(content string) string {
	replacements := FindReplaceablesSnippets(content)
	return Replace(content, replacements)
}

// FindReplaceableSnippets finds snippets in the form @{C1||C2||...||Cn}@ that
// will be replaced later by <span> elements.
func FindReplaceablesSnippets(content string) []*Replaceable {
	var replaceables []*Replaceable

	for pos, char := range content {
		// FIXME content cannot end in '@'!
		if rune(char) == '@' && []rune(content)[pos+1] == '{' {
			initIndex := pos
			finalIndex := findFinalIndex(content, pos)

			replaceables = append(replaceables, &Replaceable{
				InitPos:  initIndex,
				FinalPos: finalIndex,
				Original: content[initIndex : finalIndex+1],
			})
		}
	}
	return replaceables
}

// Replace replaces replaceables in the content with <span> elements
func Replace(content string, replacements []*Replaceable) string {
	result := content
	nonReplaceableAreas := NonReplaceableAreas(content)

	for i, r := range replacements {
		// Do not replace replaceables in non-replaceable areas... common sense
		if withinNonReplaceableArea(r, nonReplaceableAreas) {
			continue
		}

		span := replaceableToSpan(r.Original, i)
		result = strings.Replace(result, r.Original, span, 1)
	}

	return result
}

func withinNonReplaceableArea(r *Replaceable, areas []Location) bool {
	for _, a := range areas {
		if r.InitPos >= a.InitPos && r.FinalPos <= a.FinalPos {
			return true
		}
	}

	return false
}

// NonReplaceableAreas are places in the text that should be kept untouched.
// For now, take into account only text within <code> elements
func NonReplaceableAreas(c string) []Location {
	start := 0
	var result []Location
	length := len(c)
	var initloc []int
	var finloc []int

	for start < length {
		initloc = codeRegex.FindStringIndex(c[start:])

		if initloc != nil {
			finloc = endCodeRegex.FindStringIndex(c[start:])
			if finloc == nil {
				break
			}
			result = append(result, Location{InitPos: start + initloc[1], FinalPos: start + finloc[0]})
			start = start + finloc[1]
		} else {
			break
		}
	}

	return result
}

// TODO write a generic FindAllStringIndex
func replaceableToSpan(content string, index int) string {
	var res []string
	r := content[2 : len(content)-2]
	parts := strings.Split(r, separator)

	for i, part := range parts {
		x := fmt.Sprintf(spanFormat, i+1, index+1, part)
		res = append(res, x)
	}

	return strings.Join(res, "")
}

func findFinalIndex(c string, index int) int {
	for i := index; i < len(c); i++ {
		if rune(c[i]) == '}' && rune(c[i+1]) == '@' {
			return i + 1
		}
	}

	return -1
}

package internal_test

import (
	"testing"

	"github.com/marcoslar/belbo/internal"
)

func TestEnableParallelContent(t *testing.T) {
	input := `Hello @{there||here}@! It was nice to @{meet||see}@ you.`
	output := internal.EnableParallelContent(input)
	desired := `Hello <span class="belbo_v1-s1">there</span><span class="belbo_v2-s1">here</span>! It was nice to <span class="belbo_v1-s2">meet</span><span class="belbo_v2-s2">see</span> you.`

	if output != desired {
		t.Fatalf("%s != %s", output, desired)
	}
}

func TestNonReplaceableAreas(t *testing.T) {
	input := `<p>Some <code class="foo">@{hi||bye}@</code></p><pre><code> @{hey||you}@ </code></pre>`
	output := internal.EnableParallelContent(input)

	if output != input {
		t.Fatalf("content should not have changed, %s != %s", output, input)
	}
}

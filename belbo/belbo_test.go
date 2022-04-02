package belbo_test

import (
	"github.com/marcoslar/belbo/belbo"
	"testing"
	"testing/fstest"
)

const (
	BaseLayoutTest = `
{{ define "base_layout" }}
<html>
  <body>{{ template "content" . }}</body>
  {{ template "footer" . }}
</html>
{{ end }}`

	FooterPartialTest = `
{{ define "footer" }}
<footer>Some footer</footer>
{{ end }}
`
	Post1Test = `
---
title = "Post number 1"
---

Some post content here
`
	Post1ExpectedHtml = `
<html>
  <body><p>Some post content here</p>
</body>
  
<footer>Some footer</footer>

</html>
`
	Index = `
---
title = "My Blog"
---

<ul>
	{{ range $i, $page := .AllPages }}
	<li>
		<span>{{ $page.CreatedAt.Format "02 Jan 2006" }}</span>
		<a href="/{{ $page.Url }}">{{ $page.Config.title }}</a>
	</li>
	{{ end }}
</ul>
`
)

func TestBelbo(t *testing.T) {
	myfs := fstest.MapFS{
		"templates/base_layout.html":     {Data: []byte(BaseLayoutTest)},
		"templates/partials/footer.html": {Data: []byte(FooterPartialTest)},
		"posts/2021-10-22-post1.md":      {Data: []byte(Post1Test)},
		"index.md":                       {Data: []byte(Index)},
		"about.md":                       {Data: []byte("About")},
	}

	funcs := map[string]interface{}{
		"reverse": func(s string) string {
			result := ""
			for _, ch := range s {
				result = string(ch) + result
			}

			return result
		},
	}

	var siteGenerator *belbo.Belbo

	setup := func(t *testing.T) {
		siteGenerator = belbo.NewBelbo(&belbo.Config{
			"layout":          "base_layout",
			"content_dir":     []string{"posts"},
			"templates_dir":   "templates",
			"templates":       []string{"footer"},
			"output_dir":      "public",
			"static_dir":      "static",
			"local_server":    false,
			"frontmatter_sep": "---",
			"root_path":       ".",
		}, myfs, funcs)
	}

	setup(t)

	t.Run("config is setup", func(t *testing.T) {
		if siteGenerator.Config == nil {
			t.Fatalf("config is nil")
		}
	})

	t.Run("pages processed", func(t *testing.T) {
		siteGenerator.BuildPages()

		for _, p := range siteGenerator.Pages {
			if p.Html == "" {
				t.Fatalf("page was not parsed correctly (%s)", p.RelativePath)
			}

			if p.RelativePath == "posts/2021-10-22-post1.md" {
				if p.Html != Post1ExpectedHtml {
					t.Fatalf("expect (%s) but got (%s)", Post1ExpectedHtml, p.Html)
				}
			}
		}
	})

	t.Run("funcs are loaded as plugins", func(t *testing.T) {
		_, ok := siteGenerator.Plugins["reverse"]
		if !ok {
			t.Fatalf("expect `reverse` func to be loaded as plugin")
		}
	})
}

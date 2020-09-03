package belbolib_test

import (
	belbo "github.com/lessmarcos/belbo/belbolib"
	"os"
	"strings"
	"testing"
)

func TestBelbo(t *testing.T) {
	var defaultCfg belbo.Params
	setup := func(t *testing.T) {
		defaultCfg = belbo.Params{
			"content_dir":     "posts",
			"templates_dir":   "templates",
			"output_dir":      "public",
			"static_dir":      "static",
			"local_server":    true,
			"frontmatter_sep": "---",
			"root_path":       ".",
		}
		belbo.SetCfg(defaultCfg)
	}

	setup(t)

	t.Run("Integration tests", func(t *testing.T) {
		os.Chdir("example")

		t.Run("correct number of pages in example/ are processed", func(t *testing.T) {
			belbo.SetCfg(defaultCfg)
			pagesToProcess := belbo.PagesToProcess()
			if len(pagesToProcess) != 2 {
				t.Fatalf("%d != %d", 2, len(pagesToProcess))
			}
		})

		t.Run("correct HTML is outputted", func(t *testing.T) {
			pagesToProcess := belbo.PagesToProcess()
			post1Html := string(pagesToProcess[0].ToHtml())
			expectedHtml := "<h1>Hello</h1>\n\n<p>World</p>\n"

			if post1Html != expectedHtml {
				t.Fatalf("%s != %s", post1Html, expectedHtml)
			}
		})
	})

	t.Run("Unit tests", func(t *testing.T) {
		t.Run("content is read", func(t *testing.T) {
			sample := `---
title = "Some title"
---

Hi there`
			page := belbo.BuildPage("foo.md", strings.NewReader(sample))
			expected := "\nHi there"

			if page.Content != expected {
				t.Fatalf("%s != %s", page.Content, expected)
			}
		})

		t.Run("front-matter is read and merged with default values", func(t *testing.T) {
			sample := `---
title = "First post"
---
`
			page := belbo.BuildPage("./posts/2020-12-12-bar.md", strings.NewReader(sample))

			if page.Params.GetString("title") != "First post" {
				t.Fatalf("page title not equal. %s != %s", page.Params.GetString("title"), "First post")
			}

			if page.Name != "bar" {
				t.Fatalf("page name not equal. %s != %s", page.Name, "bar")
			}

			if page.RelativePath != "./posts/2020-12-12-bar.md" {
				t.Fatalf("page relative path not equal. %s != %s", page.RelativePath, "./posts/2020-12-12-bar.md")
			}
		})
	})
}

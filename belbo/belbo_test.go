package belbo_test

import (
	"fmt"
	belbo "github.com/lessmarcos/belbo/belbo"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBelbo(t *testing.T) {
	var defaultCfg belbo.Params
	setup := func(t *testing.T) {
		defaultCfg = belbo.Params{
			"content_dir":     []interface{}{"posts", "logs"},
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
		os.Chdir(filepath.Join("..", "example"))

		t.Run("correct number of pages in example/ are processed", func(t *testing.T) {
			belbo.SetCfg(defaultCfg)
			pagesToProcess := belbo.PagesToProcess()
			if len(pagesToProcess) != 3 {
				t.Fatalf("%d != %d", 3, len(pagesToProcess))
			}
		})

		t.Run("correct HTML is rendered", func(t *testing.T) {
			scenarios := map[string]string{
				"logs/2020-09-03-bye.md":    "<h1>Logs</h1>\n",
				"posts/2020-09-03-hello.md": "<h1>Hello</h1>\n\n<p>Wörld</p>\n",
				"index.md": `<p>Sessions:</p>

<ul>
    
    <li>
       <span class="post-date">2020-09-03 00:00:00 &#43;0000 UTC</span>
       <a href="logs/2020/09/bye">Bye</a>
    </li>
    
    <li>
       <span class="post-date">2020-09-03 00:00:00 &#43;0000 UTC</span>
       <a href="posts/2020/09/hello">Hello</a>
    </li>
    
    <li>
       <span class="post-date">0001-01-01 00:00:00 &#43;0000 UTC</span>
       <a href="index.html">Index</a>
    </li>
    
</ul>
`,
			}

			for _, p := range belbo.PagesToProcess() {
				p.AllPages = belbo.PagesToProcess() // TODO this should not be triggered manually

				postAsHtml := string(p.ToHtml())
				expectedHtml := scenarios[p.RelativePath]

				if postAsHtml != expectedHtml {
					fmt.Println("[" + postAsHtml + "]")
					fmt.Println("<" + expectedHtml + ">")
					t.Fatalf("%s != %s", postAsHtml, expectedHtml)
				}
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

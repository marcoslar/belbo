package internal

import (
	"bufio"
	"bytes"
	"html/template"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/russross/blackfriday.v2"
)

// Page represents anything that may be written out as html
type Page struct {
	// Front-matter
	Config Config

	RawContent string

	// Html representation of the page
	Html template.HTML

	// Path before page is processed (e.g., posts/2020-01-12-example.md)
	RelativePath string

	// e.g., /posts/2020/01/example/index.html
	Url string

	Title string

	// Pointer to all the other pages
	AllPages []*Page

	// CreationDate for pages whose name includes a date
	CreatedAt time.Time

	BuildDir []string
}

func NewPage(pagePath string, rd io.Reader, cfg *Config) *Page {
	brd := bufio.NewReader(rd)
	var tomlBuf bytes.Buffer
	var pageContentBuf bytes.Buffer
	hasToml := false

	for {
		line, err := brd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// We assume we don't write posts with broken front-matter
				if !hasToml {
					pageContentBuf.WriteString(line)
				}
				break
			}
			panic(err)
		}

		frontmatterSep := cfg.GetString("frontmatter_sep")
		if line == frontmatterSep+"\n" {
			hasToml = !hasToml
			continue
		}

		if hasToml {
			tomlBuf.WriteString(line)
		} else {
			pageContentBuf.WriteString(line)
		}
	}

	var frontMatter Config
	if _, err := toml.Decode(tomlBuf.String(), &frontMatter); err != nil {
		panic(err)
	}

	var createdAt time.Time
	if filenameWithDateRegexp.MatchString(pagePath) {
		var err error
		createdAt, err = time.Parse(DateLayout, filepath.Base(pagePath)[:len(DateLayout)])
		if err != nil {
			panic(err)
		}
	}

	page := &Page{
		Config:       cfg.MergeWith(frontMatter),
		RawContent:   pageContentBuf.String(),
		Html:         "",
		RelativePath: pagePath,
		Url:          "",
		Title:        frontMatter.GetString("title"),
		AllPages:     nil,
		CreatedAt:    createdAt,
	}

	// Quick-and-dirty. Fix me later
	var buildPath []string
	if filenameWithDateRegexp.MatchString(page.RelativePath) {
		baseDir := strings.Split(page.RelativePath, "/")
		subDirs := strings.Split(baseDir[1], "-")
		filenameWithoutDate := filenameWithDateRegexp.Split(page.RelativePath, 2)[1][1:]
		buildPath = []string{baseDir[0], subDirs[0], subDirs[1], strings.TrimSuffix(filenameWithoutDate, filepath.Ext(filenameWithoutDate))}
	} else if page.RelativePath != "index.md" {
		ext := path.Ext(page.RelativePath)
		buildPath = []string{page.RelativePath[0 : len(page.RelativePath)-len(ext)]}
	}

	buildDir := filepath.Join(cfg.GetString("root_path"), cfg.GetString("build_dir"))
	page.BuildDir = append([]string{buildDir}, buildPath...)
	page.Url = filepath.Join(page.BuildDir[1:]...)

	return page
}

func (p *Page) ToHtml(b *Belbo) template.HTML {
	baseTemplate := template.Must(template.New("base_layout").Funcs(b.Plugins).Parse(DefaultBaseLayout))
	pageLayout := p.Config.GetString("layout")

	if pageLayout != "" {
		baseTemplate = template.Must(template.New(pageLayout).Funcs(b.Plugins).
			ParseFS(b.Fsys, filepath.Join(b.TemplatesDir, pageLayout+".html")))
	}

	for _, partial := range b.Templates {
		baseTemplate, err := baseTemplate.ParseFS(b.Fsys, filepath.Join(b.TemplatesDir, "partials", partial+".html"))

		if err != nil {
			panic(err)
		}
		baseTemplate = baseTemplate.Funcs(b.Plugins)
	}

	htmlContent := string(blackfriday.Run([]byte(p.RawContent)))
	htmlContent = EnableParallelContent(htmlContent)

	baseTemplate, err := baseTemplate.Parse(`{{ define "content" }}` + htmlContent + `{{ end }}`)
	if err != nil {
		panic(err)
	}

	var dstBuffer bytes.Buffer

	if err := baseTemplate.Execute(&dstBuffer, p); err != nil {
		panic(err)
	}

	bufBytes := dstBuffer.Bytes()
	p.Html = template.HTML(bufBytes)

	return p.Html
}

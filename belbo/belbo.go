package belbo

import (
	"bufio"
	"bytes"
	"github.com/BurntSushi/toml"
	"gopkg.in/russross/blackfriday.v2"
	"html"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	DateLayout = "2006-01-02"
)

var (
	cfg                Params
	contentDir         []string
	templatesDir       string
	OutputDir          string
	rootPath           string
	partialTemplates   []interface{}
	datedFilenameRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}-(?P<all>.*)`)
	defaultBaseLayout  = `{{ define "base_layout" }}{{ template "content" . }}{{ end }}`
)

// Params represents a toml snippet
type Params map[string]interface{}

// Page represents any page that may be written out as HTML
type Page struct {
	// Front-matter
	Params Params

	// Actual content of the page
	Content string

	// Internal representation of the page
	Template *template.Template

	// HTML representation of the page
	Html template.HTML

	// Path before page is processed (e.g., posts/2020-01-12-example.md)
	RelativePath string

	// Path after page is processed (e.g., posts/2020/01/example/)
	PublicPath string

	// Page name without extension and date (e.g., example)
	Name string

	// Pointers to all the other pages
	AllPages []*Page

	// CreationDate for pages whose name includes a date
	CreatedAt time.Time
}

func SetCfg(cfg_ Params) {
	cfg = cfg_
	rootPath = cfg.GetString("root_path")
	contentDirSlice := cfg.GetStringSlice("content_dir")
	for _, contentDirEntry := range contentDirSlice {
		contentDir = append(contentDir, filepath.Join(rootPath, contentDirEntry))
	}
	templatesDir = filepath.Join(rootPath, cfg.GetString("templates_dir"))
	OutputDir = filepath.Join(rootPath, cfg.GetString("output_dir"))
	partialTemplates = func(cfg Params) []interface{} {
		if cfg.Get("partials") != nil {
			return cfg.Get("partials").([]interface{})
		}
		return []interface{}{}
	}(cfg)
}

func PagesToProcess() []*Page {
	var result []*Page
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		currDir := filepath.Dir(path)
		if currDir != rootPath && !stringInSlice(currDir, contentDir) {
			return filepath.SkipDir
		}

		ext := filepath.Ext(filepath.Base(path))

		if ext != ".md" && ext != ".html" {
			return nil
		}

		fd, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		defer fd.Close()
		page := BuildPage(path, fd)
		isDraft := page.Params.Get("draft")
		if isDraft != nil && isDraft.(bool) {
			return nil
		}

		result = append(result, page)

		return nil
	})

	if err != nil {
		panic(err)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

func BuildPage(pagePath string, fd io.Reader) *Page {
	rd := bufio.NewReader(fd)
	var tomlBuf bytes.Buffer
	var pageContentBuf bytes.Buffer
	hasToml := false

	for {
		line, err := rd.ReadString('\n')

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

	var frontMatter Params
	if _, err := toml.Decode(tomlBuf.String(), &frontMatter); err != nil {
		panic(err)
	}

	var createdAt time.Time
	if datedFilenameRegex.MatchString(pagePath) {
		var err error
		createdAt, err = time.Parse(DateLayout, filepath.Base(pagePath)[:len(DateLayout)])
		if err != nil {
			panic(err)
		}
	}

	result := &Page{
		Params:       cfg.MergeWith(frontMatter),
		Content:      pageContentBuf.String(),
		RelativePath: pagePath,
		CreatedAt:    createdAt,
		Name:         pageName(pagePath),
	}

	publicPath, _ := getOutputDir(result)
	result.PublicPath = filepath.Join(publicPath[1:]...)

	return result
}

func (page *Page) toTemplate() *template.Template {
	baseTemplate, err := getBaseTemplate(page)
	if err != nil {
		panic(err)
	}

	for _, partial := range partialTemplates {
		baseTemplate, err = template.Must(baseTemplate.Clone()).ParseFiles(
			filepath.Join(templatesDir, "partials", (partial).(string)+".html"))
		if err != nil {
			panic(err)
		}
	}

	htmlContent := string(blackfriday.Run([]byte(page.Content)))
	htmlContent = EnableParallelContent(htmlContent)
	htmlContent = html.UnescapeString(htmlContent)

	tmpl, err := template.Must(baseTemplate.Clone()).Funcs(template.FuncMap{
		"SimpleDate": FormattedDate,
	}).
		Parse(`{{ define "content" }}` + htmlContent + `{{ end }}`)

	if err != nil {
		panic(err)
	}

	page.Template = tmpl

	return page.Template
}

func (page *Page) ToHtml() template.HTML {
	page.toTemplate()

	dirs, shouldCreateIntermediaryDirs := getOutputDir(page)
	if shouldCreateIntermediaryDirs {
		if err := os.MkdirAll(filepath.Join(dirs...), os.ModePerm); err != nil {
			panic(err)
		}
		dirs = append(dirs, "index.html")
	}

	dst, err := os.Create(filepath.Join(dirs...))

	if err != nil {
		panic(err)
	}
	defer dst.Close()

	var dstBuffer bytes.Buffer
	if err := page.Template.Execute(&dstBuffer, page); err != nil {
		panic(err)
	}

	bufBytes := dstBuffer.Bytes()
	dst.Write(bufBytes)

	page.Html = template.HTML(bufBytes)
	return page.Html
}

func getBaseTemplate(page *Page) (*template.Template, error) {
	if page.Params["layout"] == nil {
		return template.New("base_layout").Parse(defaultBaseLayout)
	} else {
		layout := page.Params.GetString("layout")
		templatesDir := page.Params.GetString("templates_dir")
		return template.New(layout).ParseFiles(filepath.Join(templatesDir, layout+".html"))
	}
}

func getOutputDir(page *Page) ([]string, bool) {
	if datedFilenameRegex.MatchString(page.RelativePath) {
		dirs := strings.Split(page.CreatedAt.Format(DateLayout), "-")
		parts := strings.Split(page.RelativePath, "/")
		contentDir := parts[0] // Strong assumption: relative path is like path/file-name.ext
		return []string{OutputDir, contentDir, dirs[0], dirs[1], page.Name}, true
	} else if page.RelativePath != "index.md" {
		return []string{OutputDir, page.Name}, true
	} else {
		return []string{OutputDir, replaceExt(page.RelativePath, ".html")}, false
	}
}

func pageName(pagePath string) string {
	basename := filepath.Base(pagePath)
	if !datedFilenameRegex.MatchString(basename) {
		return strings.TrimSuffix(basename, filepath.Ext(basename))
	}

	matches := datedFilenameRegex.FindStringSubmatch(pagePath)
	return strings.TrimSuffix(matches[1], filepath.Ext(matches[1]))
}

func FormattedDate(t time.Time) string {
	return t.Format(DateLayout)
}

func replaceExt(pagePath string, newExt string) string {
	ext := path.Ext(pagePath)
	return pagePath[0:len(pagePath)-len(ext)] + newExt
}

func (cfg Params) MergeWith(b Params) Params {
	result := make(Params)
	for k, v := range cfg {
		result[k] = v
	}

	for k, v := range b {
		result[k] = v
	}
	return result
}

func (cfg Params) GetString(key string) string {
	if cfg.Get(key) != nil {
		return cfg.Get(key).(string)
	}

	return ""
}

func (cfg Params) GetStringSlice(key string) []string {
	var result []string
	values := cfg.Get(key)
	if values != nil {
		for _, val := range values.([]interface{}) {
			result = append(result, val.(string))
		}
	}

	return result
}

func (cfg Params) Get(key string) interface{} {
	if val, ok := cfg[key]; ok {
		return val
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

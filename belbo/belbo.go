package belbo

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
)

const (
	DateLayout        = "2006-01-02"
	DefaultBaseLayout = `{{ define "base_layout" }}{{ template "content" . }}{{ end }}`
)

var filenameWithDateRegexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// Belbo takes as input a file system and a configuration file
// to transform a bunch of MD and html templates into html
type Belbo struct {
	Fsys         fs.FS
	Config       *Config
	ContentDir   []string
	TemplatesDir string
	Templates    []string
	BuildDir     string
	Pages        []*Page
}

func NewBelbo(cfg *Config, fsys fs.FS) *Belbo {
	rootPath := cfg.GetString("root_path")

	var contentDir []string
	for _, entry := range cfg.GetStringSlice("content_dir") {
		contentDir = append(contentDir, filepath.Join(rootPath, entry))
	}

	return &Belbo{
		Fsys:         fsys,
		Config:       cfg,
		ContentDir:   contentDir,
		TemplatesDir: filepath.Join(rootPath, cfg.GetString("templates_dir")),
		BuildDir:     filepath.Join(rootPath, cfg.GetString("build_dir")),
		Templates:    cfg.GetStringSlice("templates"),
	}
}

func (b *Belbo) BuildPages() {
	rootPath := b.Config.GetString("root_path")
	err := fs.WalkDir(b.Fsys, rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		currDir := filepath.Dir(path)
		if currDir != rootPath && !b.IsContentDir(currDir) {
			return filepath.SkipDir
		}

		ext := filepath.Ext(filepath.Base(path))

		if ext != ".md" && ext != ".html" {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		fd, err := b.Fsys.Open(path)
		if err != nil {
			panic(err)
		}

		defer fd.Close()

		b.Pages = append(b.Pages, NewPage(path, fd, b.Config))
		return nil
	})

	allPages := b.Pages
	sort.SliceStable(allPages, func(a, b int) bool {
		return allPages[a].CreatedAt.After(allPages[b].CreatedAt)
	})

	for _, p := range b.Pages {
		p.AllPages = b.Pages
		p.ToHtml(b)
	}

	if err != nil {
		panic(err)
	}
}

func (b *Belbo) IsContentDir(dirname string) bool {
	for _, d := range b.ContentDir {
		if d == dirname {
			return true
		}
	}

	return false
}

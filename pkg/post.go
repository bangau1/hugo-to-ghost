package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
)

type MarkdownPost struct {
	// These are details from the FrontMatter
	Title    string    `yaml:"title"`
	Subtitle string    `yaml:"subtitle"`
	Date     time.Time `yaml:"date"`
	Image    string    `yaml:"image"`
	Tags     []string  `yaml:"tags"`
	Keywords []string  `yaml:"keywords"`
	IsDraft  bool      `yaml:"draft"`

	// Slug is the post url name (e.g: 2023-10-05-dont-forget-media-rescan-on-android)
	Slug string

	// Content is the post's content (the rest of the content beside the FrontMatter)
	Content string
}

func NewPostFromFrontMatterDocFile(frontMatterDocFile string) (MarkdownPost, error) {
	file, err := os.Open(frontMatterDocFile)
	if err != nil {
		return MarkdownPost{}, err
	}
	post := MarkdownPost{}
	content, err := frontmatter.Parse(file, &post)
	if err != nil {
		return MarkdownPost{}, err
	}

	post.Content = string(content)
	// assign the slug from the filename (without the extension)
	post.Slug = strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name()))
	return post, nil
}

func (m *MarkdownPost) ChangeStaticContentPrefix(oldPrefix string, newPrefix string) {
	// TODO: this is not correct implementation
	// since it also changes all occurences, not just the prefix.
	// Should improve this later, by strictly changes all content inside the
	// ![image content](<prefix>/xxx) pattern.
	m.Content = strings.ReplaceAll(m.Content, oldPrefix, newPrefix)
}

package pkg

import (
	"os"
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
	post.Slug = "" // TODO(bangau1)
	return post, nil
}

package pkg

import (
	"os"
	"time"

	"github.com/adrg/frontmatter"
)

type MarkdownPost struct {
	// These are details from the FrontMatter
	Title    string
	Subtitle string
	Date     time.Time
	Image    string
	Tags     []string
	Keywords []string

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
	return post, nil
}

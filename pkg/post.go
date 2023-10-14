package pkg

import (
	"os"
	"time"

	"github.com/adrg/frontmatter"
)

type Post struct {
	Title    string
	Subtitle string
	Date     time.Time
	Image    string
	Tags     []string
	Keywords []string

	Content string
}

func NewPostFromFrontMatterDocFile(frontMatterDocFile string) (Post, error) {
	file, err := os.Open(frontMatterDocFile)
	if err != nil {
		return Post{}, err
	}
	post := Post{}
	content, err := frontmatter.Parse(file, &post)
	if err != nil {
		return Post{}, err
	}

	post.Content = string(content)
	return post, nil
}

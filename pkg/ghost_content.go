package pkg

import (
	"encoding/json"
)

// GhostContent represent the data to be imported to Ghost
// See: https://ghost.org/docs/migration/content/
type GhostContent struct {
	Title        string `json:"title,omitempty"`
	Slug         string `json:"slug,omitempty"`
	FeatureImage string `json:"feature_image,omitempty"`
	Mobiledoc    string `json:"mobiledoc,omitempty"`
	Status       string `json:"status,omitempty"`
	PublishedAt  int64  `json:"published_at"`
}

func NewGhostContentFromMarkdownPost(post MarkdownPost) (GhostContent, error) {
	// TODO: use enum for this. It's good enough for now.
	status := "published"
	if post.IsDraft {
		status = "draft"
	}

	mobiledoc := map[string]any{
		"version": "0.3.1",
		"markups": []string{},
		"atoms":   []string{},
		"sections": [][]int{
			{10, 10},
		},
		"cards": [][]any{
			{
				"markdown", map[string]string{
					"cardName": "markdown",
					"markdown": post.Content,
				},
			},
		},
	}
	mobiledocsStringify, err := json.Marshal(mobiledoc)
	if err != nil {
		return GhostContent{}, err
	}
	return GhostContent{
		Title:        post.Title,
		Slug:         post.Slug,
		FeatureImage: post.Image,
		Mobiledoc:    string(mobiledocsStringify),
		Status:       status,
		PublishedAt:  post.Date.UnixMilli(),
	}, nil
}

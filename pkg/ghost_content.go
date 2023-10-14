package pkg

import (
	"encoding/json"
	"time"
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

type GhostImportData struct {
	Contents []GhostContent
}

type ghostImportDTO struct {
	Meta ghostImportMetadata `json:"meta"`
	Data ghostImportData     `json:"data"`
}

type ghostImportData struct {
	Posts []GhostContent `json:"posts"`
	// TODO to support tags and post_tags
}

type ghostImportMetadata struct {
	ExportedTimestampMsec int64  `json:"exported_on,omitempty"`
	Version               string `json:"version,omitempty"`
}

func NewGhostImportData(contents ...GhostContent) GhostImportData {
	return GhostImportData{
		Contents: contents,
	}
}

func (g *GhostImportData) ToJson() (string, error) {
	data := ghostImportDTO{
		Meta: ghostImportMetadata{
			ExportedTimestampMsec: time.Now().UnixMilli(),
			Version:               "2.14.0", // ghost version that is valid for (TODO: research more on this)
		},
		Data: ghostImportData{
			Posts: g.Contents,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

package pkg

import (
	"encoding/json"
	"time"
)

// GhostContent represent the data to be imported to Ghost
// See: https://ghost.org/docs/migration/content/
type GhostContent struct {
	Id           string    `json:"id,omitempty"`
	UUID         string    `json:"uuid,omitempty"`
	Title        string    `json:"title,omitempty"`
	Slug         string    `json:"slug,omitempty"`
	FeatureImage string    `json:"feature_image,omitempty"`
	Mobiledoc    string    `json:"mobiledoc,omitempty"`
	Status       string    `json:"status,omitempty"`
	PublishedAt  time.Time `json:"published_at"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

func NewGhostContentFromMarkdownPost(post MarkdownPost) (GhostContent, error) {
	// TODO: use enum for this. It's good enough for now.
	status := "published"
	if post.IsDraft {
		status = "draft"
	}
	// Note: for now, just use the working version of mobiledoc described here:
	// https://forum.ghost.org/t/markdown-to-mobiledoc-converter/5203
	// Notably the sections part is quite tricky. So just use the 1 big section there
	mobiledoc := map[string]any{
		"version": "0.3.1",
		"markups": []string{},
		"atoms":   []string{},
		"sections": [][]any{
			{10, 0},
			{1, "p", []any{}},
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
		PublishedAt:  post.Date,
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

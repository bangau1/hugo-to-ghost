package cmd

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bangau1/hugo-to-ghost/pkg"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hugo-to-ghost",
	Short: "A simple CLI that helps migrating contents from Hugo (Markdown) to Ghost",
	Long: `A simple CLI that helps migrating contents from Hugo (Markdown) to Ghost.

This CLI program will create/update the Ghost's post from Hugo post (markdown).

For example:
- hugo-to-ghost --contentDir ./content/english/post 
	reads all markdown files within --contentDir, convert it to Ghost's post content format and upload it into Ghost Admin API (create or update, based on the slug).

- hugo-to-ghost --contentDir ./content/english/post --staticContentPrefixChanges "/img/uploads/,/content/images/hugo/"
	same as previous example, but additionally it changes the image assets in Hugo (that previously located in /img/uploads/ path) to /content/images/hugo folder inside the ghost installation.
	Note that you need to upload manually the image/static assets from Hugo to the Ghost's. This tools doesn't handle image upload automatically (it just helps convert the prefix url)
`,
	Run: cmdRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	contentDir                 string
	staticContentPrefixChanges string

	ghostAdminAPIKey string
	ghostURL         string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rootCmd.Flags().StringVar(&contentDir, "contentDir", "", "The content directory where the frontmatter posts are located. Example: 'content/english/posts/'")
	rootCmd.Flags().StringVar(&staticContentPrefixChanges, "staticContentPrefixChanges", "", "Prefix changes rule, separated by comma. The total element must be even number. Example: /img/uploads/,/content/images/hugo/,/oldPrefix2/,/newPrefix2/,etc,etc ")
	rootCmd.Flags().StringVar(&ghostAdminAPIKey, "ghostAdminAPIKey", "", "The Ghost's Admin API Key. It's needed for create/update post in Ghost")
	rootCmd.Flags().StringVar(&ghostURL, "ghostUrl", "http://localhost:8080", "The Ghost's URL")

	rootCmd.MarkFlagRequired("contentDir")
	rootCmd.MarkFlagRequired("ghostAdminAPIKey")
}

func cmdRun(cmd *cobra.Command, args []string) {
	ghostApi := pkg.NewGhostAdminAPI(ghostURL, ghostAdminAPIKey)

	staticContentPrefixChangesRules := make([]string, 0)
	if len(staticContentPrefixChanges) > 0 {
		staticContentPrefixChangesRules = strings.Split(staticContentPrefixChanges, ",")
	}
	if len(staticContentPrefixChangesRules)%2 > 0 {
		log.Fatal("staticContentPrefixChanges is invalid. This should be even number")
	}

	files, err := os.ReadDir(contentDir)
	if err != nil {
		log.Fatal("error when readDir", err)
	}

	for _, file := range files {
		// if the file is a markdown file (.md)
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			mdFilePath, err := filepath.Abs(contentDir + "/" + file.Name())
			if err != nil {
				log.Fatal(err)
			}

			// read the file and convert it to MarkdownPost
			post, err := pkg.NewPostFromFrontMatterDocFile(mdFilePath)
			if err != nil {
				log.Fatal("error when reading the markdown file", err)
			}

			// apply changes to the post's static content prefix path
			// must be even number
			for i := 0; i < len(staticContentPrefixChangesRules); i += 2 {
				post.ChangeStaticContentPrefix(staticContentPrefixChangesRules[i], staticContentPrefixChangesRules[i+1])
			}

			// convert MarkdownPost to GhostContent
			ghostContent, err := pkg.NewGhostContentFromMarkdownPost(post)
			if err != nil {
				log.Fatal("error when converting markdown post to Ghost's post", err)
			}

			if post.Slug != "" {
				// check existing post with the same slug, to decide whether it's a CREATE or UPDATE operation
				existingPost, err := ghostApi.GetPostBySlug(context.Background(), ghostContent.Slug)
				if err != nil {
					// if the error is not ErrNotFound, then exit
					if !errors.Is(err, pkg.ErrNotFound) {
						log.Fatal("error when getPostBySlug ", ghostContent.Slug, err)
					}
				}

				// override the id of ghostContent to existingPost.id
				// if existingPost is empty, then the ID and UUID is also empty, so should be fine
				ghostContent.Id = existingPost.Id
				ghostContent.UUID = existingPost.UUID

				// if the same post already in the Ghost's, then it's an UPDATE operation
				if ghostContent.Id != "" {
					// note: we need to override the .UpdatedAt from the existingPost in Ghost, since Ghost uses
					// that as mechanism to detect concurrent edit. Internally if Update content is success, the .UpdatedAt is updated accordingly in the database.
					ghostContent.UpdatedAt = existingPost.UpdatedAt
					updatedData, err := ghostApi.UpdatePost(context.Background(), ghostContent)
					if err != nil {
						log.Fatal("error when UpdatePost", err)
					}
					log.Printf("⬆️ slug=%s is UPDATED at %s\n", ghostContent.Slug, updatedData.Url)
				} else { // it's a CREATE new post operation
					updatedData, err := ghostApi.CreatePost(context.Background(), ghostContent)
					if err != nil {
						log.Fatal("error when CreatePost", err)
					}
					log.Printf("✅ slug=%s is CREATED at %s\n", ghostContent.Slug, updatedData.Url)
				}
			} else {
				log.Fatal("unexpected error: the post slug is empty for file: ", file.Name())
			}
		}
	}
}

package pkg

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GhostAdminAPI struct {
	http.RoundTripper

	URL    string
	APIKey string

	jwtToken   string
	httpClient *http.Client
}

func (g GhostAdminAPI) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Add("Authorization", fmt.Sprintf("Ghost %s", g.jwtToken))
	return g.RoundTripper.RoundTrip(request)
}

const GHOST_ADMIN_API_BASE = "/ghost/api/admin/"

func NewGhostAdminAPI(url, apiKey string) GhostAdminAPI {
	splits := strings.Split(apiKey, ":")
	id, secret := splits[0], splits[1]
	log.Println("id:", id)
	log.Println("secret:", secret)
	// header := map[string]string{
	// 	"alg": "HS256",
	// 	"typ": "JWT",
	// 	"kid": id,
	// }
	iat := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": iat,
		"exp": iat + 5*60,
		"aud": "/admin/",
	})
	token.Header["kid"] = id
	key, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatal(err)
	}
	jwtToken, err := token.SignedString(key)
	if err != nil {
		log.Fatal("error when generate the jwt token for admin API access", err)
	}
	log.Println("jwtToken: ", jwtToken)
	ghostAdminApi := GhostAdminAPI{
		RoundTripper: http.DefaultTransport, // assign the DefaultTransport
		URL:          url,
		APIKey:       apiKey,
		jwtToken:     jwtToken,
	}
	ghostAdminApi.httpClient = &http.Client{
		Transport: ghostAdminApi,
	}
	return ghostAdminApi
}

type ghostGetContentPostsData struct {
	Posts []GhostContent `json:"posts,omitempty"`
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (g *GhostAdminAPI) GetPostBySlug(ctx context.Context, slug string) (GhostContent, error) {
	url := g.URL + fmt.Sprintf("%s/posts/slug/%s/", GHOST_ADMIN_API_BASE, slug)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return GhostContent{}, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return GhostContent{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GhostContent{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return GhostContent{}, fmt.Errorf("%s", body)
	}

	data := ghostGetContentPostsData{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return GhostContent{}, err
	}

	if len(data.Posts) == 0 {
		return GhostContent{}, ErrNotFound
	}

	return data.Posts[0], nil
}

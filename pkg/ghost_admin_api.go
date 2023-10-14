package pkg

import (
	"bytes"
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

	URL string
	// APIKey is the Ghost's Admin API Key
	APIKey string

	jwtToken   string
	httpClient *http.Client
}

func (g GhostAdminAPI) RoundTrip(request *http.Request) (*http.Response, error) {
	// ensure the http client of GhostAdminAPI inject the jwt token transparently
	request.Header.Add("Authorization", fmt.Sprintf("Ghost %s", g.jwtToken))
	request.Header.Add("Content-Type", "application/json; charset=UTF-8")
	return g.RoundTripper.RoundTrip(request)
}

const GHOST_ADMIN_API_BASE = "ghost/api/admin"

func NewGhostAdminAPI(url, apiKey string) GhostAdminAPI {
	// generate jwt token: based on https://ghost.org/docs/admin-api/#token-authentication
	splits := strings.Split(apiKey, ":")
	id, secret := splits[0], splits[1]

	iat := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": iat,
		"exp": iat + 10*60, // set the expiration for 10 minutes (TODO: set it longer (make it configurable) if we have a lot of posts to be migrated)
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

// ghostPostsAPIData represents Ghost's Post: both the request data and the successful response data
type ghostPostsAPIData struct {
	Posts []GhostContent `json:"posts,omitempty"`
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (g *GhostAdminAPI) doHttpCall(ctx context.Context, method string, apiPath string, body io.Reader) (
	int, []byte, error,
) {
	url := g.URL + apiPath
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, respBody, err
}

func (g *GhostAdminAPI) GetPostBySlug(ctx context.Context, slug string) (GhostContent, error) {
	apiPath := fmt.Sprintf("%s/posts/slug/%s/", GHOST_ADMIN_API_BASE, slug)
	statusCode, respBytes, err := g.doHttpCall(ctx, "GET", apiPath, nil)

	if err != nil {
		return GhostContent{}, err
	}

	if statusCode != http.StatusOK {
		if statusCode == http.StatusNotFound {
			return GhostContent{}, ErrNotFound
		}
		return GhostContent{}, fmt.Errorf("status_code=%d body=%s", statusCode, string(respBytes))
	}

	data := ghostPostsAPIData{}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return GhostContent{}, err
	}

	if len(data.Posts) == 0 {
		return GhostContent{}, ErrNotFound
	}

	return data.Posts[0], nil
}

func (g *GhostAdminAPI) UpdatePost(ctx context.Context, content GhostContent) (GhostContent, error) {
	apiPath := fmt.Sprintf("%s/posts/%s/", GHOST_ADMIN_API_BASE, content.Id)
	payload := ghostPostsAPIData{
		Posts: []GhostContent{content},
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return GhostContent{}, err
	}

	statusCode, respBytes, err := g.doHttpCall(ctx, "PUT", apiPath, bytes.NewReader(jsonBytes))
	if err != nil {
		return GhostContent{}, err
	}

	if statusCode != http.StatusOK {
		if statusCode == http.StatusNotFound {
			return GhostContent{}, ErrNotFound
		}
		return GhostContent{}, fmt.Errorf("status_code=%d body=%s", statusCode, string(respBytes))
	}

	data := ghostPostsAPIData{}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return GhostContent{}, err
	}

	if len(data.Posts) == 0 {
		return GhostContent{}, ErrNotFound
	}

	return data.Posts[0], nil
}

func (g *GhostAdminAPI) CreatePost(ctx context.Context, content GhostContent) (GhostContent, error) {
	apiPath := fmt.Sprintf("%s/posts/", GHOST_ADMIN_API_BASE)
	payload := ghostPostsAPIData{
		Posts: []GhostContent{content},
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return GhostContent{}, err
	}

	statusCode, respBytes, err := g.doHttpCall(ctx, "POST", apiPath, bytes.NewReader(jsonBytes))
	if err != nil {
		return GhostContent{}, err
	}

	if statusCode != http.StatusCreated {
		return GhostContent{}, fmt.Errorf("status_code=%d body=%s", statusCode, string(respBytes))
	}

	data := ghostPostsAPIData{}
	err = json.Unmarshal(respBytes, &data)
	if err != nil {
		return GhostContent{}, err
	}

	if len(data.Posts) == 0 {
		return GhostContent{}, ErrNotFound
	}

	return data.Posts[0], nil
}

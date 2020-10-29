package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	url "net/url"
	"time"
)

const unsplashClientId = "OYoMWLGo0m4jrwTVP5swvH6oJECxdCDA__EeCXnyVA0"

type UnsplashResponse struct {
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
	Results    []struct {
		ID             string      `json:"id"`
		CreatedAt      string      `json:"created_at"`
		UpdatedAt      string      `json:"updated_at"`
		PromotedAt     interface{} `json:"promoted_at"`
		Width          int         `json:"width"`
		Height         int         `json:"height"`
		Color          string      `json:"color"`
		Description    interface{} `json:"description"`
		AltDescription string      `json:"alt_description"`
		Urls           struct {
			Raw     string `json:"raw"`
			Full    string `json:"full"`
			Regular string `json:"regular"`
			Small   string `json:"small"`
			Thumb   string `json:"thumb"`
		} `json:"urls"`
		Links struct {
			Self             string `json:"self"`
			HTML             string `json:"html"`
			Download         string `json:"download"`
			DownloadLocation string `json:"download_location"`
		} `json:"links"`
		Categories             []interface{} `json:"categories"`
		Likes                  int           `json:"likes"`
		LikedByUser            bool          `json:"liked_by_user"`
		CurrentUserCollections []interface{} `json:"current_user_collections"`
		Sponsorship            interface{}   `json:"sponsorship"`
		User                   struct {
			ID              string      `json:"id"`
			UpdatedAt       string      `json:"updated_at"`
			Username        string      `json:"username"`
			Name            string      `json:"name"`
			FirstName       string      `json:"first_name"`
			LastName        string      `json:"last_name"`
			TwitterUsername interface{} `json:"twitter_username"`
			PortfolioURL    interface{} `json:"portfolio_url"`
			Bio             interface{} `json:"bio"`
			Location        interface{} `json:"location"`
			Links           struct {
				Self      string `json:"self"`
				HTML      string `json:"html"`
				Photos    string `json:"photos"`
				Likes     string `json:"likes"`
				Portfolio string `json:"portfolio"`
				Following string `json:"following"`
				Followers string `json:"followers"`
			} `json:"links"`
			ProfileImage struct {
				Small  string `json:"small"`
				Medium string `json:"medium"`
				Large  string `json:"large"`
			} `json:"profile_image"`
			InstagramUsername interface{} `json:"instagram_username"`
			TotalCollections  int         `json:"total_collections"`
			TotalLikes        int         `json:"total_likes"`
			TotalPhotos       int         `json:"total_photos"`
			AcceptedTos       bool        `json:"accepted_tos"`
		} `json:"user"`
		Tags []struct {
			Type   string `json:"type"`
			Title  string `json:"title"`
			Source struct {
				Ancestry struct {
					Type struct {
						Slug       string `json:"slug"`
						PrettySlug string `json:"pretty_slug"`
					} `json:"type"`
					Category struct {
						Slug       string `json:"slug"`
						PrettySlug string `json:"pretty_slug"`
					} `json:"category"`
					Subcategory struct {
						Slug       string `json:"slug"`
						PrettySlug string `json:"pretty_slug"`
					} `json:"subcategory"`
				} `json:"ancestry"`
				Title           string `json:"title"`
				Subtitle        string `json:"subtitle"`
				Description     string `json:"description"`
				MetaTitle       string `json:"meta_title"`
				MetaDescription string `json:"meta_description"`
				CoverPhoto      struct {
					ID             string `json:"id"`
					CreatedAt      string `json:"created_at"`
					UpdatedAt      string `json:"updated_at"`
					PromotedAt     string `json:"promoted_at"`
					Width          int    `json:"width"`
					Height         int    `json:"height"`
					Color          string `json:"color"`
					Description    string `json:"description"`
					AltDescription string `json:"alt_description"`
					Urls           struct {
						Raw     string `json:"raw"`
						Full    string `json:"full"`
						Regular string `json:"regular"`
						Small   string `json:"small"`
						Thumb   string `json:"thumb"`
					} `json:"urls"`
					Links struct {
						Self             string `json:"self"`
						HTML             string `json:"html"`
						Download         string `json:"download"`
						DownloadLocation string `json:"download_location"`
					} `json:"links"`
					Categories             []interface{} `json:"categories"`
					Likes                  int           `json:"likes"`
					LikedByUser            bool          `json:"liked_by_user"`
					CurrentUserCollections []interface{} `json:"current_user_collections"`
					Sponsorship            interface{}   `json:"sponsorship"`
					User                   struct {
						ID              string `json:"id"`
						UpdatedAt       string `json:"updated_at"`
						Username        string `json:"username"`
						Name            string `json:"name"`
						FirstName       string `json:"first_name"`
						LastName        string `json:"last_name"`
						TwitterUsername string `json:"twitter_username"`
						PortfolioURL    string `json:"portfolio_url"`
						Bio             string `json:"bio"`
						Location        string `json:"location"`
						Links           struct {
							Self      string `json:"self"`
							HTML      string `json:"html"`
							Photos    string `json:"photos"`
							Likes     string `json:"likes"`
							Portfolio string `json:"portfolio"`
							Following string `json:"following"`
							Followers string `json:"followers"`
						} `json:"links"`
						ProfileImage struct {
							Small  string `json:"small"`
							Medium string `json:"medium"`
							Large  string `json:"large"`
						} `json:"profile_image"`
						InstagramUsername string `json:"instagram_username"`
						TotalCollections  int    `json:"total_collections"`
						TotalLikes        int    `json:"total_likes"`
						TotalPhotos       int    `json:"total_photos"`
						AcceptedTos       bool   `json:"accepted_tos"`
					} `json:"user"`
				} `json:"cover_photo"`
			} `json:"source,omitempty"`
		} `json:"tags"`
	} `json:"results"`
}

// https://api.unsplash.com/search/photos/?client_id=OYoMWLGo0m4jrwTVP5swvH6oJECxdCDA__EeCXnyVA0&query=Livestream%2026%20Min%20Fusion%20%2B%20Bonus
func GetUnsplashImages(text string) ([]byte, error) {

	encodedText := url.QueryEscape(text)
	u := fmt.Sprintf("https://api.unsplash.com/search/photos/?client_id=%s&query=%s", unsplashClientId, encodedText)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, bErr := ioutil.ReadAll(resp.Body)
	if bErr != nil {
		return nil, bErr
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return nil, fmt.Errorf("received status code: %d from zoom, err: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func RandomImageFromUnsplashResponse(in []byte) (string, error) {
	unsplashResp := &UnsplashResponse{}

	if unmarshalUnsplashErr := json.Unmarshal(in, unsplashResp); unmarshalUnsplashErr != nil {
		return "", unmarshalUnsplashErr
	}
	imagesToUse := unsplashResp.Results

	if len(imagesToUse) == 0 {
		return "", fmt.Errorf("error in unsplash response: %+v, produced 0 results", unsplashResp)
	}

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator
	index := r.Intn(len(imagesToUse))

	randResult := imagesToUse[index].Urls.Full
	if randResult == "" {
		return "", fmt.Errorf("error parsing unsplash response, resulting image has empty Urls.Full field, result: %+v", imagesToUse[index])
	}
	return randResult, nil
}

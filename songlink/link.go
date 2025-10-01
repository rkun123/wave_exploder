package songlink

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

func (s SonglinkImpl) Info(ctx context.Context, inputURL string) (*LinkResponse, error) {
	u, err := url.Parse(s.baseURL() + "/links")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("url", url.QueryEscape(inputURL))
	u.RawQuery = q.Encode()

	log.Printf("API URL: %s", u.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("API request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-200 status: %d", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Response body: %s", string(body))
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	linkResp := &LinkResponse{}

	if err := json.NewDecoder(resp.Body).Decode(linkResp); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return nil, err
	}

	log.Printf("LinkResponse: %+v", linkResp)

	return linkResp, nil
}

func (s SonglinkImpl) Link(ctx context.Context, inputURL string) (string, error) {
	resp, err := s.Info(ctx, inputURL)
	if err != nil {
		return "", err
	}
	return resp.PageURL, nil
}

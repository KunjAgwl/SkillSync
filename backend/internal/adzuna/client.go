package adzuna

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"SkillSync/internal/db"
)

type Client struct {
	AppID  string
	AppKey string
	Client *http.Client
}

func NewClient(appID, appKey string) *Client {
	return &Client{
		AppID:  appID,
		AppKey: appKey,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}


func (c *Client) FetchJobs(query, location string, pages int) ([]db.Job, error) {
	allJobs := []db.Job{}

	for page := 1; page <= pages; page++ {
		jobs, err := c.fetchPage(query, location, page)
		if err != nil {
			return nil, err
		}
		allJobs = append(allJobs, jobs...)
	}

	return allJobs, nil
}


func (c *Client) fetchPage(query, location string, page int) ([]db.Job, error) {
	baseURL := fmt.Sprintf(
		"https://api.adzuna.com/v1/api/jobs/in/search/%d",
		page,
	)

	params := url.Values{}
	params.Set("app_id", c.AppID)
	params.Set("app_key", c.AppKey)
	params.Set("results_per_page", "50")
	params.Set("what", query)
	params.Set("where", location)

	reqURL := baseURL + "?" + params.Encode()

	resp, err := c.Client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("adzuna API error: %s", resp.Status)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	jobs := []db.Job{}
	for _, j := range apiResp.Results {
		job := db.Job{
			Title:       j.Title,
			Company:     j.Company.DisplayName,
			Location:    j.Location.DisplayName,
			Description: j.Description,
			ApplyURL:    j.RedirectURL,
			Source:      "adzuna",
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

package adzuna

type apiResponse struct {
	Results []apiJob `json:"results"`
}

type apiJob struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	RedirectURL string `json:"redirect_url"`

	Company struct {
		DisplayName string `json:"display_name"`
	} `json:"company"`

	Location struct {
		DisplayName string `json:"display_name"`
	} `json:"location"`
}

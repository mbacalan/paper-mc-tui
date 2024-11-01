package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIResponse struct {
	Versions []string `json:"versions"`
	Builds   []string `json:"builds"`
}

// FetchAPIData makes an HTTP GET request to the specified URL and returns the parsed JSON response
func FetchAPIData(url string) (*APIResponse, error) {
	// Create an HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a new request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers if needed
	req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Authorization", "Bearer YOUR_TOKEN") // Uncomment if needed

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse JSON response
	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &apiResponse, nil
}

const url = "https://api.papermc.io/v2"

func GetLatestVersionNr() (string, error) {
	versions, err := FetchAPIData(url + "/projects/paper/")

	if err != nil {
		return "", err
	}

	latestVersion := versions.Versions[len(versions.Versions)-1]
	return latestVersion, nil
}

func GetLatestBuild() (string, error) {
	versions, err := FetchAPIData(url + "/projects/paper/")
	// https://api.papermc.io/v2/projects/paper/versions/1.21.1/builds/

	if err != nil {
		return "", err
	}

	latestVersion := versions.Versions[len(versions.Versions)-1]
	builds, err := FetchAPIData(url + "/projects/paper/versions/" + latestVersion + "/builds/")
	latestBuild := builds.Builds[len(builds.Builds)-1]

	if err != nil {
		return "", err
	}

	return latestBuild, nil
}

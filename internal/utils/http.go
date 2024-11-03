package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type APIResponse struct {
	Versions []string `json:"versions"`
	Builds   []Build  `json:"builds"`
}

type Build struct {
	Build     int       `json:"build"`
	Downloads Downloads `json:"downloads"`
	Channel   string    `json:"channel"`
}

type Downloads struct {
	Application Application `json:"application"`
}

type Application struct {
	Name string `json:"name"`
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

const BASE_URL = "https://api.papermc.io/v2"

func GetLatestStableVersion() (string, error) {
	versions, err := FetchAPIData(BASE_URL + "/projects/paper/")

	if err != nil {
		return err.Error(), err
	}

	for i := len(versions.Versions) - 1; i >= 0; i-- {
		version := versions.Versions[i]
		builds, err := FetchAPIData(BASE_URL + "/projects/paper/versions/" + version + "/builds/")

		if err != nil {
			return err.Error(), err
		}

		for j := len(builds.Builds) - 1; j >= 0; j-- {
			if builds.Builds[j].Channel == "default" {
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("No version with default channel builds found!")
}

// get version internally?
func GetLatestBuild(version string) (string, error) {
	builds, err := FetchAPIData(BASE_URL + "/projects/paper/versions/" + version + "/builds/")
	latestBuild := builds.Builds[len(builds.Builds)-1].Downloads.Application.Name

	if err != nil {
		return err.Error(), err
	}

	return latestBuild, nil
}

// get version internally?
func GetLatestBuildURL(version string) (string, error) {
	buildURL := BASE_URL + "/projects/paper/versions/" + version + "/builds/"
	builds, err := FetchAPIData(buildURL)
	latestBuild := builds.Builds[len(builds.Builds)-1]
	latestBuildURL := buildURL + strconv.Itoa(latestBuild.Build) + "/downloads/" + latestBuild.Downloads.Application.Name

	if err != nil {
		return err.Error(), err
	}

	return latestBuildURL, nil
}

func DownloadLatestBuild(version string) error {
	latestBuildURL, err := GetLatestBuildURL(version)
	if err != nil {
		return err
	}

	return downloadFile(latestBuildURL)
}

func downloadFile(url string) error {
	// dynamic download path later
	file := "./paper.jar"
	_, err := os.Stat(file)

	if errors.Is(err, os.ErrNotExist) {
		// File doesn't exist, safe to create
		out, err := os.Create(file)
		if err != nil {
			return err
		}
		defer out.Close()

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Error downloading: %s", resp.Status)
		}

		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	}

	return fmt.Errorf("File already exists! Refusing to overwrite.")
}

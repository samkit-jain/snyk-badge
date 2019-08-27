package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Handle the API request for badge creation.
// Hit Snyk's List Projects API and get a list of projects. Check if the caller has access to the repo referred to by
// username/repo and return a badge if access.
func badgeHandler(w http.ResponseWriter, r *http.Request, username string, repo string) {
	// Default shields.io badge URL
	badgeURL := "https://img.shields.io/badge/vulnerabilities-unknown-inactive"

	client := &http.Client{}

	// Generate the Snyk API URL
	apiURL := fmt.Sprintf("https://snyk.io/api/v1/org/%s/projects", os.Getenv("SNYK_ORG_ID"))

	// Setup the GET request
	req, _ := http.NewRequest("GET", apiURL, nil)

	// Setup the headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", os.Getenv("SNYK_API_KEY"))

	// Perform the request
	resp, err := client.Do(req)

	// Could not perform the request
	// Likely network issues
	if err != nil {
		log.Println("Errored when sending request to the Snyk server")
		writeBadge(w, badgeURL)
		return
	}

	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	// Non-200 response received
	// Likely the creds are wrong
	if resp.StatusCode != 200 {
		log.Println("Did not receive a 200 OK response from the Snyk server")
		writeBadge(w, badgeURL)
		return
	}

	// Perform JSON loading of the response
	var data map[string]interface{}
	err = json.Unmarshal([]byte(string(respBody)), &data)

	if err != nil {
		writeBadge(w, badgeURL)
		return
	}

	projects := data["projects"].([]interface{})

	// Check if the requested repo is available
	for _, project := range projects {
		project := project.(map[string]interface{})

		if strings.HasPrefix(project["name"].(string), username+"/"+repo+":") {
			// Count the number of issues
			issues := project["issueCountsBySeverity"].(map[string]interface{})
			highCount := int(issues["high"].(float64))
			mediumCount := int(issues["medium"].(float64))
			lowCount := int(issues["low"].(float64))

			totalIssues := highCount + mediumCount + lowCount

			if totalIssues == 0 {
				// No vulnerabilities
				badgeURL = "https://img.shields.io/badge/vulnerabilities-0-brightgreen"
			} else {
				// Vulnerabilities found
				badgeURL = fmt.Sprintf("https://img.shields.io/badge/vulnerabilities-%d-red", totalIssues)
			}

			break
		}
	}

	writeBadge(w, badgeURL)
	return
}

// Return the badge image from the shields.io URL
func writeBadge(w http.ResponseWriter, badgeURL string) {
	// Set the response header
	w.Header().Set("Content-Type", "image/svg+xml;charset=utf-8")

	client := &http.Client{}

	req, _ := http.NewRequest("GET", badgeURL, nil)
	resp, err := client.Do(req)

	// Could not perform the request
	// Likely network issues
	if err != nil {
		log.Println("Errored when sending request to the Shields server")
		fmt.Fprintf(w, badgeURL)
		return
	}

	defer resp.Body.Close()

	// Non-200 response received
	// Likely the service is down
	if resp.StatusCode != 200 {
		log.Println("Did not receive a 200 OK response from the Shields server")
		fmt.Fprintf(w, badgeURL)
		return
	}

	// Write the SVG image to the response object
	io.Copy(w, resp.Body)
}

var validPath = regexp.MustCompile("^/badge/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)/$")

func Handler(w http.ResponseWriter, r *http.Request) {
	// Path validation
	m := validPath.FindStringSubmatch(r.URL.Path)

	if m == nil {
		writeBadge(w, "https://img.shields.io/badge/vulnerabilities-unknown-inactive")
		return
	}

	badgeHandler(w, r, m[1], m[2])
}

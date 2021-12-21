package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	// UnknownURL string
	UnknownURL = "https://img.shields.io/badge/vulnerabilities-unknown-inactive?logo=snyk"
	// GreenURL string
	GreenURL = "https://img.shields.io/badge/vulnerabilities-0-brightgreen?logo=snyk"
)

var (
	// Version string
	Version = "1.0.0"
	// Commit string
	Commit = "no-data"
)

// Handle the API request for badge creation.
// Hit Snyk's List Projects API and get a list of projects. Check if the caller has access to the repo referred to by
// username/repo and return a badge if access.
func badgeHandler(w http.ResponseWriter, r *http.Request, username, repo string, projectID []string) {
	// Default shields.io badge URL
	badgeURL := UnknownURL

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

	// Check number of vulnerabilities based on search critiria
	// org + name without projectID
	// org + name + projectID
	// org + multiple projectIDs
	var totalIssues int
	name := fmt.Sprintf(username + "/" + repo + ":")
	switch len(projectID) {
	case 0:
		totalIssues = vulnerabilitiesFound(projects, name, "")
	case 1:
		totalIssues = vulnerabilitiesFound(projects, name, projectID[0])
	default:
		for _, id := range projectID {
			totalIssues = vulnerabilitiesFound(projects, name, id)
		}
	}
	if totalIssues == 0 {
		// No vulnerabilities
		badgeURL = GreenURL
	} else {
		// Vulnerabilities found
		// RedURL is created here based on number of vulnerabilities found
		badgeURL = fmt.Sprintf("https://img.shields.io/badge/vulnerabilities-%d-red?logo=snyk", totalIssues)
	}

	writeBadge(w, badgeURL)
	return
}

func vulnerabilitiesFound(projects []interface{}, name string, projectID string) int {
	var totalIssues int
	for _, project := range projects {
		project := project.(map[string]interface{})
		if strings.HasPrefix(project["name"].(string), name) || project["id"].(string) == projectID {
			totalIssues = countVulnerabilities(project)
			break
		}
	}
	return totalIssues
}

func countVulnerabilities(project map[string]interface{}) int {
	var criticalCount, highCount, mediumCount, lowCount, totalIssues int
	log.Println("project: ", project)
	// Count the number of issues
	issues := project["issueCountsBySeverity"].(map[string]interface{})
	// fmt.Println("full issues list: ",issues)
	log.Println("issues: ", issues["critical"], issues["high"], issues["medium"], issues["low"])
	criticalCount = int(issues["critical"].(float64))
	highCount = int(issues["high"].(float64))
	mediumCount = int(issues["medium"].(float64))
	lowCount = int(issues["low"].(float64))

	totalIssues = criticalCount + highCount + mediumCount + lowCount
	return totalIssues
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

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println("url path: ", r.URL.Path)
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		// if fails to parse URL parameters, just answer it quickly
		writeBadge(w, UnknownURL)
		return
	}
	// Parse URL parameters
	org := values.Get("org")
	name := values.Get("name")
	projectID := values["id"]
	if len(projectID) > 1 {
		log.Println("found more than one id: ", projectID)
	}
	// Required values
	// org is always required
	// name == repository name OR
	// projectID
	switch {
	case org == "":
		writeBadge(w, UnknownURL)
		return
	case name == "" || len(projectID) == 0:
		writeBadge(w, UnknownURL)
		return
	}
	// try to write a correct badge
	badgeHandler(w, r, org, name, projectID)
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("url path: ", r.URL.Path)
	message := fmt.Sprintf("{\"version\": \"%s\", \"commit\":\"%s\"}", Version, Commit)
	fmt.Fprint(w, message)
}

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}

	http.HandleFunc("/api/badges", Handler)
	http.HandleFunc("/api/version", versionHandler)
	log.Printf("Listen on %s. Go to https://127.0.0.1%s/", listenAddr, listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

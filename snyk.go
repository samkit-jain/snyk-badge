package main

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

func badgeHandler(w http.ResponseWriter, r *http.Request, username string, repo string) {
	badgeUrl := "https://img.shields.io/badge/vulnerabilities-unknown-inactive"

	client := &http.Client{}

	apiUrl := fmt.Sprintf("https://snyk.io/api/v1/org/%s/projects", os.Getenv("SNYK_ORG_ID"))

	req, _ := http.NewRequest("GET", apiUrl, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", os.Getenv("SNYK_API_KEY"))

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		writeBadge(w, badgeUrl)
		return
	}

	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Println("Did not receive a 200 OK response from the server")
		writeBadge(w, badgeUrl)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(string(respBody)), &data)

	if err != nil {
		writeBadge(w, badgeUrl)
		return
	}

	projects := data["projects"].([]interface{})

	for _, project := range projects {
		project := project.(map[string]interface{})

		if strings.HasPrefix(project["name"].(string), username+"/"+repo) {
			issues := project["issueCountsBySeverity"].(map[string]interface{})
			highCount := int(issues["high"].(float64))
			mediumCount := int(issues["medium"].(float64))
			lowCount := int(issues["low"].(float64))

			totalIssues := highCount + mediumCount + lowCount
			if totalIssues == 0 {
				badgeUrl = "https://img.shields.io/badge/vulnerabilities-0-brightgreen"
			} else {
				badgeUrl = fmt.Sprintf("https://img.shields.io/badge/vulnerabilities-%d-red", totalIssues)
			}

			break
		}
	}

	writeBadge(w, badgeUrl)
	return
}

func writeBadge(w http.ResponseWriter, badgeUrl string) {
	w.Header().Set("Content-Type", "image/svg+xml;charset=utf-8")
	client := &http.Client{}

	req, _ := http.NewRequest("GET", badgeUrl, nil)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Errored when sending request to the server")
		fmt.Fprintf(w, badgeUrl)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Did not receive a 200 OK response from the server")
		fmt.Fprintf(w, badgeUrl)
		return
	}

	io.Copy(w, resp.Body)
}

var validPath = regexp.MustCompile("^/badge/([a-zA-Z0-9-]+)/([a-zA-Z0-9-]+)/$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)

		if m == nil {
			http.NotFound(w, r)
			return
		}

		fn(w, r, m[1], m[2])
	}
}

func main() {
	http.HandleFunc("/badge/", makeHandler(badgeHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

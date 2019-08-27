# snyk-badge 
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/f3f9b3d2b4364331a9afceb346a1f7bd)](https://app.codacy.com/app/samkit-jain/snyk-badge?utm_source=github.com&utm_medium=referral&utm_content=samkit-jain/snyk-badge&utm_campaign=Badge_Grade_Dashboard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Snyk badge generator for private GitHub repositories.

As of August 2019, [Snyk](https://snyk.io/) badges currently only work for public npm packages and GitHub repositories, and will fail if pointed at a private repository. This Go app aims to solve that problem by providing badges for private repositories.

**NOTE:** Will only work for repositories you have integrated in Snyk.

## Setup
1. Integrate Snyk with your GitHub account
2. Install Go
3. Set environment variables
    ```
    SNYK_ORG_ID="Your Snyk Organisation ID"
    SNYK_API_KEY="Your Snyk API key"
    ```
4. Run `go build snyk.go && ./snyk`
5. Visit http://localhost:8080/badge/{username}/{repo_name}/ (Replace `{username}` and `{repo_name}` with your own GitHub username and the private repository you have access to, respectively)

**Note:** Trailing `/` is mandatory.

**Note:** Directory `api/` is for serverless deployment on now.sh

## How it works
Hits the [List All Projects](https://snyk.docs.apiary.io/#reference/projects/all-projects/list-all-projects) API and gets a list of all the projects in your organisation. Searches for the repo you mentioned in the URL and counts the number of issues in it. If the number of issues is 0, gives a green badge. If more than 0, gives a red badge with the total number of issues as the value. If access unavailable gives a grey badge.

## Badge generation
Badges are generated with the help of the awesome [Shields](https://github.com/badges/shields) project. Badges look like
* <img src="https://img.shields.io/badge/vulnerabilities-0-brightgreen" alt="no vulnerabilities"/>
* <img src="https://img.shields.io/badge/vulnerabilities-10-red" alt="10 vulnerabilities"/>
* <img src="https://img.shields.io/badge/vulnerabilities-unknown-inactive" alt="vulnerabilities unknown"/>

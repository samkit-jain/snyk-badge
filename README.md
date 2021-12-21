# snyk-badge 
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
5. Visit http://localhost:8080/badge/?org={username}&name={repo_name} (Replace `{username}` and `{repo_name}` with your own GitHub username and the private repository you have access to, respectively)

**Note:** You can use http://localhost:8080/badge/?org={username}&name={repo_name}&id={project_id_snyk} to be more precisely which repository you want to have a badge. Or you can sum multiple ids: http://localhost:8080/badge/?org={username}&name={repo_name}&id={project_id_snyk}&id={another_project_id_snyk}

## How it works
Hits the [List All Projects](https://snyk.docs.apiary.io/#reference/projects/all-projects/list-all-projects) API and gets a list of all the projects in your organisation. Searches for the repo you mentioned in the URL and counts the number of issues in it. If the number of issues is 0, gives a green badge. If more than 0, gives a red badge with the total number of issues as the value. If access unavailable gives a grey badge.

## Badge generation
Badges are generated with the help of the awesome [Shields](https://github.com/badges/shields) project. Badges look like
* <img src="https://img.shields.io/badge/vulnerabilities-0-brightgreen?logo=snyk" alt="no vulnerabilities"/>
* <img src="https://img.shields.io/badge/vulnerabilities-10-red?logo=snyk" alt="10 vulnerabilities"/>
* <img src="https://img.shields.io/badge/vulnerabilities-unknown-inactive?logo=snyk" alt="vulnerabilities unknown"/>

## Azure config and deploy

Create a resource group:
```bash
az group create --name snykbadges-group --location eastus
```
Create a storage:
```bash 
az storage account create --name snykbadgessvc --location eastus --resource-group snykbadges-group --sku Standard_LRS
```
Create a function:
```bash
az functionapp create --name snykbadgessvc --storage-account snykbadgessvc --consumption-plan-location eastus --resource-group snykbadges-group --runtime custom --os-type Linux --functions-version 3
```

Add in functions settings in Azure Portal:
```
SNYK_API_KEY=asdasd...
SNYK_ORG_ID=adsasd...
```
And Save it.

Generate binary for linux
```bash
GOOS=linux GOARCH=amd64 go build snyk.go
```
Deploy it using func binary:
```bash
func azure functionapp publish snykbadgessvc
```


## References

https://docs.microsoft.com/en-us/azure/azure-functions/create-first-function-vs-code-other  

https://www.hildeberto.com/2021/01/azure-function-golang-2.html  

https://acloudguru.com/blog/engineering/how-to-build-a-serverless-app-using-go-and-azure-functions  
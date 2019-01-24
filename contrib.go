package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Structs
type CONTRIBYDAY struct {
	DayIndex      int
	Contriubtions int
}

type CONTRIBYDAYS []CONTRIBYDAY

// UserData useful data to pass around
type UserData struct {
	Owner    string
	UserName string
	PassWord string
	Emails   []string
}

// InputParams struct to carry needed input parameters
type InputParams struct {
	StartDate time.Time
	EndDate   time.Time
}

//BuildGithubApiURI( user UserData, params InputParams, uri UriEnum) string {
//
//	return ""
//}

// BuildRepositoryListCmd constucts command for getting list of repositories
func BuildRepositoryListCmd(owner string) string {
	uri := "https://api.github.com/users/"
	uri += owner
	uri += "/repos?per_page=100"
	return uri
}

// RetrieveGitHubData - obtains data based on endpoint uri and returns results
func RetrieveGitHubData(user UserData, uri string, useAuth bool) []byte {

	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)

	if useAuth {
		req.SetBasicAuth(user.UserName, user.PassWord)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil
		//panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			//panic(err2)
			return nil
		}
		return bodyBytes
	}
	return nil

}

// GetContributionsIssues - returns the day and contributions
func GetContributionsIssues(user UserData, params InputParams) []CONTRIBYDAY {
	sURI := "https://api.github.com/issues?filter=created"
	bodyBytes := RetrieveGitHubData(user, sURI, true)

	_debug := false
	if bodyBytes != nil {
		var issues ISSUESPERUSER

		if _debug {
			err := ioutil.WriteFile("./ISSUES1", []byte(bodyBytes), 0644)
			checkerror(err)
		}

		if _debug {
			fmt.Printf("len bodyBytes: %d\n", len(bodyBytes))
		}

		err := json.Unmarshal([]byte(bodyBytes), &issues)
		if err != nil {
			panic(err)
		}

		c := []CONTRIBYDAY{}
		fmt.Printf("num issues %d\n", len(issues))
		for i := 0; i < len(issues); i++ {
			if !issues[i].Repository.Fork {
				day := DaysBetween(params.StartDate, issues[i].CreatedAt)
				fmt.Printf("ID: %d, DateTime: %s, Duration(d): %d\n", issues[i].ID, issues[i].CreatedAt, day)
				c = append(c, CONTRIBYDAY{day, 1})
			}
		}
		return c
	} else {
		return nil
	}
}

// GetContributionsPulls : gets contributions via pull requests
func GetContributionsPulls(user UserData, params InputParams) []CONTRIBYDAY {

	sURI := "https://api.github.com/pulls?filter=created"
	bodyBytes := RetrieveGitHubData(user, sURI, true)

	_debug := false
	if bodyBytes != nil {

		if _debug {
			err := ioutil.WriteFile("./PULLS1", []byte(bodyBytes), 0644)
			checkerror(err)
		}

		if _debug {
			fmt.Printf("len bodyBytes: %d\n", len(bodyBytes))
		}

		var pulls PULLREQUESTPERUSER
		err := json.Unmarshal([]byte(bodyBytes), &pulls)
		if err != nil {
			panic(err)
		}

		c := []CONTRIBYDAY{}
		for i := 0; i < len(pulls); i++ {
			if true {
				day := DaysBetween(params.StartDate, pulls[i].CreatedAt)
				fmt.Printf("ID: %d, DateTime: %s, Duration(d): %d\n", pulls[i].ID, pulls[i].CreatedAt, day)
				c = append(c, CONTRIBYDAY{day, 1})
			}
		}
		return c
	} else {
		return nil
	}
}

// GetRepoPrefixCmd returns URI for requesting commits from the rep since given date
func GetRepoPrefixCmd(owner string, repoName string, params InputParams) string {
	uri := "https://api.github.com/repos/"
	uri += owner // owner
	uri += "/"
	uri += repoName // repos name
	return uri
}

// BuildRepoURICmd returns URI for requesting commits from the rep since given date
func BuildRepoURICmd(owner string, repoName string, cmd string, params InputParams) string {
	uri := GetRepoPrefixCmd(owner, repoName, params)
	uri += "/"
	uri += cmd
	uri += "?since="
	uri += params.StartDate.Format(time.RFC3339)
	uri += "&per_page=100"
	return uri
}

// BuildUserEmailURI get user email
func BuildUserEmailURI() string {
	uri := "https://api.github.com/user/email"
	return uri
}

// ValidEmail check if commit email match one of the user's emails
func ValidEmail(user UserData, commitEmail string) bool {

	for _, em := range user.Emails {
		//fmt.Printf("%s, %s\n", commitEmail, em)
		if strings.Compare(em, commitEmail) == 0 {
			//fmt.Printf("Match\n")
			return true
		}
	}
	return false
}

// ValidCommitContribution logic to determine if a commit counts as a contribution
func ValidCommitContribution(commit COMMIT, repo REPO, user UserData) bool {

	bok := false
	if (ValidEmail(user, strings.ToLower(commit.Commit.Author.Email))) && (!repo.Fork) && (repo.DefaultBranch == "master") {
		//if (!repo.Fork) && (repo.DefaultBranch == "master") {
		bok = true
	}
	return bok
}

// CountCommitsRepo loops through commints and checks if a contribution
func CountCommitsRepo(user UserData, repo REPO, params InputParams) []CONTRIBYDAY {

	contribs := make([]CONTRIBYDAY, 0)
	if !repo.Fork {
		sURI := BuildRepoURICmd(user.Owner, repo.Name, "commits", params)
		bodyBytes := RetrieveGitHubData(user, sURI, true)

		if len(bodyBytes) > 0 {
			var commits COMMITS

			err := json.Unmarshal([]byte(bodyBytes), &commits)
			checkerror(err)

			ncom := len(commits)
			if ncom > 0 {
				for j := 0; j < ncom; j++ {
					// Check criteria for commits
					if ValidCommitContribution(commits[j], repo, user) {
						day := DaysBetween(params.StartDate, commits[j].Commit.Author.Date)
						c := CONTRIBYDAY{day, 1}
						contribs = append(contribs, c)
					}
				}
			}
		}
	}
	return contribs
}

// GetContributionsCommits : gets contributions via commits
func GetContributionsCommits(user UserData, params InputParams) []CONTRIBYDAY {
	fmt.Println("GetContributionsCommits")
	sURI := BuildRepositoryListCmd(user.Owner)
	fmt.Println(sURI)
	bodyBytes := RetrieveGitHubData(user, sURI, true)
	if bodyBytes == nil {
		fmt.Println("nil bodyBytes")
	}
	if bodyBytes != nil {

		var repos REPOS
		err := json.Unmarshal([]byte(bodyBytes), &repos)
		checkerror(err)

		// Setup jobs
		nrepos := len(repos)
		jobs := make(chan REPO, nrepos)
		ans := make(chan CONTRIBYDAYS, nrepos)
		done := make(chan bool)
		go func() {
			for {
				r, more := <-jobs
				if more {
					//s := CountCommitsRepo(user, repos[i], params)
					s := CountCommitsRepo(user, r, params)
					ans <- s
				} else {
					done <- true
					return
				}
			}
		}()

		// Send out jobs
		for i := 0; i < nrepos; i++ {
			fmt.Println(repos[i].Name)
			jobs <- repos[i]
		}
		close(jobs)

		// Wait for all jobs to be done
		<-done
		results := []CONTRIBYDAY{}
		for i := 0; i < nrepos; i++ {

			results = append(results, <-ans...)
		}
		return results
	} else {
		return nil
	}
}

// GetUserInfo fills out UserData information
func GetUserInfo(user *UserData) {
	sURI := "https://api.github.com/user/emails"
	bodyBytes := RetrieveGitHubData(*user, sURI, true)
	fmt.Printf("getemails bb: %d\n", len(bodyBytes))
	var emails EMAILS
	err := json.Unmarshal([]byte(bodyBytes), &emails)
	checkerror(err)

	for _, email := range emails {
		tmp := strings.ToLower(strings.TrimRight(email.Email, "\n"))
		user.Emails = append(user.Emails, tmp)
	}

}

func main() {

	//	user := UserData{Owner: <username>, UserName: <username>, PassWord: <password>}
	user := UserData{Owner: "vmg", UserName: "vmg", PassWord: "DontKnow"}

	// Get information for later
	GetUserInfo(&user)

	// Current time
	t1 := time.Now()

	// 365 days ago <- -1 year + 1 day
	t0 := t1.AddDate(0, 0, -356)
	params := InputParams{StartDate: t0, EndDate: t1}

	var wg sync.WaitGroup

	wg.Add(3)
	contribIssues := []CONTRIBYDAY{}
	go func() {
		defer wg.Done()
		contribIssues = GetContributionsIssues(user, params)
	}()

	contribPulls := []CONTRIBYDAY{}
	go func() {
		defer wg.Done()
		contribPulls = GetContributionsPulls(user, params)
	}()

	contribCommits := []CONTRIBYDAY{}

	go func() {
		defer wg.Done()
		contribCommits = GetContributionsCommits(user, params)
	}()

	wg.Wait()

	// Convert to array
	var vContribs [365]int
	for i := 0; i < 365; i++ {
		vContribs[i] = 0
	}

	for _, c := range contribCommits {
		vContribs[c.DayIndex] += c.Contriubtions
	}

	for _, c := range contribIssues {
		vContribs[c.DayIndex]++
	}

	for _, c := range contribPulls {
		vContribs[c.DayIndex]++
	}

	fmt.Printf("[")
	for i := 0; i < len(vContribs); i++ {
		fmt.Printf("%d", vContribs[i])
		if i < len(vContribs)-1 {
			fmt.Printf(", ")
		}
	}
	fmt.Printf("]")
	fmt.Printf("\n")

}

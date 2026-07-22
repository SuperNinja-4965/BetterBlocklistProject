package manage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	h "better-blocklist/src/internal/helpers"
	t "better-blocklist/src/internal/terminal"

	tea "github.com/charmbracelet/bubbletea"
)

// GitHubIssueMenu asks for a GitHub issue URL (and optional token) and
// processes the issue body to add or remove domains from lists.
func GitHubIssueMenu() tea.Cmd {
	return t.RequestInput("Enter GitHub issue URL (https://github.com/owner/repo/issues/123):", "", func(issueURL string) tea.Cmd {
		return t.RequestInput("Enter GitHub token (leave blank for unauthenticated):", "", func(token string) tea.Cmd {
			return func() tea.Msg {
				out := processIssueURL(issueURL, token)
				return t.OutputLine{Line: out}
			}
		}, "Process issue")
	}, "GitHub Issue")
}

func processIssueURL(issueURL, token string) string {
	owner, repo, number, err := parseIssueURL(issueURL)
	if err != nil {
		return "Invalid issue URL: " + err.Error()
	}

	body, err := fetchIssueBody(owner, repo, number, token)
	if err != nil {
		return "Failed to fetch issue: " + err.Error()
	}

	adds, removes := parseIssueBody(body)

	var results []string

	// Process removes
	for _, r := range removes {
		listFile := resolveListFilename(r.List)
		exists, err := h.FileExists(listFile)
		if err != nil || !exists {
			results = append(results, fmt.Sprintf("Remove %s -> list %s: list not found", r.Domain, r.List))
			continue
		}
		if err := removeFromFile(listFile, r.Domain); err != nil {
			results = append(results, fmt.Sprintf("Remove %s -> %s: %v", r.Domain, r.List, err))
		} else {
			results = append(results, fmt.Sprintf("Removed %s from %s", r.Domain, r.List))
		}
	}

	// Process adds
	for _, a := range adds {
		listFile := resolveListFilename(a.List)
		exists, err := h.FileExists(listFile)
		if err != nil || !exists {
			results = append(results, fmt.Sprintf("Add %s -> list %s: list not found", a.Domain, a.List))
			continue
		}
		if err := addToFile(listFile, a.Domain); err != nil {
			results = append(results, fmt.Sprintf("Add %s -> %s: %v", a.Domain, a.List, err))
		} else {
			results = append(results, fmt.Sprintf("Added %s to %s", a.Domain, a.List))
		}
	}

	if len(results) == 0 {
		return "No actionable items found in issue."
	}
	return strings.Join(results, "\n")
}

func parseIssueURL(u string) (owner, repo, number string, err error) {
	// Expect https://github.com/{owner}/{repo}/issues/{number}
	// Trim trailing slash
	u = strings.TrimSpace(u)
	u = strings.TrimSuffix(u, "/")
	parts := strings.Split(u, "/")
	if len(parts) < 5 {
		return "", "", "", fmt.Errorf("unexpected URL format")
	}
	// find "issues" segment
	for i := 0; i < len(parts)-2; i++ {
		if parts[i] == "issues" {
			if i-2 >= 0 {
				owner = parts[i-2]
				repo = parts[i-1]
				number = parts[i+1]
				return owner, repo, number, nil
			}
		}
	}
	return "", "", "", fmt.Errorf("could not parse owner/repo/number")
}

func fetchIssueBody(owner, repo, number, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%s", owner, repo, number)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var j struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil {
		return "", err
	}
	return j.Body, nil
}

type action struct{ Domain, List string }

func parseIssueBody(body string) (adds, removes []action) {
	lines := strings.Split(body, "\n")

	var removeDomains, removeLists []string
	var addDomains, addLists []string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "###") {
			continue
		}
		header := strings.TrimSpace(strings.TrimPrefix(line, "###"))
		headerLower := strings.ToLower(header)

		// collect block lines until next heading or EOF
		j := i + 1
		var block []string
		for j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "###") {
			block = append(block, lines[j])
			j++
		}
		// advance i to j-1
		i = j - 1

		first := firstNonEmptyLine(strings.Join(block, "\n"))
		switch headerLower {
		case "domain to remove":
			if first != "" {
				removeDomains = append(removeDomains, first)
			}
		case "current blocklist":
			if first != "" {
				removeLists = append(removeLists, first)
			}
		case "domain to add":
			if first != "" {
				addDomains = append(addDomains, first)
			}
		case "target blocklist":
			if first != "" {
				addLists = append(addLists, first)
			}
		}
	}

	for i := 0; i < len(removeDomains) && i < len(removeLists); i++ {
		removes = append(removes, action{Domain: removeDomains[i], List: removeLists[i]})
	}
	for i := 0; i < len(addDomains) && i < len(addLists); i++ {
		adds = append(adds, action{Domain: addDomains[i], List: addLists[i]})
	}

	return adds, removes
}

func firstNonEmptyLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			return t
		}
	}
	return ""
}

func resolveListFilename(list string) string {
	list = strings.TrimSpace(list)
	if list == "" {
		return ""
	}
	dir := fmt.Sprintf("%sLists/", h.GetCurrentDir())

	// If the user supplied a filename with extension, prefer that exact file.
	candidates := []string{
		list,
		list + ".txt",
		list + ".txt.gz",
		list + ".ip",
		list + ".ip.gz",
	}

	for _, c := range candidates {
		path := dir + c
		if exists, _ := h.FileExists(path); exists {
			return path
		}
	}

	// Fallback to .txt (will later be written as .txt.gz)
	return dir + list + ".txt"
}

// --- Repository-level issue processing ---

// GitHubIssuesMenu asks for an owner/repo and token, lists open issues,
// lets the user select one, shows parsed actions, and asks confirmation.
func GitHubIssuesMenu() tea.Cmd {
	return t.RequestInput("Enter repository (owner/repo):", "", func(repo string) tea.Cmd {
		return t.RequestInput("Enter GitHub token (leave blank for unauthenticated):", "", func(token string) tea.Cmd {
			return func() tea.Msg {
				// fetch issues
				ownerRepo := strings.TrimSpace(repo)
				parts := strings.Split(ownerRepo, "/")
				if len(parts) != 2 {
					return t.OutputLine{Line: "Invalid owner/repo format"}
				}
				owner, repoName := parts[0], parts[1]
				issues, err := fetchRepoIssues(owner, repoName, token)
				if err != nil {
					return t.OutputLine{Line: "Failed to fetch issues: " + err.Error()}
				}

				if len(issues) == 0 {
					return t.OutputLine{Line: "No open issues found."}
				}

				// Build choices and open submenu
				var choices []string
				for _, is := range issues {
					choices = append(choices, fmt.Sprintf("#%d %s", is.Number, is.Title))
				}

				return t.Submenu(fmt.Sprintf("Select issue for %s/%s", owner, repoName), choices, func(selected string) tea.Cmd {
					// find selected issue
					idx := -1
					for i, c := range choices {
						if c == selected {
							idx = i
							break
						}
					}
					if idx < 0 {
						return func() tea.Msg { return t.OutputLine{Line: "Unknown selection"} }
					}
					issue := issues[idx]

					// Show parsed actions
					adds, removes := parseIssueBody(issue.Body)
					var summary []string
					if len(removes) > 0 {
						for _, r := range removes {
							summary = append(summary, fmt.Sprintf("Remove %s from %s", r.Domain, r.List))
						}
					}
					if len(adds) > 0 {
						for _, a := range adds {
							summary = append(summary, fmt.Sprintf("Add %s to %s", a.Domain, a.List))
						}
					}
					if len(summary) == 0 {
						summary = append(summary, "No actionable items found in issue.")
					}

					// Ask for confirmation
					return t.RequestInput(fmt.Sprintf("Issue #%d summary:\n%s\n\nType 'yes' to apply, 'no' to close (closing requires token).", issue.Number, strings.Join(summary, "\n")), "", func(answer string) tea.Cmd {
						ans := strings.ToLower(strings.TrimSpace(answer))
						if ans == "yes" || ans == "y" {
							return func() tea.Msg {
								out := applyIssueByNumber(owner, repoName, issue.Number, token)
								return t.OutputLine{Line: out}
							}
						}

						// No -> close issue with comment (requires token)
						if token == "" {
							return func() tea.Msg { return t.OutputLine{Line: "Closing issues disabled without token."} }
						}

						// Ask for close comment
						return t.RequestInput("Enter comment to post when closing the issue:", "", func(comment string) tea.Cmd {
							return func() tea.Msg {
								if err := closeIssueWithComment(owner, repoName, issue.Number, token, comment); err != nil {
									return t.OutputLine{Line: "Failed to close issue: " + err.Error()}
								}
								return t.OutputLine{Line: "Issue closed and commented."}
							}
						}, "Close issue")
					}, fmt.Sprintf("Issue #%d summary:\n%s", issue.Number, strings.Join(summary, "\n")))
				})()
			}
		}, "GitHub Token")
	}, "Repository")
}

type ghIssue struct {
	Number      int                    `json:"number"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	PullRequest map[string]interface{} `json:"pull_request,omitempty"`
}

func fetchRepoIssues(owner, repo, token string) ([]ghIssue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open&per_page=100", owner, repo)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var items []ghIssue
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	// filter out pull requests
	var out []ghIssue
	for _, it := range items {
		if it.PullRequest != nil {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

func applyIssueByNumber(owner, repo string, number int, token string) string {
	body, err := fetchIssueBody(owner, repo, fmt.Sprintf("%d", number), token)
	if err != nil {
		return "Failed to fetch issue body: " + err.Error()
	}
	adds, removes := parseIssueBody(body)

	var results []string
	// process removes then adds
	for _, r := range removes {
		listFile := resolveListFilename(r.List)
		exists, err := h.FileExists(listFile)
		if err != nil || !exists {
			results = append(results, fmt.Sprintf("Remove %s -> list %s: list not found", r.Domain, r.List))
			continue
		}
		if err := removeFromFile(listFile, r.Domain); err != nil {
			results = append(results, fmt.Sprintf("Remove %s -> %s: %v", r.Domain, r.List, err))
		} else {
			results = append(results, fmt.Sprintf("Removed %s from %s", r.Domain, r.List))
		}
	}
	for _, a := range adds {
		listFile := resolveListFilename(a.List)
		exists, err := h.FileExists(listFile)
		if err != nil || !exists {
			results = append(results, fmt.Sprintf("Add %s -> list %s: list not found", a.Domain, a.List))
			continue
		}
		if err := addToFile(listFile, a.Domain); err != nil {
			results = append(results, fmt.Sprintf("Add %s -> %s: %v", a.Domain, a.List, err))
		} else {
			results = append(results, fmt.Sprintf("Added %s to %s", a.Domain, a.List))
		}
	}

	if len(results) == 0 {
		return "No actionable items found in issue."
	}
	return strings.Join(results, "\n")
}

func closeIssueWithComment(owner, repo string, number int, token, comment string) error {
	if token == "" {
		return fmt.Errorf("token required to close issues")
	}
	// Post comment
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, number)
	body := fmt.Sprintf(`{"body": %q}`, comment)
	req, _ := http.NewRequest("POST", commentURL, strings.NewReader(body))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to post comment: %s", resp.Status)
	}

	// Close issue
	closeURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repo, number)
	req, _ = http.NewRequest("PATCH", closeURL, strings.NewReader(`{"state":"closed"}`))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to close issue: %s", resp.Status)
	}
	return nil
}

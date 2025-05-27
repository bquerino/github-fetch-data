
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	CSV_INPUT   = "repos.csv" // Arquivo CSV de entrada com IDs dos repositórios - deve ser ter uma coluna com repository_id e todos os ids
	CSV_OUTPUT  = "resultado.csv" // Arquivo CSV de saída com os resultados
	GITHUB_TOKEN = "ghp_xxxxx..." // Substitua com seu token
)

type Repository struct {
	ID            int    `json:"id"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
}

type Commit struct {
	Commit struct {
		Author struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commit"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
}

func main() {
	repoIDs, err := readRepositoryIDs(CSV_INPUT)
	if err != nil {
		panic(err)
	}

	total := len(repoIDs)
	if total == 0 {
		fmt.Println("Nenhum ID encontrado no arquivo de entrada.")
		return
	}

	outputFile, err := os.Create(CSV_OUTPUT)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	writer.Write([]string{"repository_id", "repository_name", "author_name", "author_email"})

	start := time.Now()

	for i, id := range repoIDs {
		fmt.Printf("🔍 [%d/%d] Processando ID: %s
", i+1, total, id)

		repo, err := fetchRepoByID(id)
		if err != nil {
			fmt.Printf("Erro ao buscar repositório %s: %v
", id, err)
			continue
		}

		commit, err := fetchLastCommit(repo.FullName, repo.DefaultBranch)
		if err != nil {
			fmt.Printf("Erro ao buscar commit de %s: %v
", repo.FullName, err)
			continue
		}

		authorName := commit.Commit.Author.Name
		if commit.Author != nil && commit.Author.Login != "" {
			authorName = commit.Author.Login
		}

		writer.Write([]string{id, repo.FullName, authorName, commit.Commit.Author.Email})

		// Exibe progresso e estimativa
		elapsed := time.Since(start)
		remaining := time.Duration((float64(elapsed) / float64(i+1)) * float64(total-i-1))
		fmt.Printf("✅ [%d/%d] %s - ETA: %s
", i+1, total, repo.FullName, remaining.Round(time.Second))
	}
}

func readRepositoryIDs(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var ids []string
	for i, row := range records {
		if i == 0 {
			continue // Skip header
		}
		if len(row) > 0 {
			ids = append(ids, row[0])
		}
	}
	return ids, nil
}

func fetchRepoByID(id string) (*Repository, error) {
	url := fmt.Sprintf("https://api.github.com/repositories/%s", id)
	resp, err := makeGitHubRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, err
	}
	return &repo, nil
}

func fetchLastCommit(fullName, branch string) (*Commit, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", fullName, branch)
	resp, err := makeGitHubRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var commit Commit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return nil, err
	}
	return &commit, nil
}

func makeGitHubRequest(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token " + GITHUB_TOKEN)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status %d for %s", resp.StatusCode, url)
	}
	return resp, nil
}

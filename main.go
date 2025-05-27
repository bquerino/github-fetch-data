package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	CSV_INPUT    = "repos.csv"
	CSV_OUTPUT   = "resultado.csv"
	GITHUB_TOKEN = "ghp_xxxxx..." // Substitua com seu token
	MAX_WORKERS  = 5
)

type Repository struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
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
	fmt.Printf("🔎 Total de repository_ids: %d\n", total)

	outputFile, err := os.Create(CSV_OUTPUT)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()
	writer.Write([]string{"repository_id", "repository_name", "author_name", "author_email"})

	var wg sync.WaitGroup
	tasks := make(chan string, MAX_WORKERS)

	// Workers
	for i := 0; i < MAX_WORKERS; i++ {
		go func(workerID int) {
			for repoID := range tasks {
				processRepository(repoID, writer)
				wg.Done()
			}
		}(i)
	}

	start := time.Now()
	for _, id := range repoIDs {
		wg.Add(1)
		tasks <- id
	}
	wg.Wait()
	close(tasks)

	fmt.Printf("✅ Concluído em %s\n", time.Since(start).Round(time.Second))
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

func processRepository(id string, writer *csv.Writer) {
	// Busca repositório
	repoURL := fmt.Sprintf("https://api.github.com/repositories/%s", id)
	resp, err := makeGitHubRequest(repoURL)
	if err != nil {
		writer.Write([]string{id, "Not Found", "-", "-"})
		writer.Flush()
		fmt.Printf("⚠️  [%s] Repositório não encontrado ou erro: %v\n", id, err)
		return
	}
	defer resp.Body.Close()

	var repo Repository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		writer.Write([]string{id, "Erro ao decodificar", "-", "-"})
		writer.Flush()
		fmt.Printf("❌ [%s] Erro ao decodificar repositório: %v\n", id, err)
		return
	}

	// Busca commit
	commitURL := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", repo.FullName, repo.DefaultBranch)
	resp2, err := makeGitHubRequest(commitURL)
	if err != nil {
		writer.Write([]string{id, repo.FullName, "Commit não encontrado", "-"})
		writer.Flush()
		fmt.Printf("⚠️  [%s] Commit não encontrado: %v\n", repo.FullName, err)
		return
	}
	defer resp2.Body.Close()

	var commit Commit
	if err := json.NewDecoder(resp2.Body).Decode(&commit); err != nil {
		writer.Write([]string{id, repo.FullName, "Erro no commit", "-"})
		writer.Flush()
		fmt.Printf("❌ [%s] Erro ao buscar commit: %v\n", repo.FullName, err)
		return
	}

	authorName := commit.Commit.Author.Name
	if commit.Author != nil && commit.Author.Login != "" {
		authorName = commit.Author.Login
	}

	writer.Write([]string{
		id,
		repo.FullName,
		authorName,
		commit.Commit.Author.Email,
	})
	writer.Flush()

	fmt.Printf("✅ [%s] Último commit por %s\n", repo.FullName, authorName)
}

func makeGitHubRequest(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+GITHUB_TOKEN)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("404 not found: %s", url)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return resp, nil
}

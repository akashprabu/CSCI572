package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// UserAgent to mimic browser requests
var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"

// SearchEngine contains details for each search engine
type SearchEngine struct {
	Name     string
	BaseURL  string
	Selector string
}

// QueryResults stores the scraped results for each query
type QueryResults map[string][]string

// Engines data for Bing, Yahoo!, Ask, and DuckDuckGo
var engines = []SearchEngine{
		{"Bing", "https://www.bing.com/search?q=", "li.b_algo h2 a"},
	// {"Yahoo!", "https://search.yahoo.com/search?p=", "a.ac-algo.fz-l.ac-21th.lh-24"},
	// {"Ask", "https://www.ask.com/web?q=", "div.PartialSearchResults-item-title a"},
	// {"DuckDuckGo", "https://duckduckgo.com/html/?q=", "a.result__a"},
}

// Scrape performs the web scraping for a given query and search engine
func (s SearchEngine) Scrape(query string) []string {
	// Random delay between 10 to 100 seconds
	// randSleep()

	// Construct the search URL
	tempURL := strings.Join(strings.Split(query, " "), "+")
	url := s.BaseURL + tempURL

	// Create the HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", userAgent)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return s.scrapeSearchResults(doc)
}

// scrapeSearchResults parses and extracts the top 10 results
func (s SearchEngine) scrapeSearchResults(doc *goquery.Document) []string {
	var results []string
	seen := make(map[string]bool)

	// Use the specified selector to find search result links
	doc.Find(s.Selector).Each(func(i int, item *goquery.Selection) {
		if len(results) >= 10 {
			return
		}
		// Extract the link from the anchor tag
		link, exists := item.Attr("href")
		if exists && !seen[link] {
			seen[link] = true
			results = append(results, link)
		}
	})

	return results
}

// randSleep pauses execution for a random time between 10 and 100 seconds
func randSleep() {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Intn(90)+10) * time.Second)
}

// SaveResultsToJSON saves the query results to a JSON file
func SaveResultsToJSON(filename string, results QueryResults) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		log.Fatalf("Failed to write to file: %s", err)
	}
}

// ReadQueriesFromFile reads queries line by line from a text file
func ReadQueriesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var queries []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query != "" {
			queries = append(queries, query)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return queries, nil
}

func main() {
	queryFile := "100QueriesSet1.txt"
	queries, err := ReadQueriesFromFile(queryFile)
	if err != nil {
		log.Fatalf("Error reading queries from file: %s", err)
	}

	queryResults := make(QueryResults)

	for _, query := range queries {
		for _, engine := range engines {
			fmt.Printf("Scraping %s for query: %s\n", engine.Name, query)
			results := engine.Scrape(query)
			key := fmt.Sprintf("%s", query)
			queryResults[key] = results
		}
	}

	SaveResultsToJSON("hw2.json", queryResults)
	fmt.Println("Results saved to hw2.json")
	compare()
}

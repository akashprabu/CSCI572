package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Function to read a JSON file containing your search results
func readScrapedResults(filename string) (map[string][]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var results map[string][]string
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Function to read a Google JSON file containing results as a map
func readGoogleJSON(filename string) (map[string][]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var results map[string][]string
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Function to calculate overlap between two sets of URLs
func calculateOverlap(yourResults, googleResults []string) (float64, float64) {
	overlapCount := 0
	for _, url := range yourResults {
		for _, gUrl := range googleResults {
			if url == gUrl {
				overlapCount++
				break
			}
		}
	}
	// Overlap percentage: (overlapCount / total Google results) * 100
	return float64(overlapCount), (float64(overlapCount) / float64(len(googleResults))) * 100
}

// Function to calculate Spearman's rank correlation coefficient
func spearmanRank(yourResults, googleResults []string) float64 {
	yourRank := rankURLs(yourResults, googleResults)
	googleRank := rankURLs(googleResults, googleResults)

	n := float64(len(googleResults))
	if n == 0 {
		return 0
	}

	var dSquareSum float64
	for i := range yourRank {
		d := yourRank[i] - googleRank[i]
		dSquareSum += d * d
	}

	return 1 - (6*dSquareSum)/(n*(n*n-1))
}

// Function to rank URLs by their positions
func rankURLs(results, googleResults []string) []float64 {
	rankMap := make(map[string]int)
	for i, url := range googleResults {
		rankMap[url] = i + 1 // Rank starts at 1
	}

	ranks := make([]float64, len(results))
	for i, url := range results {
		if rank, exists := rankMap[url]; exists {
			ranks[i] = float64(rank)
		} else {
			ranks[i] = float64(len(googleResults) + 1) // If URL not found, assign a max rank
		}
	}

	return ranks
}

// Function to save results to a CSV
func saveResultsToCSV(queryResults []string, overLapCount []float64, overlaps []float64, correlations []float64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"Queries", "Number of Overlapping Results", "Percent Overlap", "Spearman Correlation"})
	if err != nil {
		return err
	}

	// Write rows
	for i := range queryResults {
		err = writer.Write([]string{
			// queryResults[i],
			fmt.Sprintf("Query %d", i+1),
			fmt.Sprintf("%.2f", overLapCount[i]),
			fmt.Sprintf("%.2f", overlaps[i]),
			fmt.Sprintf("%.2f", correlations[i]),
		})
		if err != nil {
			return err
		}
	}

	// Calculate and write averages
	averageOverlapCount := average(overLapCount)
	averageOverlap := average(overlaps)
	averageCorrelation := average(correlations)
	err = writer.Write([]string{
		"Averages",
		fmt.Sprintf("%.2f", averageOverlapCount),
		fmt.Sprintf("%.2f", averageOverlap),
		fmt.Sprintf("%.2f", averageCorrelation),
	})
	if err != nil {
		return err
	}

	return nil
}

// Function to calculate the average of a slice
func average(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum / float64(len(numbers))
}

// Main function
func compare() {
	// Load your JSON file with scraped results
	yourResultsFile := "hw2.json"              // Replace with actual path
	googleResultsFile := "Google_Result1.json" // Replace with actual path

	yourResults, err := readScrapedResults(yourResultsFile)
	if err != nil {
		log.Fatalf("Error reading your results file: %v", err)
	}
	googleResults, err := readGoogleJSON(googleResultsFile)
	if err != nil {
		log.Fatalf("Error reading Google results file: %v", err)
	}

	// Assuming both files contain the same number of queries
	var overlaps []float64
	var overLapCounts []float64
	var correlations []float64
	var queryNames []string

	for query, googleResultURLs := range googleResults {
		yourResultURLs, exists := yourResults[query]
		if !exists {
			fmt.Println()
			fmt.Printf("Query %s not found in your results\n", query)
			continue
		}

		// Calculate overlap percentage
		overLapCount, overlap := calculateOverlap(yourResultURLs, googleResultURLs)
		// fmt.Sprintf("Query %s | OverLap Count: %d | Overlap Percentage: %.2f", yourResultURLs, overLapCount, overlap)
		overlaps = append(overlaps, overlap)
		overLapCounts = append(overLapCounts, overLapCount)
		// Calculate Spearman correlation
		correlation := spearmanRank(yourResultURLs, googleResultURLs)
		correlations = append(correlations, correlation)

		// Collect query names for CSV
		queryNames = append(queryNames, query)
	}

	// Save the results to CSV
	err = saveResultsToCSV(queryNames, overLapCounts, overlaps, correlations, "results.csv")
	if err != nil {
		log.Fatalf("Error saving CSV file: %v", err)
	}

	fmt.Println("Comparison complete. Results saved to results.csv.")
}

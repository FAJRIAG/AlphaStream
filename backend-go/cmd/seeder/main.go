package main

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/alphastream/backend-go/config"
)

var sectors = []string{
	"Basic Materials",
	"Consumer Cyclicals",
	"Consumer Non-Cyclicals",
	"Energy",
	"Financials",
	"Healthcare",
	"Industrials",
	"Infrastructures",
	"Properties & Real Estate",
	"Technology",
	"Transportation & Logistic",
}

const csvBaseURL = "https://raw.githubusercontent.com/wildangunawan/Dataset-Saham-IDX/master/List%20Emiten/Sectors/"

func main() {
	log.Println("[Seeder] Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Seeder] Failed to load config: %v", err)
	}

	log.Println("[Seeder] Connecting to MySQL...")
	db, err := config.NewDatabaseConnection(cfg.Database)
	if err != nil {
		log.Fatalf("[Seeder] Failed to connect DB: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Clear existing stock records to prevent stale mapping indices
	log.Println("[Seeder] Clearing old stock master list...")
	_, err = db.ExecContext(ctx, "DELETE FROM stocks")
	if err != nil {
		log.Fatalf("[Seeder] Failed to clear stocks table: %v", err)
	}
	_, err = db.ExecContext(ctx, "ALTER TABLE stocks AUTO_INCREMENT = 1")
	if err != nil {
		log.Printf("[Seeder] Warning: could not reset auto-increment: %v", err)
	}

	totalSeeded := 0
	client := &http.Client{}

	for _, sector := range sectors {
		// URL encode the sector filename
		encodedSector := url.PathEscape(sector + ".csv")
		fileURL := csvBaseURL + encodedSector

		log.Printf("[Seeder] Fetching %s list from %s", sector, fileURL)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
		if err != nil {
			log.Printf("[Seeder] Error creating request for %s: %v", sector, err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Seeder] HTTP error for %s: %v", sector, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("[Seeder] Failed to fetch %s (HTTP Status %d)", sector, resp.StatusCode)
			resp.Body.Close()
			continue
		}

		reader := csv.NewReader(resp.Body)
		// Read header
		header, err := reader.Read()
		if err != nil {
			log.Printf("[Seeder] Error reading CSV header for %s: %v", sector, err)
			resp.Body.Close()
			continue
		}

		// Find column indices
		codeIdx, nameIdx := -1, -1
		for i, h := range header {
			if strings.ToLower(h) == "code" {
				codeIdx = i
			} else if strings.ToLower(h) == "name" {
				nameIdx = i
			}
		}

		if codeIdx == -1 || nameIdx == -1 {
			log.Printf("[Seeder] CSV missing code/name columns in %s", sector)
			resp.Body.Close()
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("[Seeder] Failed to begin transaction for %s: %v", sector, err)
			resp.Body.Close()
			continue
		}

		stmt, err := tx.PrepareContext(ctx, "INSERT INTO stocks (symbol, name, exchange, currency, is_active) VALUES (?, ?, 'IDX', 'IDR', 1)")
		if err != nil {
			log.Printf("[Seeder] Failed to prepare statement: %v", err)
			tx.Rollback()
			resp.Body.Close()
			continue
		}

		seededInSector := 0
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("[Seeder] Error reading row in %s: %v", sector, err)
				break
			}

			symbol := strings.TrimSpace(record[codeIdx])
			name := strings.TrimSpace(record[nameIdx])

			if symbol == "" || name == "" {
				continue
			}

			_, err = stmt.ExecContext(ctx, symbol, name)
			if err != nil {
				log.Printf("[Seeder] Error inserting %s: %v", symbol, err)
				continue
			}
			seededInSector++
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			log.Printf("[Seeder] Failed to commit transaction for %s: %v", sector, err)
		} else {
			log.Printf("[Seeder] Seeded %d stocks in %s", seededInSector, sector)
			totalSeeded += seededInSector
		}
		resp.Body.Close()
	}

	log.Printf("[Seeder] Database seeding completed. Total IDX stocks loaded: %d", totalSeeded)
}

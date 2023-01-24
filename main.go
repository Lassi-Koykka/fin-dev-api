package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"github.com/lassi-koykka/fin-dev-api/src/appdb"
	"github.com/lassi-koykka/fin-dev-api/src/models"
)

func main() {
	mux := http.NewServeMux()
	s := gocron.NewScheduler(time.UTC)
	db := appdb.Instance()
	// keywords := db.GetKeywords()
	// keywords := fileutils.ParseFileLines("keywords/technologies.txt")

	db.UpdateData()
	s.Every(1).Hours().Do(func() {  db.UpdateData() })
	s.StartAt(time.Time{})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		fmt.Println(g.TimeStamp(), "---", "GET /", query.Encode())
		results := db.GetPostings(&appdb.SearchTerms{
			Query: query.Get("q"),
			Exact:    query.Has("exact"),
			Company:  query.Get("company"),
			Location: query.Get("location"),
		})
		json.NewEncoder(w).Encode(struct {
			Count      int                 `json:"count"`
			Postings   []models.Posting  `json:"postings"`
			TechCounts models.TechCounts `json:"techCounts"`
		}{
			Count:      len(results),
			Postings:   results,
			TechCounts: models.CountKeywords(results),
		})
	})

	fmt.Println("Listening on port 5050")
	http.ListenAndServe(":5050", mux)
}

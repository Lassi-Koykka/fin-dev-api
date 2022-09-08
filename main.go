package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	appdb "github.com/lassi-koykka/fin-dev-api/src/appdb"
	"github.com/lassi-koykka/fin-dev-api/src/postings"
	"github.com/lassi-koykka/fin-dev-api/src/utils/fileutils"
)

func main() {
	mux := http.NewServeMux()
	s := gocron.NewScheduler(time.UTC)
	db := appdb.Instance()
	keywords := fileutils.ParseFileLines("keywords/technologies.txt")

	db.UpdateData(keywords)
	s.Every(1).Hours().Do(func() {  db.UpdateData(keywords) })
	s.StartAt(time.Time{})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
		}
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		fmt.Println(g.TimeStamp(), "---", "GET /", query.Encode())
		results := db.GetPostings(&appdb.SearchTerms{
			Exact:    query.Has("exact"),
			Company:  query.Get("company"),
			Location: query.Get("location"),
			TechName: query.Get("tech"),
		})
		json.NewEncoder(w).Encode(struct {
			Count      int                 `json:"count"`
			Postings   []postings.Posting  `json:"postings"`
			TechCounts postings.TechCounts `json:"techCounts"`
		}{
			Count:      len(results),
			Postings:   results,
			TechCounts: postings.CountKeywordOccurances(results),
		})
	})

	fmt.Println("Listening on port 5050")
	http.ListenAndServe(":5050", mux)
}

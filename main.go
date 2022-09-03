package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	appdb "github.com/lassi-koykka/fin-dev-api/src/appdb"
	"github.com/lassi-koykka/fin-dev-api/src/postings"
	"github.com/lassi-koykka/fin-dev-api/src/utils/fileutils"
)

func timeStamp() string {
	return time.Now().Format(time.UnixDate)
}

func UpdateData(db appdb.AppDB, keywords []string) {
	fmt.Println("\n------------ UPDATING DATA ------------ ", time.Now().Format(time.UnixDate), "\n ")
	result := postings.FetchAndProcessPostings(keywords)
	db.UpsertPostingsAndPruneDangling(result)
	fmt.Println("\n------------ UPDATING DONE ------------\n ")
}

func main() {
	mux := http.NewServeMux()
	s := gocron.NewScheduler(time.UTC)
	db := appdb.Instance()
	keywords := fileutils.ParseFileLines("keywords/technologies.txt")

	UpdateData(db, keywords)
	s.Every(1).Hours().Do(UpdateData)
	s.StartAsync()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
		}
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		fmt.Println(timeStamp(), "---", "GET /", query.Encode())
		results := db.GetPostings(&appdb.SearchTerms{
			Exact: query.Has("exact"),
			Company: query.Get("company"),
			Location: query.Get("location"),
			TechName: query.Get("tech"),
		})
		json.NewEncoder(w).Encode(struct {
			Postings []postings.Posting
			Counts postings.TechCounts
		}{
			Postings: results,
			Counts: postings.CountKeywordOccurances(results),
		})
	})

	fmt.Println("Listening on port 5050")
	http.ListenAndServe(":5050", mux)
}

package postings

import (
	"fmt"
	"math"
	"os"

	"github.com/lassi-koykka/fin-dev-api/src/datastructures/set"
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"github.com/lassi-koykka/fin-dev-api/src/utils/jsonutils"

	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
	"regexp"
	"strings"
	"sync"
)

type JsonPosting struct {
	Slug              string
	Heading           string
	Date_posted       string
	Municipality_name string
	Export_image_url  string
	Company_name      string
	Descr             string
}

type ApiData struct {
	Count    int
	Next     *string
	Previous *string
	Results  []JsonPosting
}

type Posting struct {
	Slug       string `json:"slug"`
	Heading    string `json:"heading"`
	DatePosted string `json:"datePosted"`
	Url        string `json:"url"`
	ImageUrl   string `json:"imageUrl"`
	Descr      string `json:"descr"`
	Location   string `json:"location"`
	Company    string `json:"company"`
	Keywords   []string `json:"keywords"`
}

// techCountsOverall := countmap.New[int]()
// techCountsByLocation := make(map[string]countmap.CountMap[int])
// techCountsByCompany := make(map[string]countmap.CountMap[int])

const (
	BASE_URL       = "https://duunitori.fi/api/v1/61588b0a2479932129f8ec01c20c16c9179b337d/jobentries?search=koodari&search_also_descr=1&format=json"
	POSTS_PER_PAGE = 100
)

func FetchAndProcessPostings(keywords []string) []Posting {
	_, debug := os.LookupEnv("DEBUG")
	postingsChan := make(chan []Posting, 300)

	var data ApiData
	data = jsonutils.JsonParse[ApiData](g.Fetch(BASE_URL))

	pages := int(math.Ceil(float64(data.Count) / POSTS_PER_PAGE))
	println("Postings:", data.Count, "\tpages:", pages)

	var wg sync.WaitGroup
	wg.Add(pages)

	postingsChan <- parseKeywordsInPostings(data.Results, keywords)

	wg.Done()

	for i := 2; i <= pages; i++ {
		go func(i int) {
			var newData1 ApiData
			var newData2 ApiData
			url := BASE_URL + "&page=" + fmt.Sprintf("%d", i)

			// This is hacky, I know.
			// But how else can you fix a broken api which occasionally returns sligthly different output at random?

			newData1 = jsonutils.JsonParse[ApiData](g.Fetch(url))
			newData2 = jsonutils.JsonParse[ApiData](g.Fetch(url))
			results1 := newData1.Results
			results2 := newData2.Results

			for i := 0; i < len(results1) && i < len(results2); i++ {
				if results1[i].Slug != results2[i].Slug {
					println("MISMATCH", results1[i].Slug)
					results1 = append(results1, results2[i])
				}
			}

			if debug {
				println("GET " + url)
			}
			postings := parseKeywordsInPostings(results1, keywords)
			postingsChan <- postings
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(postingsChan)
	}()

	postingMap := make(map[string]Posting)
	for postings := range postingsChan {
		for _, posting := range postings {
			postingMap[posting.Slug] = posting
		}
	}

	uniquePostings := []Posting{}
	for _, val := range postingMap {
		uniquePostings = append(uniquePostings, val)
	}

	return uniquePostings
}

func matchKeyword(text string, keyword string) bool {
	kw := strings.ToLower(keyword)
	if len(kw) == 1 {
		re, err := regexp.Compile(`\b(` + kw + `\.|` + kw + `\,)\b`)
		g.Check(err)
		return re.Match([]byte(text))
	} else if strings.HasSuffix(strings.ToLower(kw), ".js") || strings.HasSuffix(kw, "JS") {
		re, err := regexp.Compile(`\b(` + kw + "|" + strings.ReplaceAll(strings.ReplaceAll(kw, ".js", ""), "JS", "") + `)\b`)
		g.Check(err)
		return re.Match([]byte(text))
	} else if strings.ContainsAny(kw, "#+.- ") {
		if strings.Contains(text, kw) {
			return true
		}
	} else {
		re, err := regexp.Compile(`\b(` + kw + `)\b`)
		g.Check(err)
		return re.Match([]byte(text))
	}
	return false
}

func tokenizeDescr (str string) []string {
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, "\n", " ")
	replacer := strings.NewReplacer(
		",", " ", 
		". ", " ", 
		"- ", " ", 
		" -", " ", 
		"/", " ", 
		"\\", " ", 
		"(", " ", 
		"[", " ", 
		"{", " ", 
		")", " ", 
		"]", " ", 
		"}", " ", 
		"*", " ", 
		"!", " ", 
		"?", " ", 
		":", " ", 
		"\"", " ", 
		"\t", " ",
		"'", " ")
	result := replacer.Replace(strings.TrimSpace(strings.ToLower(str)))
	return strings.Split(result, " ")
}

func parseKeywordsInPostings(postings []JsonPosting, keywords []string) []Posting {
	resultsChan := make(chan Posting, len(postings))

	var wg sync.WaitGroup
	wg.Add(len(postings))

	for _, p := range postings {
		go func(p JsonPosting) {
			keywordsFound := set.New[string]()
			tokens := tokenizeDescr(p.Heading + " " + p.Descr)
			for _, token :=  range tokens {
				if len(token) < 1 { continue }
				for _, keyword := range keywords {
					if token == strings.ToLower(keyword) {
						keywordsFound.Add(strings.ReplaceAll(keyword, ".js", ""))
					}
				}
			}

			// for _, keyword := range keywords {
			// 	if matchKeyword(text, keyword) {
			// 		keywordsFound.Add(strings.ReplaceAll(keyword, ".js", ""))
			// 	}
			// }

			location, company := "none", "none"
			if len(strings.TrimSpace(p.Municipality_name)) > 0 {
				location = p.Municipality_name
			}
			if len(strings.TrimSpace(p.Company_name)) > 0 {
				company = p.Company_name
			}

			resultsChan <- Posting{
				Slug:       p.Slug,
				Heading:    p.Heading,
				DatePosted: p.Date_posted,
				ImageUrl:   p.Export_image_url,
				Descr:      p.Descr,
				Location:   location,
				Company:    company,
				Keywords:   keywordsFound.Value(),
			}
			defer wg.Done()
		}(p)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := []Posting{}
	for p := range resultsChan {
		results = append(results, p)
	}
	return results
}

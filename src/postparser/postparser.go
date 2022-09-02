package postparser

import (
	"fmt"
	"math"
	"os"
	"github.com/lassi-koykka/fin-dev-api/src/datastructures/countmap"
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
	Slug          string
	Heading       string
	DatePosted    string
	ImageUrl      string
	Descr         string
	Location      string
	Company       string
	KeywordsFound []string
}

// techCountsOverall := countmap.New[int]()
// techCountsByLocation := make(map[string]countmap.CountMap[int])
// techCountsByCompany := make(map[string]countmap.CountMap[int])
type TechCounts struct {
	Overall    countmap.CountMap[int]
	ByLocation map[string]countmap.CountMap[int]
	ByCompany  map[string]countmap.CountMap[int]
}

type ProcessingResult struct {
	Postings   []Posting
	TechCounts TechCounts
}

const (
	BASE_URL       = "https://duunitori.fi/api/v1/jobentries?search=koodari&search_also_descr=1&format=json"
	POSTS_PER_PAGE = 20
)

func FetchAndProcessPosts(keywords []string) ProcessingResult {
	_, debug := os.LookupEnv("DEBUG")
	postingsChan := make(chan []Posting, 300)

	var data ApiData
	data = jsonutils.JsonParse[ApiData](g.Fetch(BASE_URL))

	pages := int(math.Ceil(float64(data.Count) / POSTS_PER_PAGE))
	println("Postings:", data.Count, "\tpages:", pages)

	postings := ParseKeywordsInPostings(data.Results, keywords)
	postingsChan <- postings

	var wg sync.WaitGroup
	wg.Add(pages - 1)

	for i := 2; i <= pages; i++ {
		go func(i int) {
			var newData ApiData
			url := BASE_URL + "&page=" + fmt.Sprintf("%d", i)
			bodyData := g.Fetch(url)
			if debug { println("GET " + url) }
			newData = jsonutils.JsonParse[ApiData](bodyData)
			postings := ParseKeywordsInPostings(newData.Results, keywords)
			postingsChan <- postings
			wg.Done()
		}(i)
	}

	go func() {
		wg.Wait()
		close(postingsChan)
	}()

	allPostings := []Posting{}
	for postings := range postingsChan {
		allPostings = append(allPostings, postings...)
	}
	return ProcessingResult{
		Postings:   allPostings,
		TechCounts: CountKeywordOccurances(allPostings),
	}
}

func CountKeywordOccurances(postings []Posting) TechCounts {
	techCountsOverall := *countmap.New[int]()
	techCountsByLocation := make(map[string]countmap.CountMap[int])
	techCountsByCompany := make(map[string]countmap.CountMap[int])

	for _, r := range postings {
		// Increment overall
		techCountsOverall.IncAll(r.KeywordsFound)
		// Increment company tech counts
		companyMap, ok := techCountsByCompany[r.Company]
		if ok {
			companyMap.IncAll(r.KeywordsFound)
		} else {
			companyTechCounts := countmap.New[int]()
			companyTechCounts.IncAll(r.KeywordsFound)
			techCountsByCompany[r.Company] = *companyTechCounts
		}

		// Increment city tech counts
		cityMap, ok := techCountsByLocation[r.Location]
		if ok {
			cityMap.IncAll(r.KeywordsFound)
		} else {
			cityTechCounts := countmap.New[int]()
			cityTechCounts.IncAll(r.KeywordsFound)
			techCountsByLocation[r.Location] = *cityTechCounts
		}
	}

	return TechCounts{
		Overall:    techCountsOverall,
		ByLocation: techCountsByLocation,
		ByCompany:  techCountsByCompany,
	}
}

func matchKeyword(text string, keyword string) bool {
	kw := strings.ToLower(keyword)
	if len(kw) == 1 {
		matchString := `\b(` + kw + `\.|` + kw + `\,)\b`
		result, err := regexp.MatchString(matchString, text)
		g.Check(err)
		return result
	} else if strings.Contains(kw, ".js") {
		matchString := `\b(` + kw + "|" + strings.ReplaceAll(kw, ".js", "") + `)\b`
		result, err := regexp.MatchString(matchString, text)
		g.Check(err)
		return result
	} else if strings.ContainsAny(kw, "#+.- ") {
		if strings.Contains(text, kw) {
			return true
		}
	} else {
		matchString := `\b(` + kw + `)\b`
		result, err := regexp.MatchString(matchString, text)
		g.Check(err)
		return result
	}
	return false
}

func ParseKeywordsInPostings(postings []JsonPosting, keywords []string) []Posting {
	var wg sync.WaitGroup
	wg.Add(len(postings))
	resultsChan := make(chan Posting, len(postings))

	for _, p := range postings {
		go func(p JsonPosting) {
			keywordsFound := set.New[string]()
			text := strings.ToLower(p.Heading) + " " + strings.ToLower(p.Descr)
			for _, keyword := range keywords {
				if matchKeyword(text, keyword) {
					keywordsFound.Add(strings.ReplaceAll(keyword, ".js", ""))
				}
			}

			location, company := "none", "none"
			if len(strings.TrimSpace(p.Municipality_name)) > 0 {
				location = p.Municipality_name
			}
			if len(strings.TrimSpace(p.Company_name)) > 0 {
				company = p.Company_name
			}

			resultsChan <- Posting{
				Slug:          p.Slug,
				Heading:       p.Heading,
				DatePosted:    p.Date_posted,
				ImageUrl:      p.Export_image_url,
				Descr:         p.Descr,
				Location:      location,
				Company:       company,
				KeywordsFound: keywordsFound.Value(),
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

package models

import (
	"strings"
	"time"

	"github.com/lassi-koykka/fin-dev-api/src/datastructures/countmap"
	"github.com/lassi-koykka/fin-dev-api/src/datastructures/set"
)

type Keyword struct {
	Name      string     `gorm:"primaryKey"`
	Postings  []*Posting `gorm:"many2many:posting_keywords;"`
	Aliases   []Alias
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToKeyword(word string, aliases ...string) Keyword {
	keyword := Keyword{}
	for _, a := range aliases {
		if(len(a) > 0) {
			keyword.Aliases = append(keyword.Aliases, Alias{Name: a})
		}
	}
	return keyword
}

func (keyword Keyword) matchesAlias(token string) bool {
	for _, a := range keyword.Aliases {
		if token == strings.ToLower(a.Name) {
			return true
		}
	}
	return false
}

func parseKeywords(p JsonPosting, keywords []Keyword) []string {
	keywordsFound := set.New[string]()

	simpleKws := []Keyword{}
	complexKws := []Keyword{}
	for _, kw := range keywords {
		if strings.ContainsRune(kw.Name, ' ') {
			complexKws = append(complexKws, kw)
		} else {
			simpleKws = append(simpleKws, kw)
		}
	}

	tokens := tokenizeText(p.Heading + " " + p.Descr)
	fullText := strings.Join(tokens, " ")

	for _, kw := range complexKws {
		if strings.Contains(fullText, strings.ToLower(kw.Name)) {
			keywordsFound.Add(kw.Name)
		}
	}

	for _, token := range tokens {
		if len(token) < 1 || keywordsFound.Includes(token) {
			continue
		}
		for _, kw := range simpleKws {
			if token == strings.ToLower(kw.Name) || kw.matchesAlias(token) {
				keywordsFound.Add(kw.Name)
			}
		}
	}

	return keywordsFound.Value()
}

func CountKeywords(postings []Posting) TechCounts {
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

	counts := TechCounts{}
	counts.ByLocation = map[string][]countmap.Entry{}
	counts.ByCompany = map[string][]countmap.Entry{}

	counts.Overall = techCountsOverall.SortDec()
	for key, val := range techCountsByLocation {
		counts.ByLocation[key] = val.SortDec()
	}

	for key, val := range techCountsByCompany {
		counts.ByCompany[key] = val.SortDec()
	}

	return counts
}

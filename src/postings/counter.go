package postings

import (
	"github.com/lassi-koykka/fin-dev-api/src/datastructures/countmap"
)


type TechCounts struct {
	Overall    countmap.CountMap[int] `json:"overall"`
	ByLocation map[string]countmap.CountMap[int] `json:"byLocation"`
	ByCompany  map[string]countmap.CountMap[int] `json:"byCompany"`
}

func CountKeywordOccurances(postings []Posting) TechCounts {
	techCountsOverall := *countmap.New[int]()
	techCountsByLocation := make(map[string]countmap.CountMap[int])
	techCountsByCompany := make(map[string]countmap.CountMap[int])

	for _, r := range postings {
		// Increment overall
		techCountsOverall.IncAll(r.Keywords)
		// Increment company tech counts
		companyMap, ok := techCountsByCompany[r.Company]
		if ok {
			companyMap.IncAll(r.Keywords)
		} else {
			companyTechCounts := countmap.New[int]()
			companyTechCounts.IncAll(r.Keywords)
			techCountsByCompany[r.Company] = *companyTechCounts
		}

		// Increment city tech counts
		cityMap, ok := techCountsByLocation[r.Location]
		if ok {
			cityMap.IncAll(r.Keywords)
		} else {
			cityTechCounts := countmap.New[int]()
			cityTechCounts.IncAll(r.Keywords)
			techCountsByLocation[r.Location] = *cityTechCounts
		}
	}

	return TechCounts{
		Overall:    techCountsOverall,
		ByLocation: techCountsByLocation,
		ByCompany:  techCountsByCompany,
	}
}

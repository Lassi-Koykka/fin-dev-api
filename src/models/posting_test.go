package models

import (
	"fmt"
	"os"
	"testing"

	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	json "github.com/lassi-koykka/fin-dev-api/src/utils/jsonutils"
)

// func TestFetchAndProcessPostings(t *testing.T) {
// 	keywords := fileutils.ParseFileLines("../../keywords/technologies.txt")
// 	result := FetchAndProcessPostings(keywords)
// 	if len(result) < 1 {
// 		t.Error("No results returned")
// 	}
// }

func TestTokenizePosting(t *testing.T) {
	jsonData, err := os.ReadFile("../../testdata/posts.json")
	g.Check(err)
	data := json.JsonParse[ApiData](jsonData)
	postings := data.Results
	for i, p := range postings {
		content := p.Descr + " " + p.Heading
		tokens := tokenizeText(content)
		fmt.Println(i, len(tokens))
		if i == 4 {
			fmt.Println(tokens)
		}
	}
}

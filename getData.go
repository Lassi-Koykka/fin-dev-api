package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "strings"
    "regexp"
    "net/http"
    "encoding/json"
    "github.com/lassi-koykka/fin-dev-api/set"
)


const BASE_URL = "https://duunitori.fi/api/v1/jobentries?search=koodari&search_also_descr=1&format=json"

type Posting struct {
    Slug string
    Heading string
    Date_posted string
    Municipality_name *string
    Export_image_url *string
    Company_name string
    Descr string
}

type ApiData struct {
    Count int
    Next *string
    Previous *string
    Results []Posting
}

func check (err error) {
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
}
func readKeywords() []string {
    data, err := os.ReadFile("keywords/technologies.txt")
    check(err)
    content := string(data)
    keywords := strings.Split(content, "\n")
    result := []string{}
    for _, kw := range keywords {
        result = append(result, strings.TrimSpace(kw))
    }
    return keywords
}

func readPostsFromFile() ApiData {
    data, err := os.ReadFile("posts.json")
    check(err)
    return ParseApiData(data)
}

func GetApiData(url string) []byte {
    res, err := http.Get(url)
    check(err)
    defer res.Body.Close()
    bodyData, parseBodyErr := ioutil.ReadAll(res.Body)
    check(parseBodyErr)
    return bodyData
}

func ParseApiData(bodyData []byte) ApiData {
    var data ApiData
    err := json.Unmarshal(bodyData, &data)
    check(err)
    return data
}

func getAllPosts(saveToFile bool) ApiData {
    bodyData := GetApiData(BASE_URL)

    var data ApiData
    data = ParseApiData(bodyData)

    next := data.Next


    i := 1
    for next != nil {
        println("GET " + *next)
        fmt.Printf("Try %d \n", i)
        var newData ApiData
        newBodyData := GetApiData(*next)
        newData = ParseApiData(newBodyData)
        data.Results = append(data.Results, newData.Results...)
        next = newData.Next
        i++
    }

    if saveToFile {
        res, err := json.Marshal(data)
        check(err)
        writeErr := os.WriteFile("posts.json", res, 0666)
        check(writeErr)
    }

    return data
}

// Go through each posting and
// 1. regex all match all words in keywords list
func handleData(postings []Posting, keywords []string) {
    keywordsFound := set.NewStrSet()
    technologiesByPosting := make(map[string][]string)
    //TODO FINISH COUNTING
    // technologyCountsOverall := make(map[string]int)
    // technologyCountsByCity := make(map[string]map[string]int)
    for _, p := range postings {
        descr := strings.ToLower(p.Descr)
        for _, keyword := range keywords {
            kw := strings.ToLower(keyword)
            resultKw := strings.ReplaceAll(keyword, ".js", "")
            found := false
            if len(kw) == 1 {
                matchString := `\b(` + kw + `\.|` + kw + `\,)\b`
                result, err := regexp.MatchString(matchString, descr)
                check(err)
                found = result
            } else if strings.Contains(kw, ".js") {
                matchString := `\b(` + kw + "|" + strings.ReplaceAll(kw, ".js", "") + `)\b`
                result, err := regexp.MatchString(matchString, descr)
                check(err)
                found = result
            } else if strings.ContainsAny(kw, "#+.- ") {
                if strings.Contains(descr, kw) { found = true }
            } else {
                matchString := `\b(` + kw + `)\b`
                result, err := regexp.MatchString(matchString, descr)
                check(err)
                found = result 
            }

            if found { 
                keywordsFound.Add(resultKw) 
            }

        }

        // city := strings.ToLower(*p.Municipality_name)
        fmt.Println(keywordsFound.ToSlice())
        technologiesByPosting[p.Slug] = keywordsFound.ToSlice()
    }
}

func main() {
    println("Running main")
    keywords := readKeywords()

    // apiData := getAllPosts(true)
    apiData := readPostsFromFile()
    handleData(apiData.Results, keywords)
    fmt.Println(len(apiData.Results))
}

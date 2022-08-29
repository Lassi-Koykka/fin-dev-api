package main

import (
    "io/ioutil"
    "fmt"
    "os"
    "strings"
    "net/http"
    "encoding/json"
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


func main() {
    println("Running main")
    // keywords := readKeywords()
    // fmt.Println(keywords)

    // apiData := getAllPosts(true)
    apiData := readPostsFromFile()
    fmt.Println(len(apiData.Results))
}

package main

import (
	"fmt"
	appdb "github.com/lassi-koykka/fin-dev-api/src/db"
	"github.com/lassi-koykka/fin-dev-api/src/postparser"
	// g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"github.com/lassi-koykka/fin-dev-api/src/utils/fileutils"
)


func main() {

	// DB
	db := appdb.ConnectAndMigrate()

	fmt.Println("Migrated")

	keywords := fileutils.ParseFileLines("keywords/technologies.txt")
	result := postparser.FetchAndProcessPosts(keywords)

	appdb.UpsertPostingsAndPruneDangling(db, result)

}

package appdb

import (
	"fmt"

	models "github.com/lassi-koykka/fin-dev-api/src/models"
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DB_NAME = "database.db"
)

type AppDB struct {
	Db *gorm.DB
}

type SearchTerms struct {
	Query string
	Location string
	Company  string
	Exact    bool
}

func Instance() AppDB {
	db, err := gorm.Open(sqlite.Open(DB_NAME), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	g.Check(err)

	fmt.Println("Running auto migrations")
	db.AutoMigrate(&models.Posting{})
	db.AutoMigrate(&models.Keyword{})
	db.AutoMigrate(&models.Alias{})
	fmt.Println("Migration has completed")

	return AppDB{
		Db: db,
	}
}

func searchInput(str string, exact bool) string {
	if !exact {
		return "%" + str + "%"
	}
	return str
}

func (db *AppDB) UpdateData() {
	keywords := db.GetKeywords();
	fmt.Println("\nFOUND", len(keywords),"KEYWORDS", g.TimeStamp(), "\n ")
	fmt.Println("\n------------ UPDATING DATA ------------ ", g.TimeStamp(), "\n ")
	result := models.FetchAndProcessPostings(keywords)
	db.UpsertAndPrunePostings(result)
	fmt.Println("\n------------ UPDATING DONE ------------\n ")
}

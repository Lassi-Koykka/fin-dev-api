package appdb

import (
	models "github.com/lassi-koykka/fin-dev-api/src/models"
	"gorm.io/gorm/clause"
)

func (appdb *AppDB) GetKeywords () []models.Keyword {
	db := appdb.Db
	var queryResult []models.Keyword
	db.Find(&queryResult)
	return queryResult;
}
// UPSERT
func (appdb *AppDB) UpsertKeyword (word string, aliases ...string) models.Keyword  {
	if(len(word) < 0) {
		return models.Keyword{}
	} 
	db := appdb.Db
	keyword := models.ToKeyword(word, aliases...)
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&keyword)
	db.Save(&keyword)
	return keyword
}

// UPSERT
func (appdb *AppDB) UpsertKeywords (input [][]string) []models.Keyword {
	keywords := []models.Keyword{}
	if len(input) < 1 {
		return keywords
	}
	db := appdb.Db
	var count int64
	for _, kw := range input {
		if len(kw) < 1 { continue }
		keywords = append(keywords, models.ToKeyword(kw[0], kw...))
	}
	db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&keywords).Count(&count)
	db.Save(&keywords)
	return keywords
}

// DELETE
func (appdb *AppDB) DeleteKeyword (kw string) models.Keyword {
	db := appdb.Db
	keyword := models.Keyword{Name: kw}
	db.Delete(&keyword)
	return keyword;
}

package rtimage

import (
	"github.com/HouzuoGuo/tiedot/db"
	// "github.com/HouzuoGuo/tiedot/db"
	//	"github.com/HouzuoGuo/tiedot/dberr"
)

type SearchDoc struct {
	id       string
	sha256   string
	filename string
}

type SearchEngine interface {
	getByID(id string) SearchDoc
	save(doc SearchDoc)
}

var database *db.DB

/*
func OpenDB(appSettings Settings) *db.DB {
	if database == nil {
		_db, err := db.OpenDB(fmt.Sprintf("%s/_db", appSettings.WorkDir))
		if err != nil {
			panic(err)
		}
		database = _db

		// Create collection Pictures
		if err := database.Create("Pictures"); err != nil {
			log.Printf("collection %s esistente", "Pictures")
		}

		picturesColl := database.Use("Pictures")
		if err := picturesColl.Index([]string{"sha256"}); err != nil {
			log.Printf("errore creando indici %s", err.Error())
		}

	}
	return database
}

func CloseDB() {
	database.Close()
}
*/

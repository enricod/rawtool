package rtimage

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/enricod/golibraw"
	"github.com/nfnt/resize"
	"github.com/xiam/exif"
	// "github.com/HouzuoGuo/tiedot/db"
	//	"github.com/HouzuoGuo/tiedot/dberr"
)

// RawExtensions elenco delle estensioni dei file raw
func RawExtensions() []string {
	return []string{".ORF", ".CR2", ".RAF", ".ARW"}
}

type SearchDoc struct {
	id       string
	sha256   string
	filename string
}
type SearchEngine interface {
	getByID(id string) SearchDoc
	save(doc SearchDoc)
}

// Settings contiene impostazioni del programma
type Settings struct {
	ImagesDir string // directory with images
	WorkDir   string // directory where are saved the thumbnails
	Recursive bool
}

// MyImage struct used to save informations about the image which is processed
type MyImage struct {
	Thumb    string // thumbnail path
	Filename string // original filename
	Path     string // percorso completo immagine originale
}

// JobQueue make a channel with a capacity of 100.
type JobQueue struct {
	Settings                Settings
	numeroImmaginiElaborate int
	JobChan                 chan MyImage
}

var database *db.DB

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
func Worker(queue *JobQueue) {
	for img := range queue.JobChan {
		ProcessMyimage(img, queue.Settings)
		// time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)
	}
}

func calcolaOutputDir(imgMeta golibraw.ImgMetadata) string {
	year := string(imgMeta.ScattoDataOra[0:4])
	day := string(imgMeta.ScattoDataOra[0:10])
	return year + "/" + day
}

func calcolaOutputDirFromExifTags(exifData *exif.Data) string {
	year := string(exifData.Tags["Date and Time"][0:4])
	day := strings.ReplaceAll(string(exifData.Tags["Date and Time"][0:10]), ":", "-")
	return year + "/" + day
}

// NewQueue coda di elaborazione
func NewQueue(appSettings Settings) *JobQueue {
	q := JobQueue{
		Settings:                appSettings,
		JobChan:                 make(chan MyImage),
		numeroImmaginiElaborate: 0,
	}

	return &q
}

func (q *JobQueue) EnqueueImage(img MyImage) {
	q.JobChan <- img
	q.numeroImmaginiElaborate++
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func calcolaSha256(file string) string {
	//t0 := time.Now()
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	// log.Printf("sha256sum %s required %d", file, time.Since(t0))
	return fmt.Sprintf("%x", h.Sum(nil))
}

/*
func queryImageDocument(picturesColl *db.Col, sha256 string) map[string]interface{} {
	var query interface{}
	json.Unmarshal([]byte(`[{"eq": "`+sha256+`", "in": ["sha256"]}]`), &query)
	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys

	if err := db.EvalQuery(query, picturesColl, &queryResult); err != nil {
		log.Printf("errore nella query %s", err.Error())
	}

	if len(queryResult) > 0 {
		id := queryResult[0]
		readBack, err := picturesColl.Read(id)
		if err != nil {
			log.Printf("errore nella query %s\n", err.Error())
			return nil
		}
		log.Printf("Query returned document %v\n", readBack)
		return readBack

	}
	return nil
}
*/
// ProcessMyimage elabora immagine
func ProcessMyimage(myimg MyImage, settings Settings) (MyImage, error) {

	//picturesColl := database.Use("Pictures")

	searchClient := SolrClient{host: "http://localhost:8983/solr"}

	sha256 := calcolaSha256(myimg.Path)
	searchDoc, err := searchClient.getByID(sha256)
	// indexDoc := queryImageDocument(picturesColl, sha256)
	// cerchiamo se file è già stato elaborato

	if err == nil {
		log.Printf("documento già elaborato: %s", myimg.Path)
		return MyImage{}, nil
	}

	searchDoc = SearchDoc{sha256: sha256, id: sha256, filename: myimg.Path}
	searchClient.save(searchDoc)

	/*
		docID, err := picturesColl.Insert(map[string]interface{}{
			"sha256":           sha256,
			"originalFilename": myimg.Path,
			"stars":            0,
			"tags":             ""})
		if err != nil {
			panic(err)
		}
		log.Printf("documento inserito nel database: %v", docID)
	*/
	ext := filepath.Ext(myimg.Filename)

	if strings.ToUpper(ext) == ".JPG" {
		// devo solo ridimensionare l'immagine
		exifData, err := exif.Read(myimg.Path)
		if err != nil {
			fmt.Println(myimg.Path + " error reading exif data")
			return MyImage{}, err
		}
		//for key, val := range data.Tags {
		//	fmt.Printf("%s = %s\n", key, val)
		//}
		outDir := fmt.Sprintf("%s/%s", settings.WorkDir, calcolaOutputDirFromExifTags(exifData))
		imgfile, err := os.Open(myimg.Path)
		if err != nil {
			fmt.Println(myimg.Path + " file not found!")
		} else {
			defer imgfile.Close()
			img, _, err := image.Decode(imgfile)
			if err != nil {
				log.Printf("errore %v \n", err)
			} else {
				log.Printf("loading  %s, saving in %s", myimg.Path, outDir)
				writeThumb(outDir, myimg.Filename, &img, settings)
			}
		}

	} else if IsStringInSlice(ext, RawExtensions()) {
		img, imgMetadata, err2 := golibraw.Raw2Image(myimg.Path)
		if err2 == nil {
			//log.Printf("img metadata %v", imgMetadata)
			outDir := fmt.Sprintf("%s/%s", settings.WorkDir, calcolaOutputDir(imgMetadata))
			outfilename := fmt.Sprintf("%s/%s.thumb.jpg", outDir, myimg.Filename)
			if FileExists(outfilename) {
				return myimg, nil
			}
			log.Printf("loading  %s, saving in %s", myimg.Path, outDir)
			writeThumb(outDir, myimg.Filename, &img, settings)
		} else {
			log.Printf("error decoding %s %s\n", myimg.Filename, err2.Error())
		}
	}
	return myimg, nil
}

// IsStringInSlice true if the slice contains the string a
func IsStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a || b == strings.ToUpper(a) {
			return true
		}
	}
	return false
}

func writeThumb(outDir string, filename string, img *image.Image, settings Settings) error {
	t0 := time.Now()
	var opt jpeg.Options
	opt.Quality = 75
	// ok, write out the data into the new JPEG file
	// perchè serve questo??
	rand.Seed(time.Now().UTC().UnixNano())
	outfilename := fmt.Sprintf("%s/%s.thumb.jpg", outDir, filename)
	os.MkdirAll(outDir, os.ModePerm)

	out, err := os.Create(outfilename)
	if err != nil {
		log.Printf("ERROR %s", err.Error())
		return err
	}
	newImage := resize.Resize(1280, 0, *img, resize.Lanczos3)
	err = jpeg.Encode(out, newImage, &opt)
	if err != nil {
		log.Printf("ERROR %s", err.Error())
		return err
	}
	log.Printf("created and saved thumbnail %s , required %v", outfilename, time.Since(t0))
	return nil
}

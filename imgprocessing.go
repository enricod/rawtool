package main

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

	"github.com/enricod/golibraw"
	"github.com/nfnt/resize"
	"github.com/xiam/exif"
)

// RawExtensions elenco delle estensioni dei file raw
func RawExtensions() []string {
	return []string{".ORF", ".CR2", ".RAF", ".ARW"}
}

// MyImage struct used to save informations about the image which is processed
type MyImage struct {
	ThumbPath string // thumbnail path
	Filename  string // original filename
	Path      string // percorso completo immagine originale
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

// ProcessMyimage elabora immagine
func ProcessMyimage(myimg MyImage, settings Settings) (MyImage, error) {

	//picturesColl := database.Use("Pictures")
	//useSearchEngine := false

	sha256 := calcolaSha256(myimg.Path)
	fmt.Printf("image: %s, sha256 %s\n", myimg.Path, sha256)
	/*
		if useSearchEngine {
			searchClient := SolrClient{host: "http://localhost:8983/solr"}
			searchDoc, err := searchClient.getByID(sha256)
			// indexDoc := queryImageDocument(picturesColl, sha256)
			// cerchiamo se file è già stato elaborato

			if err == nil {
				log.Printf("documento già elaborato: %s", myimg.Path)
				return MyImage{}, nil
			}

			searchDoc = SearchDoc{sha256: sha256, id: sha256, filename: myimg.Path}
			searchClient.save(searchDoc)
		}
	*/
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

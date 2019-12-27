package rtimage

import (
	"fmt"
	"image"
	"image/jpeg"
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

// make a channel with a capacity of 100.
type JobQueue struct {
	Settings                Settings
	numeroImmaginiElaborate int
	JobChan                 chan MyImage
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

func RawExtensions() []string {
	return []string{".ORF", ".CR2", ".RAF", ".ARW"}
}

func ProcessMyimage(myimg MyImage, settings Settings) (MyImage, error) {

	ext := filepath.Ext(myimg.Filename)

	if strings.ToUpper(ext) == ".JPG" {
		// I just create the thumbnail

		exifData, err := exif.Read(myimg.Path)
		if err != nil {
			fmt.Println(myimg.Path + " error reading exif data")
			return MyImage{}, err
		}
		//log.Printf("data scatto = %s", exifData.Tags["Date and Time"])
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

/*
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}*/

func writeThumb(outDir string, filename string, img *image.Image, settings Settings) error {
	t0 := time.Now()
	var opt jpeg.Options

	opt.Quality = 70
	// ok, write out the data into the new JPEG file

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

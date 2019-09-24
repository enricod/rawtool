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
)

type Settings struct {
	ImagesDir string // directory with images
	WorkDir   string // directory .rawtool
}

type MyImage struct {
	Thumb    string
	Filename string
	Path     string
}

// make a channel with a capacity of 100.
type JobQueue struct {
	Settings                Settings
	numeroImmaginiElaborate int
	jobChan                 chan MyImage
}

func Worker(queue *JobQueue) {
	for img := range queue.jobChan {
		ProcessMyimage(img, queue.Settings)
		// time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)
	}
}

// NewQueue coda di elaborazione
func NewQueue(appSettings Settings) *JobQueue {
	q := JobQueue{
		Settings:                appSettings,
		jobChan:                 make(chan MyImage),
		numeroImmaginiElaborate: 0,
	}

	return &q
}

func (q *JobQueue) EnqueueImage(img MyImage) {
	q.jobChan <- img
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
	outpath := fmt.Sprintf("%s/.rawtool", myimg.Path)
	outfilename := fmt.Sprint(outpath, "/", myimg.Filename, ".thumb.jpg")
	log.Printf("generazione thumb %s", outfilename)
	ext := filepath.Ext(myimg.Filename)
	if FileExists(outfilename) {
		log.Printf("thumb already created %s", outfilename)
	} else {
		if strings.ToUpper(ext) == ".JPG" {
			// I just create the thumbnail
			infile := myimg.Path + "/" + myimg.Filename
			imgfile, err := os.Open(infile)

			if err != nil {
				fmt.Println(infile + " file not found!")
			} else {
				defer imgfile.Close()
				img, _, err := image.Decode(imgfile)
				if err != nil {
					log.Printf("errore %v \n", err)
				} else {
					writeThumb(outpath, myimg.Filename, &img, settings)
				}
			}
		} else if IsStringInSlice(ext, RawExtensions()) {

			fileAbsPath := fmt.Sprintf("%s/%s", myimg.Path, myimg.Filename)
			log.Printf("loading  %s", fileAbsPath)
			fileInfo, err3 := os.Stat(fileAbsPath)
			if err3 != nil {
				log.Printf("error decoding %s  %s\n", fileAbsPath, err3.Error())
			}
			img, err2 := golibraw.Raw2Image(settings.ImagesDir, fileInfo)
			if err2 == nil {
				//result = (result, myImage{Image: img, Filename: f})
				writeThumb(outpath, myimg.Filename, &img, settings)
			} else {
				log.Printf("error decoding %s %s\n", myimg.Filename, err2.Error())
			}
		}
	}
	myimg.Thumb = outfilename
	return myimg, nil
}

// IsStringInSlice true if the slice contains the string a
func IsStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == strings.ToUpper(a) {
			return true
		}
	}
	return false
}

func writeThumb(path string, filename string, img *image.Image, settings Settings) error {
	t0 := time.Now()
	var opt jpeg.Options

	opt.Quality = 70
	// ok, write out the data into the new JPEG file

	rand.Seed(time.Now().UTC().UnixNano())
	outfilename := fmt.Sprint(settings.WorkDir, "/", filename, ".thumb.jpg")

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

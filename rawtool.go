package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enricod/rawtool/rtimage"
)

const thumbSize = 1280

//var imageLabel *widgets.QLabel

var imageIndex int
var images []rtimage.MyImage

// var imagesNumLabel *widgets.QLabel

func rawExtensions() []string {
	return []string{".ORF", ".CR2", ".RAF", ".ARW"}
}

var appSettings rtimage.Settings

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func createWorkDirIfNecessary(_appSettings rtimage.Settings) {
	outdir := fmt.Sprint(_appSettings.ImagesDir, "/", ".rawtool")
	createDirIfNotExist(outdir)
}

func processDir(dirname string, appSettings rtimage.Settings) {
	imagesInWorkDir, _ := readImagesInDir(dirname)
	images = imagesInWorkDir
	useQueues := false
	if useQueues {
		q := rtimage.NewQueue(appSettings)

		go rtimage.Worker(q)

		for _, f := range imagesInWorkDir {
			go q.EnqueueImage(f)
		}
		time.Sleep(5 * time.Minute)
	} else {
		log.Printf("trovate %d immagini ", len(images))
		for index, img := range images {
			log.Printf("%d/%d elaborazione di %s", index, len(images), img.Filename)

			rtimage.ProcessMyimage(img, appSettings)
		}
	}
}

func intMin(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
func processNextImages(images []rtimage.MyImage, start int, howmany int) {
	for i := start + 1; i <= intMin(len(images)-1, start+1+howmany); i++ {
		rtimage.ProcessMyimage(images[i], appSettings)
	}
}

func main() {
	//defer profile.Start(profile.MemProfile).Stop()

	imageIndex = 0

	imagesdir := flag.String("d", ".", "imagesdir")
	flag.Parse()

	dir, err := filepath.Abs(filepath.Dir(*imagesdir))
	if err != nil {
		log.Fatal(err)
	}

	outdir := "/data1/thumbs" // fmt.Sprint(dir, "/", ".rawtool")
	appSettings = rtimage.Settings{ImagesDir: *imagesdir, WorkDir: outdir}
	// createWorkDirIfNecessary(appSettings)

	processDir(dir, appSettings)
	// FIXME per ora facciamo solo da CLI
	// createUI()
}

func readImagesInDir(dirname string) ([]rtimage.MyImage, error) {

	result := []rtimage.MyImage{}
	skipDirs := []string{".", "..", ".dtrash", ".Trash-1000", ".rawtool"}

	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && rtimage.IsStringInSlice(info.Name(), skipDirs) {
			//fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		if info.IsDir() {
			log.Printf("walking into dir %s %s", path, info.Name())
		}
		//files, _ := ioutil.ReadDir(path)

		//for _, f := range files {
		ext := filepath.Ext(path)
		if strings.ToUpper(ext) == ".JPG" || rtimage.IsStringInSlice(ext, rtimage.RawExtensions()) {
			//log.Printf("trovata immagine %s | %s", path, info.Name())
			result = append(result, rtimage.MyImage{Path: path, Filename: info.Name()})
		}
		//}

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dirname, err)
		return result, nil
	}

	return result, nil
}

func isImage(filename string) bool {
	ext := filepath.Ext(filename)
	if strings.ToUpper(ext) == ".JPG" {
		return true
	} else if rtimage.IsStringInSlice(ext, rtimage.RawExtensions()) {
		return true
	}
	return false
}

func writeAsJpeg(filename os.FileInfo, img image.Image) error {
	var opt jpeg.Options

	opt.Quality = 80
	// ok, write out the data into the new JPEG file

	rand.Seed(time.Now().UTC().UnixNano())

	out, err := os.Create("./.rawtool/out.jpg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = jpeg.Encode(out, img, &opt) // put quality to 80%
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

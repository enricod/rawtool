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
)

const thumbSize = 1280

var images []MyImage

var appSettings Settings

// Settings contiene impostazioni del programma
type Settings struct {
	ImagesDir string // directory with images
	WorkDir   string // directory where are saved the thumbnails
	Recursive bool
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func createWorkDirIfNecessary(_appSettings Settings) {
	outdir := fmt.Sprint(_appSettings.ImagesDir, "/", ".rawtool")
	createDirIfNotExist(outdir)
}

// ProcessDir elabora immagini in directory
func ProcessDir(dirname string, appSettings Settings) {
	imagesInWorkDir, _ := readImagesInDir(dirname)
	images = imagesInWorkDir

	log.Printf("found %d images ", len(images))
	for index, img := range images {
		log.Printf("step %d/%d, processing image %s", index+1, len(images), img.Filename)

		workReq := WorkRequest{SourceImage: img, Delay: 10}

		// Push the work onto the queue.
		WorkQueueWorkRequestChan <- workReq
	}

}

func intMin(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func processNextImages(images []MyImage, start int, howmany int) {
	for i := start + 1; i <= intMin(len(images)-1, start+1+howmany); i++ {
		ProcessMyimage(images[i], appSettings)
	}
}

func readImagesInDir(dirname string) ([]MyImage, error) {

	result := []MyImage{}
	skipDirs := []string{".", "..", ".dtrash", ".Trash-1000", ".rawtool"}

	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && IsStringInSlice(info.Name(), skipDirs) {
			//fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}

		ext := filepath.Ext(path)
		if strings.ToUpper(ext) == ".JPG" || IsStringInSlice(ext, RawExtensions()) {

			result = append(result, MyImage{Path: path, Filename: info.Name()})
		}

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
	} else if IsStringInSlice(ext, RawExtensions()) {
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

func main() {
	//defer profile.Start(profile.MemProfile).Stop()
	imagesdir := flag.String("d", ".", "imagesdir")
	flag.Parse()

	//outdir := flag.String("o", "/data1/thumbs", "outputdir")
	outdir := flag.String("o", "/home/enrico/tmp/thumbs", "outputdir")

	dir, err := filepath.Abs(filepath.Dir(*imagesdir))
	if err != nil {
		log.Fatal(err)
	}

	appSettings = Settings{ImagesDir: *imagesdir, WorkDir: *outdir}

	StartDispatcher(4)

	ProcessDir(dir, appSettings)

	time.Sleep(100000 * time.Millisecond)
}

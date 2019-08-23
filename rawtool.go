package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enricod/golibraw"
	"github.com/nfnt/resize"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

//	When we run `go generate` from the cli it will run the
//	`go run` command outlined below
//	**Important: sure to include the comment below for the generator to see**
//go:generate go run generators/generator.go

const thumbSize = 1280

var imageLabel *widgets.QLabel

var imageIndex int
var images []myImage
var imagesNumLabel *widgets.QLabel

func rawExtensions() []string {
	return []string{".ORF", ".CR2", ".RAF", ".ARW"}
}

type Settings struct {
	ImagesDir string // directory with images
	WorkDir   string // directory .rawtool
}

type myImage struct {
	Thumb    string
	Filename string
	Path     string
}

var appSettings Settings

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func createUI() {
	// needs to be called once before you can start using the QWidgets
	app := widgets.NewQApplication(len(os.Args), os.Args)

	// create main window
	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(1024+400, 1024)
	window.SetWindowTitle("RawTool")

	mainWidget := widgets.NewQWidget(nil, 0)
	mainWidget.SetLayout(widgets.NewQHBoxLayout())
	window.SetCentralWidget(mainWidget)

	leftWidget := widgets.NewQWidget(mainWidget, 0)
	leftWidget.SetLayout(widgets.NewQVBoxLayout())
	mainWidget.Layout().AddWidget(leftWidget)

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	input := widgets.NewQLineEdit(nil)
	input.SetPlaceholderText("Write something ...")
	leftWidget.Layout().AddWidget(input)

	// create a button
	// connect the clicked signal
	// and add it to the central widgets layout
	button := widgets.NewQPushButton2("Select dir", nil)
	button.ConnectClicked(func(bool) {
		//dialog := widgets.NewQFileDialog(window, core.Qt__Dialog)
		// dialog.OpenDefault()
		// dialog.ConnectFileSelected(dirSelected)

		var selecteddir string
		selecteddir = widgets.QFileDialog_GetExistingDirectory(window, "select dir", appSettings.ImagesDir, 1)

		if !strings.HasSuffix(selecteddir, "/") {
			selecteddir = selecteddir + "/"
		}

		//go processImagesInDir(selecteddir)

		imagesInWorkDir, _ := readImagesInDir(selecteddir)
		images = imagesInWorkDir
		imageIndex = 0

		showImage(images, imageIndex)
		//widgets.QMessageBox_Information(nil, "OK", input.Text(), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	leftWidget.Layout().AddWidget(button)

	rightWidget := widgets.NewQWidget(mainWidget, 0)
	rightWidget.SetLayout(widgets.NewQVBoxLayout())
	mainWidget.Layout().AddWidget(rightWidget)

	imageLabel = widgets.NewQLabel(nil, 0)

	// esistono sicuramente metodi migliori che non fare embde dell'immagine vuota, scriverla su disco e rileggerla!!!
	// ma per ora lasciamo così
	err := ioutil.WriteFile("/tmp/empty.png", empty1024, 0644)
	if err != nil {
		panic(err.Error())
	}

	testImageFileName := "/tmp/empty.png"
	image := gui.NewQImage9(testImageFileName, "")

	pixmap := gui.QPixmap_FromImage(image, core.Qt__AutoColor)

	imageLabel.SetPixmap(pixmap)
	imageLabel.Resize(pixmap.Size())

	rightWidget.Layout().AddWidget(imageLabel)

	// barra di navigazione tra le immagini
	imagesNavBarWidget := widgets.NewQWidget(mainWidget, 0)
	imagesNavBarWidget.SetLayout(widgets.NewQHBoxLayout())

	imagesNumLabel = widgets.NewQLabel(nil, 0)
	imagesNumLabel.SetText("")
	imagesNavBarWidget.Layout().AddWidget(imagesNumLabel)

	prevImageBtn := widgets.NewQPushButton2("<", nil)
	imagesNavBarWidget.Layout().AddWidget(prevImageBtn)
	nextImageBtn := widgets.NewQPushButton2(">", nil)
	nextImageBtn.ConnectClicked(func(bool) {
		imageIndex = imageIndex + 1
		if imageIndex >= len(images) {
			imageIndex = 0
		}
		showImage(images, imageIndex)
	})
	imagesNavBarWidget.Layout().AddWidget(nextImageBtn)

	rightWidget.Layout().AddWidget(imagesNavBarWidget)

	window.Show()
	app.Exec()
}

func showImage(images []myImage, index int) {
	if index >= len(images) {
		return
	}
	img, err := processMyimage(images[index])
	if err != nil {
		log.Printf("ERROR %s", err.Error())
	} else {
		log.Printf("%s", img.Thumb)

		image := gui.NewQImage9(img.Thumb, "")
		pixmap := gui.QPixmap_FromImage(image, core.Qt__AutoColor)
		imageLabel.SetPixmap(pixmap)
		imagesNumLabel.SetText(fmt.Sprintf("%d / %d images", imageIndex+1, len(images)))
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
	fmt.Println(dir)

	outdir := fmt.Sprint(dir, "/", ".rawtool")
	createDirIfNotExist(outdir)

	appSettings = Settings{ImagesDir: *imagesdir, WorkDir: outdir}

	createUI()
}

// called when user selects a directory
func dirSelected(dirname string) {
	log.Printf("selected dir %s \n", dirname)
	myImages, err := readImagesInDir(dirname)
	if err != nil {
		log.Printf("error %s", err.Error())
	} else {
		for _, myimg := range myImages {
			processMyimage(myimg)
		}
	}
	//go processImagesInDir(dirname)
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

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func readImagesInDir(dirname string) ([]myImage, error) {
	files, _ := ioutil.ReadDir(dirname)
	result := []myImage{}
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if strings.ToUpper(ext) == ".JPG" || IsStringInSlice(ext, rawExtensions()) {
			result = append(result, myImage{Path: dirname, Filename: f.Name()})
		}
	}
	return result, nil
}

func processMyimage(myimg myImage) (myImage, error) {
	outpath := fmt.Sprint(myimg.Path, "/.rawtool")
	outfilename := fmt.Sprint(outpath, "/", myimg.Filename, ".thumb.jpg")
	ext := filepath.Ext(myimg.Filename)
	if FileExists(outfilename) {
		log.Printf("thumb already created %s", outfilename)
	} else {
		if ext == ".JPG" {
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
					writeThumb(outpath, myimg.Filename, &img)
				}
			}
		} else if IsStringInSlice(ext, rawExtensions()) {
			log.Printf("loading %s ...\n", myimg.Filename)

			fileInfo, err3 := os.Stat(fmt.Sprintf(myimg.Path, "/", myimg.Filename))
			if err3 != nil {
				log.Fatal(err3)
			}
			img, err2 := golibraw.Raw2Image(appSettings.WorkDir, fileInfo)
			if err2 == nil {
				//result = (result, myImage{Image: img, Filename: f})
				writeThumb(outpath, myimg.Filename, &img)
			} else {
				log.Printf("error decoding %s \n", myimg.Filename)
			}
		}
	}
	myimg.Thumb = outfilename
	return myimg, nil
}

func processImagesInDir(dirname string) ([]myImage, error) {
	files, err := ioutil.ReadDir(dirname)

	result := make([]myImage, 1)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		outfilename := fmt.Sprint(dirname, "/.rawtool/", f.Name(), ".thumb.jpg")
		ext := filepath.Ext(f.Name())
		if FileExists(outfilename) {
			log.Printf("thumb already created %s", outfilename)
		} else {
			if ext == ".JPG" {
				// I just create the thumbnail
				infile := dirname + "/" + f.Name()
				imgfile, err := os.Open(infile)

				if err != nil {
					fmt.Println(infile + " file not found!")
				} else {
					defer imgfile.Close()
					img, _, err := image.Decode(imgfile)
					if err != nil {
						log.Printf("errore %v \n", err)
					} else {
						writeAsThumb(f, &img)
					}
				}
			} else if IsStringInSlice(ext, rawExtensions()) {
				log.Printf("loading %s ...\n", f.Name())
				img, err2 := golibraw.Raw2Image(appSettings.WorkDir, f)
				if err2 == nil {
					//result = append(result, myImage{Image: img, Filename: f})
					writeAsThumb(f, &img)
				} else {
					log.Printf("error decoding %s \n", f.Name())
				}
			}
		}
	}
	log.Printf("reading dir done\n")
	/*
		for _, f := range files {
			//fmt.Printf("%s", f.Name())
			if IsStringInSlice(filepath.Ext(f.Name()), extensions()) {
				exportedImage, _ := libraw.ExportEmbeddedJPEG(dirname, f, exportPath)
				fmt.Printf("exported image %s \n", exportedImage)
			} else if filepath.Ext(f.Name()) == ".JPG" {
				copyFile(dirname+"/"+f.Name(), exportPath+"/"+f.Name())
				fmt.Printf("copyed image %s \n", f.Name())
			}
		}
	*/
	return result, nil
}

func writeThumb(path string, filename string, img *image.Image) error {
	t0 := time.Now()
	var opt jpeg.Options

	opt.Quality = 75
	// ok, write out the data into the new JPEG file

	rand.Seed(time.Now().UTC().UnixNano())
	outfilename := fmt.Sprint(appSettings.WorkDir, "/", filename, ".thumb.jpg")
	out, err := os.Create(outfilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check err

	newImage := resize.Resize(1280, 0, *img, resize.Lanczos3)

	// Encode uses a Writer, use a Buffer if you need the raw []byte
	//err = jpeg.Encode(someWriter, newImage, nil)

	err = jpeg.Encode(out, newImage, &opt) // put quality to 80%
	if err != nil {
		fmt.Println(err)
		return err
	}
	log.Printf("    created and saved thumbnail %s, required %v", filename, time.Since(t0))
	return nil
}

func writeAsThumb(filename os.FileInfo, img *image.Image) error {
	t0 := time.Now()
	var opt jpeg.Options

	opt.Quality = 75
	// ok, write out the data into the new JPEG file

	rand.Seed(time.Now().UTC().UnixNano())
	outfilename := fmt.Sprint(appSettings.WorkDir, "/", filename.Name(), ".thumb.jpg")
	out, err := os.Create(outfilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check err

	newImage := resize.Resize(1280, 0, *img, resize.Lanczos3)

	// Encode uses a Writer, use a Buffer if you need the raw []byte
	//err = jpeg.Encode(someWriter, newImage, nil)

	err = jpeg.Encode(out, newImage, &opt) // put quality to 80%
	if err != nil {
		fmt.Println(err)
		return err
	}
	log.Printf("    created and saved thumbnail %s, required %v", filename.Name(), time.Since(t0))
	return nil
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

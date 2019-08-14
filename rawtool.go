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
	"time"

	"github.com/enricod/golibraw"
	"github.com/gotk3/gotk3/gtk" //"github.com/gotk3/gotk3/gtk"
	"github.com/nfnt/resize"
	"github.com/therecipe/qt/widgets"
)

func extensions() []string {
	return []string{".ORF", ".CR2", ".RAF", ".ARW"}
}

type Settings struct {
	ImagesDir string
	WorkDir   string
	OutDir    string
}

type myImage struct {
	Image    image.Image
	Filename os.FileInfo
}

var appSettings Settings
var flowbox *gtk.FlowBox

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	//defer profile.Start(profile.MemProfile).Stop()

	imagesdir := flag.String("d", ".", "outputdir")
	flag.Parse()

	dir, err := filepath.Abs(filepath.Dir(*imagesdir))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	outdir := fmt.Sprint(dir, "/", ".rawtool")
	createDirIfNotExist(outdir)

	appSettings = Settings{ImagesDir: *imagesdir, WorkDir: dir, OutDir: outdir}

	go readImagesInDir(appSettings.WorkDir)

	// needs to be called once before you can start using the QWidgets
	app := widgets.NewQApplication(len(os.Args), os.Args)

	// create a window
	// with a minimum size of 250*200
	// and sets the title to "Hello Widgets Example"
	window := widgets.NewQMainWindow(nil, 0)
	window.SetMinimumSize2(250, 200)
	window.SetWindowTitle("Hello Widgets Example")

	// create a regular widget
	// give it a QVBoxLayout
	// and make it the central widget of the window
	widget := widgets.NewQWidget(nil, 0)
	widget.SetLayout(widgets.NewQVBoxLayout())
	window.SetCentralWidget(widget)

	// create a line edit
	// with a custom placeholder text
	// and add it to the central widgets layout
	input := widgets.NewQLineEdit(nil)
	input.SetPlaceholderText("Write something ...")
	widget.Layout().AddWidget(input)

	// create a button
	// connect the clicked signal
	// and add it to the central widgets layout
	button := widgets.NewQPushButton2("and click me!", nil)
	button.ConnectClicked(func(bool) {
		widgets.QMessageBox_Information(nil, "OK", input.Text(), widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
	})
	widget.Layout().AddWidget(button)

	// make the window visible
	window.Show()

	// start the main Qt event loop
	// and block until app.Exit() is called
	// or the window is closed by the user
	app.Exec()

}

// IsStringInSlice true if the slice contains the string a
func IsStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func readImagesInDir(dirname string) ([]myImage, error) {
	files, err := ioutil.ReadDir(dirname)

	result := make([]myImage, 1)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if IsStringInSlice(filepath.Ext(f.Name()), extensions()) {
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

func writeAsThumb(filename os.FileInfo, img *image.Image) error {
	t0 := time.Now()
	var opt jpeg.Options

	opt.Quality = 75
	// ok, write out the data into the new JPEG file

	rand.Seed(time.Now().UTC().UnixNano())
	outfilename := fmt.Sprint(appSettings.OutDir, "/", filename.Name(), ".thumb.jpg")
	out, err := os.Create(outfilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check err

	newImage := resize.Resize(1024, 0, *img, resize.Lanczos3)

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

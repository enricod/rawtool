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
	"github.com/gotk3/gotk3/gtk"
	"github.com/nfnt/resize"
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
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Simple Example")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Create a new label widget to show in the window.

	// Add the label to the window.
	win.Add(mainPanel())

	// Set the default window size.
	win.SetDefaultSize(800, 600)

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()

	/*
		fileInfo, err := os.Stat("./_7070001.ORF")
		if os.IsNotExist(err) {
			fmt.Printf("errore caricamento %s %s", fileInfo.Name(), err.Error())
		} else {
			fmt.Printf("caricamento %s\n", fileInfo.Name())
			golibraw.Raw2Image(".", fileInfo)

		}
	*/

}

func mainPanel() *gtk.Widget {

	horBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	if err != nil {
		log.Fatal("Unable to create horBox:", err)
	}
	horBox.SetHomogeneous(false)

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	grid.SetRowSpacing(6)
	grid.SetMarginStart(6)
	grid.SetMarginTop(6)

	entry, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("Unable to create entry:", err)
	}
	s, _ := entry.GetText()
	label, err := gtk.LabelNew(s)
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	grid.Add(entry)
	entry.SetHExpand(true)
	grid.AttachNextTo(label, entry, gtk.POS_RIGHT, 1, 1)
	label.SetHExpand(true)

	// Connects this entry's "activate" signal (which is emitted whenever
	// Enter is pressed when the Entry is activated) to an anonymous
	// function that gets the current text of the entry and sets the text of
	// the label beside it with it.  Unlike with native GTK callbacks,
	// (*glib.Object).Connect() supports closures.  In this example, this is
	// demonstrated by using the label variable.  Without closures, a
	// pointer to the label would need to be passed in as user data
	// (demonstrated in the next example).
	entry.Connect("activate", func() {
		s, _ := entry.GetText()
		label.SetText(s)
	})

	sb, err := gtk.SpinButtonNewWithRange(0, 1, 0.1)
	if err != nil {
		log.Fatal("Unable to create spin button:", err)
	}
	/*
		pb, err := gtk.ProgressBarNew()
		if err != nil {
			log.Fatal("Unable to create progress bar:", err)
		}
	*/
	grid.Add(sb)
	sb.SetHExpand(true)
	//grid.AttachNextTo(pb, sb, gtk.POS_RIGHT, 1, 1)
	label.SetHExpand(true)

	// Pass in a ProgressBar and the target SpinButton as user data rather
	// than using the sb and pb variables scoped to the anonymous func.
	// This can be useful when passing in a closure that has already been
	// generated, but when you still wish to connect the callback with some
	// variables only visible in this scope.
	/*
		sb.Connect("value-changed", func(sb *gtk.SpinButton, pb *gtk.ProgressBar) {
			pb.SetFraction(sb.GetValue() / 1)
		}, pb)
	*/
	label, err = gtk.LabelNew("")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	s = "Hyperlink to <a href=\"https://www.cyphertite.com/\">Cyphertite</a> for your clicking pleasure"
	label.SetMarkup(s)
	grid.AttachNextTo(label, sb, gtk.POS_BOTTOM, 2, 1)

	dirChooserBtn, err := gtk.FileChooserButtonNew("Dir selection", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)
	if err != nil {
		log.Fatal("Unable to create FileChooserDialogNewWith1Button:", err)

	}
	dirChooserBtn.Connect("selection-changed", dirSelectionChanged)
	// grid.Add(dirChooserBtn)
	grid.AttachNextTo(dirChooserBtn, label, gtk.POS_BOTTOM, 3, 1)

	// Some GTK callback functions require arguments, such as the
	// 'gchar *uri' argument of GtkLabel's "activate-link" signal.
	// These can be used by using the equivalent go type (in this case,
	// a string) as a closure argument.
	//label.Connect("activate-link", func(_ *gtk.Label, uri string) {
	//	fmt.Println("you clicked a link to:", uri)
	//})

	horBox.PackStart(grid, false, true, 6)

	flowbox, err = gtk.FlowBoxNew()
	if err != nil {
		log.Fatal("Unable to create FileChooserDialogNewWith1Button:", err)

	}

	// popolaFlowbox(flowbox)
	horBox.PackStart(flowbox, true, true, 6)
	return &horBox.Container.Widget
	//return &grid.Container.Widget
}

func dirSelectionChanged(widget *gtk.FileChooserButton) {
	fmt.Printf("dir selected %s\n", widget.GetFilename())
	//appSettings.ImagesDir = widget.GetFilename()
	//DoExtract(widget.GetFilename())
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

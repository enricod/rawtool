package main

import (
	"log"

	"github.com/gotk3/gotk3/gtk"
)

var dirNameLabel *gtk.Label

func setPicturesDir(dirname string) {
	dirNameLabel.SetText(dirname)
	appSettings.ImagesDir = dirname
}

func scegliDir(btn *gtk.Button) {
	dialog, err := gtk.FileChooserDialogNewWith2Buttons(
		"images dir",
		nil, // parent window - non so casa dare
		gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER,
		"Select",
		gtk.RESPONSE_OK,
		"Cancel",
		gtk.RESPONSE_CANCEL)

	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	resp := dialog.Run()

	if resp == gtk.RESPONSE_OK {
		setPicturesDir(dialog.GetFilename())
	} else {
		log.Printf("%v", resp)
	}
	dialog.Destroy()
}

func startGui() {
	// Initialize GTK without parsing any command line arguments.
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

	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		log.Fatal("Unable to create mainBox:", err)
	}

	colonnaSx, err := creaColonnaSinistra()
	colonnaCentrale, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	colonnaDestra, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)

	// Create a new label widget to show in the window.

	l2, err := gtk.LabelNew("Hello, centrale!")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	l3, err := gtk.LabelNew("Hello, centrale!")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	colonnaCentrale.Add(l2)
	colonnaDestra.Add(l3)

	mainBox.Add(colonnaSx)
	mainBox.Add(colonnaCentrale)
	mainBox.PackEnd(colonnaDestra, false, false, 10)

	// Add the label to the window.
	win.Add(mainBox)

	// Set the default window size.
	win.SetDefaultSize(800, 600)

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}

func avviaScansione() {
	ProcessDir(appSettings)
}

func creaColonnaSinistra() (*gtk.Box, error) {
	result, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	dirNameLabel, err = gtk.LabelNew(appSettings.ImagesDir)
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	dirBtn, err := gtk.ButtonNewWithLabel("Change Dir")
	// l, err := gtk.LabelNew("Hello, enrico!")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}

	scanBtn, err := gtk.ButtonNewWithLabel("Scan")
	// l, err := gtk.LabelNew("Hello, enrico!")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}

	dirBtn.Connect("clicked", scegliDir)
	scanBtn.Connect("clicked", avviaScansione)

	result.Add(dirNameLabel)
	result.Add(dirBtn)
	result.Add(scanBtn)
	return result, err
}

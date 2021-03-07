package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/mdirkse/i3ipc"
)

func main() {

	var verbose = flag.Bool("v", false, "verbose")
	flag.Parse()

	ipcsocket, err := i3ipc.GetIPCSocket()
	if err != nil {
		log.Fatalln(err)
	}

	tree, err := ipcsocket.GetTree()
	if err != nil {
		log.Fatalln(err)
	}

	current := tree.FindFocused()
	if current == nil {
		log.Fatalln("Cannot find focused")
	}

	leaves := current.Workspace().Leaves()
	var biggest *i3ipc.I3Node = current

	for _, l := range leaves {
		win := tree.FindByID(l.ID)
		if win.Rect.Width > biggest.Rect.Width ||
			(win.Rect.Width == biggest.Rect.Width &&
				win.Rect.Height > biggest.Rect.Height) {
			biggest = win
		}
		if *verbose {
			log.Printf("%s [%dx%d]", win.Name, win.Rect.Width, win.Rect.Height)
		}
	}

	if *verbose {
		log.Printf("Biggest win: %s", biggest.Name)
	}

	if biggest != nil && biggest != current {
		_, err := ipcsocket.Command("swap container with con_id " + strconv.FormatInt(biggest.ID, 10))
		if err != nil {
			log.Println(err)
		}
	}
}

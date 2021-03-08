package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"unsafe"

	"bitbucket.org/avd/go-ipc/mmf"
	"bitbucket.org/avd/go-ipc/shm"
	"github.com/mdirkse/i3ipc"
)

const (
	shared_memory_name string = "swap_biggest_mem"
)

func open_mem() *mmf.MemoryRegion {
	sz := int64(unsafe.Sizeof(int64(0)))

	obj, _, err := shm.NewMemoryObjectSize(shared_memory_name, os.O_CREATE|os.O_RDWR, 0666, sz)
	if err != nil {
		log.Fatal(err)
	}

	reg, err := mmf.NewMemoryRegion(obj, mmf.MEM_READWRITE, 0, int(sz))
	if err != nil {
		log.Fatal(err)
	}

	return reg
}

func save(id int64) {
	reg := open_mem()

	b := [8]byte{0}
	for i := 0; i < 8; i++ {
		b[i] = byte(0xff & (id >> (i * 8)))
	}

	copy(reg.Data(), b[:])
}

func read() int64 {
	reg := open_mem()
	var t int64

	for i := 0; i < 8; i++ {
		t = t + (int64(reg.Data()[i]) << (i * 8))
	}

	return t
}

type vPrinter func(format string, v ...interface{})

func get_verbose_print(verbose bool) vPrinter {
	return func(format string, v ...interface{}) {
		if verbose {
			fmt.Printf(format, v...)
		}
	}
}

func get_tree(ipcsocket *i3ipc.IPCSocket) i3ipc.I3Node {
	tree, err := ipcsocket.GetTree()
	if err != nil {
		log.Fatalln(err)
	}
	return tree
}

func get_current_and_biggest_neighbour(tree i3ipc.I3Node) (*i3ipc.I3Node, *i3ipc.I3Node) {
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
	}
	return current, biggest
}

func main() {
	var verbose = flag.Bool("v", false, "verbose")
	var back = flag.Bool("b", false, "swap back")
	flag.Parse()

	var printf = get_verbose_print(*verbose)

	ipcsocket, err := i3ipc.GetIPCSocket()
	if err != nil {
		log.Fatalln(err)
	}

	tree := get_tree(ipcsocket)
	current, biggest := get_current_and_biggest_neighbour(tree)

	printf("current: %s [%dx%d]\nbiggest: %s [%dx%d] %d\n", current.Name, current.Rect.Width, current.Rect.Height, biggest.Name, biggest.Rect.Width, biggest.Rect.Height, biggest.ID)

	if (current == biggest) && *back {
		// we need to swap back
		// check stored id
		last_id := read()
		last_node := tree.FindByID(last_id)
		if last_node == nil {
			printf("Window has been closed/moved\n")
			return
		}

		command := fmt.Sprintf("swap container with con_id %d", last_id)
		ipcsocket.Command(command)
	} else {
		ss, err := ipcsocket.Command("swap container with con_id " + strconv.FormatInt(biggest.ID, 10))
		if err != nil {
			log.Println(err)
		}
		if ss && *back {
			save(biggest.ID)
			printf("save %d\n", biggest.ID)
		}
	}
}

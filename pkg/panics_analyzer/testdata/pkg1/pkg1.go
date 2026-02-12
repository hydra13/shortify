package pkg1

import (
	"log"
	"os"
)

func panicCheckFunc() {
	panic("testing") // want "call of panic()"
}

func logFatalCheckFunc() {
	log.Fatalln("testing") // want "call of log.Fatalln()"
}

func osExitCheckFunc() {
	os.Exit(-1) // want "call of os.Exit"
}

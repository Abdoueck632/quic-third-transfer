package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	//myKey := []byte{25, 100, 176, 211, 81, 230, 19, 178, 109, 58, 234, 145, 168, 237, 81, 151}
	// If previous cache present, let it as it
	chaine := "Bonjour"
	f, err := os.Create("./data.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(chaine)

	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("done")

}

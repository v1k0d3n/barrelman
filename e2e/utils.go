package e2e

import (
	"fmt"
	"log"
	"os"
)

func fullPath() string {
	barrelmanPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(barrelmanPath)
	return barrelmanPath
}

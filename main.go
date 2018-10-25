package main

import (
	"fmt"

	"github.com/charter-se/barrelman/yamlpack"
)

func main() {
	fmt.Printf("Barrelman Engage!\n")
	yp := yamlpack.New()
	if err := yp.Import("testdata/armada-osh.yaml"); err != nil {
		fmt.Printf("Error importing \"this\": %v\n", err)
	}

	for name, f := range yp.Files {
		fmt.Println("_________________________")
		fmt.Printf("This: %v\n", name)
		for _, k := range f {
			fmt.Printf("Schema: %v\n", k.Viper.Get("schema"))
			fmt.Printf("Metdata Name: %v\n", k.Viper.Get("metadata.name"))
			fmt.Printf("Metdata Schema: %v\n", k.Viper.Get("metadata.schema"))
			y, err := k.Yaml()
			if err != nil {
				fmt.Printf("Failed to marshal data: %v\n", err)
				return
			}
			fmt.Printf("Data:\n%v\n", string(y))
		}
		fmt.Printf("\n\n")
	}
	fmt.Printf("Yaml Sections:\n")
	for _, s := range yp.ListYamls() {
		fmt.Printf("\t%v\n", s)
	}
}

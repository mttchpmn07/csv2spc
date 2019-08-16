package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mttchpmn07/spctools/pkg/spcgo"
)

func checkEmpty(s string) string {
	if s == "" {
		return "empty string"
	}
	return s
}

func main() {

	argsWithoutProg := os.Args[1:]
	for i, filename := range argsWithoutProg {

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("File <%s> does not exist\n", checkEmpty(filename))
		} else {
			fmt.Printf("Reading <%s> |  %d of %d\n", filename, i, len(argsWithoutProg))
		}

		csvFilename := strings.TrimSuffix(filename, filepath.Ext(filename))
		csvFilename = csvFilename + ".csv"

		SPC := spcgo.ReadSPC(filename, false)

		var numpts int32 = 5
		reportdata := true
		if reportdata {
			fmt.Printf("\nData in file:\n")
			for i := int32(0); i < numpts; i++ {
				fmt.Printf("%d: %f, %f\n", i+1, (*SPC.Data.X)[i], (*SPC.Data.Y)[i])
			}
			fmt.Printf("...\n...\n...\n")
			for i := SPC.Head.Fnpts - numpts; i < SPC.Head.Fnpts; i++ {
				fmt.Printf("%d: %f, %f\n", i+1, (*SPC.Data.X)[i], (*SPC.Data.Y)[i])
			}
		}

		spcgo.SaveCSV(SPC, csvFilename)
	}
}

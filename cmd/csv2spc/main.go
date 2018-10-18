package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	b, err := ioutil.ReadFile("c:/Users/ua17996/go/src/github.com/mttchpmn07/spctools/cmd/csv2spc/test.csv") // b has type []byte
	if err != nil {
		log.Fatal(err)
	}
	in := string(b)

	r := csv.NewReader(strings.NewReader(in))

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Print(len(records))
	//i := 1
	//raw := records[19:21]
	//fmt.Print(raw)
	for i := 1; i < len(records[0]); i += 5 {
		fmt.Printf("%d\t%v\n", i, records[i][0])
	}
}

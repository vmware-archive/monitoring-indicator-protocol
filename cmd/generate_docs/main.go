package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"github.com/cloudfoundry-incubator/event-producer/pkg/kpi"
	"github.com/cloudfoundry-incubator/event-producer/pkg/docs"
)

func main() {
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("cannot read file %s, %s", os.Args[1], err)
	}

	kpis, err := kpi.ReadKPIsFromYaml(file)
	if err != nil {
		log.Fatalf("cannot parse file %s", err)
	}

	for _, indicator := range kpis {
		html, err := docs.HTML(indicator)
		if err != nil {
			log.Fatalln("error formatting indicator: ", err)
		}
		fmt.Printf(html + "\n")
	}
}

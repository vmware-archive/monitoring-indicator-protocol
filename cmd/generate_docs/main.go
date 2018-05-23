package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"log"

	"github.com/cloudfoundry-incubator/event-producer/pkg/indicator"
	"github.com/cloudfoundry-incubator/event-producer/pkg/docs"
)

func main() {
	fileBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("cannot read file %s, %s", os.Args[1], err)
	}

	d, err := indicator.ReadIndicatorDocument(fileBytes)
	if err != nil {
		log.Fatalf("cannot parse file %s", err)
	}

	for _, indicator := range d.Indicators {
		html, err := docs.IndicatorToHTML(indicator)
		if err != nil {
			log.Fatalln("error formatting indicator: ", err)
		}
		fmt.Printf(html + "\n")
	}

	for _, m := range d.Metrics {
		html, err := docs.MetricToHTML(m)
		if err != nil {
			log.Fatalln("error formatting metric: ", err)
		}
		fmt.Printf(html + "\n")
	}
}

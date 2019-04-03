package main

import (
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

var (
	isbnColumn    = kingpin.Flag("isbnColumn", "The column containing the isbn (first column starting at 1)").Default("1").Int()
	withHeaderRow = kingpin.Flag("withHeaderRow", "Indicates if a header row should be read from the input and written in the output").Default("true").Bool()
	priceRegexp   = regexp.MustCompile("row_price_(\\d+.\\d+)_")
)

func main() {
	kingpin.Version("1.0.0")
	kingpin.Parse()
	r := csv.NewReader(os.Stdin)

	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("Error parsing csv: %s", err.Error())
	}

	headers := records[0]
	fmt.Printf("header: %s\n", headers)

	isbn := records[1][*isbnColumn-1]
	fmt.Printf("ISBN: %s\n", isbn)

	page, err := downloadPriceURL(fmt.Sprintf("https://isbn.nu/%s", isbn))
	if err != nil {
		log.Fatalf("Error loading url to get prices: %s", err.Error())
	}

	matches := priceRegexp.FindAllSubmatch(page, -1)
	log.Println(getLowestAndHighestPricesFromMatches(matches))

	// Find all news item.

	// for record, _ := r.Read(); record != nil; record, _ = r.Read() {
	// 	fmt.Printf("ISBN: %s\n", record[*isbnColumn-1])
	// }
}

func downloadPriceURL(priceURL string) (content []byte, err error) {
	resp, err := http.Get(priceURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading url [%s]", priceURL)
	}

	b, err := ioutil.ReadAll(resp.Body)

	return b, err
}

func getLowestAndHighestPricesFromMatches(matches [][][]byte) (lowest string, average string, highest string) {
	for _, m := range matches {
		price := string(m[0])

		log.Printf("price: %s\n", price)
	}

	return "0.0", "0.0", "0.0"
}

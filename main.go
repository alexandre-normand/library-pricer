package main

import (
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	isbnColumn    = kingpin.Flag("isbnColumn", "The column containing the isbn to use when looking up prices (first column starting at 1)").Required().Int()
	withHeaderRow = kingpin.Flag("withHeaderRow", "Indicates if a header row should be read from the input and written in the output").Default("true").Bool()
	throttleTime  = kingpin.Flag("throttleTime", "Sleep time between calls to isbn.nu (be easy on them).").Default("15s").Short('t').Duration()
)

var (
	priceRegexp             = regexp.MustCompile("row_price_(\\d+.\\d+)_")
	originalListPriceRegexp = regexp.MustCompile("<span class=\"bi_col_value\">\\$(\\d+.\\d+)<")
	logger                  = log.New(os.Stderr, "", 0)
)

func main() {
	kingpin.Version("1.0.0")
	kingpin.Parse()
	r := csv.NewReader(os.Stdin)

	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("Error parsing csv: %s", err.Error())
	}

	w := csv.NewWriter(os.Stdout)

	for i, record := range records {
		if i == 0 && *withHeaderRow {
			w.Write(outputHeaderFor(record))
		} else {
			isbn := record[*isbnColumn-1]

			priceURL := ""
			min := math.NaN()
			max := math.NaN()
			avg := math.NaN()
			listPrice := math.NaN()

			if len(isbn) > 0 {
				priceURL = fmt.Sprintf("https://isbn.nu/%s", isbn)
				page, err := downloadPricePage(fmt.Sprintf("https://isbn.nu/%s", isbn))
				if err != nil {
					logger.Fatalf("Error loading url [%s] to get prices: %s", priceURL, err.Error())
				}

				min, avg, max, err = getPrices(page)
				if err != nil {
					logger.Printf("Error getting prices for isbn [%s] from url [%s]: %s, page content:\n%s\n\n", isbn, priceURL, err.Error(), page)
				}

				listPrice, err = getListPrice(page)
				if err != nil {
					logger.Printf("Error getting original list price for isbn [%s] from url [%s]: %s, page content:\n%s\n", isbn, priceURL, err.Error(), page)
				}
			}

			w.Write(outputDataRowFor(record, priceURL, min, avg, max, listPrice))
			w.Flush()

			time.Sleep(*throttleTime)
		}
	}

	w.Flush()
}

// outputHeaderFor takes a input header record and adds the additional price column headers: price_url, min_price, average_price, max_price and list_price
func outputHeaderFor(inHeader []string) (outHeader []string) {
	outHeader = make([]string, 0)

	outHeader = append(outHeader, inHeader...)
	outHeader = append(outHeader, "price_url")
	outHeader = append(outHeader, "min_price")
	outHeader = append(outHeader, "average_price")
	outHeader = append(outHeader, "max_price")
	outHeader = append(outHeader, "list_price")

	return outHeader
}

// outputDataRowFor takes an input record and adds the additional price columns: price url, min, average, max and original list price
func outputDataRowFor(inRecord []string, url string, min float64, average float64, max float64, listPrice float64) (outRecord []string) {
	outRecord = make([]string, 0)

	outRecord = append(outRecord, inRecord...)
	outRecord = append(outRecord, url)
	if math.IsNaN(min) {
		outRecord = append(outRecord, "")
	} else {
		outRecord = append(outRecord, fmt.Sprintf("$%.02f", min))
	}

	if math.IsNaN(average) {
		outRecord = append(outRecord, "")
	} else {
		outRecord = append(outRecord, fmt.Sprintf("$%.02f", average))
	}

	if math.IsNaN(max) {
		outRecord = append(outRecord, "")
	} else {
		outRecord = append(outRecord, fmt.Sprintf("$%.02f", max))
	}

	if math.IsNaN(listPrice) {
		outRecord = append(outRecord, "")
	} else {
		outRecord = append(outRecord, fmt.Sprintf("$%.02f", listPrice))
	}

	return outRecord
}

// downloadPricePage fetches the html for the priceURL and returns the full page
func downloadPricePage(priceURL string) (content []byte, err error) {
	resp, err := http.Get(priceURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading url [%s]", priceURL)
	}

	return ioutil.ReadAll(resp.Body)
}

// getListPrice tries to find the original list price in the page. Since it's not
// always available, an error is returned when no match is found along with NaN
func getListPrice(page []byte) (originalPrice float64, err error) {
	matches := originalListPriceRegexp.FindAllSubmatch(page, -1)

	if len(matches) > 0 && len(matches[0]) > 1 {
		originalVal, err := strconv.ParseFloat(string(matches[0][1]), 32)
		if err != nil {
			return 0., err
		}

		if originalVal > 0 {
			return originalVal, nil
		}

		return math.NaN(), fmt.Errorf("List price not available")
	}

	return math.NaN(), fmt.Errorf("Can't find match for original price")
}

// getPrices finds matches for price listings from various online stores listed on a page.
// The min, average and max values are returned if at least one price is found. Otherwise,
// NaN values are returned along with an error
func getPrices(page []byte) (min float64, average float64, max float64, err error) {
	matches := priceRegexp.FindAllSubmatch(page, -1)

	min = math.MaxFloat32
	max = 0.
	count := 0
	sum := 0.

	for _, m := range matches {
		val := string(m[1])

		price, err := strconv.ParseFloat(val, 32)
		if err == nil && price > 0 {
			count = count + 1
			sum = sum + price
			min = math.Min(min, price)
			max = math.Max(max, price)
		}
	}

	if count > 0 {
		return min, sum / float64(count), max, nil
	}

	return math.NaN(), math.NaN(), math.NaN(), fmt.Errorf("Couldn't find any prices in page")
}

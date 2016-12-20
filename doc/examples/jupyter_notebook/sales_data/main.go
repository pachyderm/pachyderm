package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const price = 5

var (
	inDir  = "/pfs/trips/"
	outDir = "/pfs/out/"
)

func main() {

	if err := filepath.Walk(inDir, func(path string, info os.FileInfo, err error) error {

		// Open the data file.
		f, err := os.Open(filepath.Join(inDir, info.Name()))
		if err != nil {
			return err
		}
		defer f.Close()

		// Create a new CSV reader reading from the opened file.
		reader := csv.NewReader(f)

		// Assume we don't know the number of fields per line.  By setting
		// FieldsPerRecord negative, each row may have a variable
		// number of fields.
		reader.FieldsPerRecord = -1

		// Read in all of the CSV records.
		rawCSVData, err := reader.ReadAll()
		if err != nil {
			return err
		}

		// Calculate and output the sales.
		var sales int
		var date string
		for idx, each := range rawCSVData {

			// Skip the header.
			if idx == 0 {
				continue
			}

			// Calculate the sales.
			trips, err := strconv.Atoi(each[1])
			if err != nil {
				return err
			}

			// Calculate the sales.
			sales = trips * price
			date = each[0]
		}

		// Create the output file if it doesn't exist.
		if _, err := os.Stat("/pfs/out/sales.csv"); os.IsNotExist(err) {

			outFile, err := os.Create("/pfs/out/sales.csv")
			if err != nil {
				return err
			}
			outFile.Close()
		}

		// Format the output line.
		output := fmt.Sprintf("%s,%d\n", date, sales)

		outFile, err := os.OpenFile(filepath.Join(outDir, "sales.csv"), os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// Append the output
		if _, err = outFile.WriteString(output); err != nil {
			return err
		}

		return nil

	}); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/signintech/gopdf"
	"github.com/signintech/gopdf/fontmaker/core"
)

const (
	// FontName the name of the font to use
	FontName string = "SourceCodePro-Regular"
)

func main() {

	file, err := os.Open("tmp/sample.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var headers []string
	var prices []string
	userData := make([][]string, 0)
	r := csv.NewReader(file)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if headers == nil {
			headers = record
		} else if prices == nil {
			prices = record
		} else {
			userData = append(userData, record)
		}
		fmt.Println(record)
	}
	generatePdf(headers, prices, userData)
}

func generatePdf(headers, prices []string, userData [][]string) {
	pdf := gopdf.GoPdf{}
	pageSize := gopdf.Rect{W: 595.28, H: 841.89}
	pdf.Start(gopdf.Config{PageSize: pageSize}) //595.28, 841.89 = A4
	fontLocation := fmt.Sprintf("./ttf/%s.ttf", FontName)
	var parser core.TTFParser
	err := parser.Parse(fontLocation)
	if err != nil {
		log.Print(err.Error())
		return
	}
	err = pdf.AddTTFFont(FontName, fontLocation)
	if err != nil {
		log.Print(err.Error())
		return
	}
	// pages
	for _, userData := range userData {
		pdf.AddPage()
		// title
		fontSize := 24
		err = pdf.SetFont(FontName, "", fontSize)
		if err != nil {
			log.Print(err.Error())
			return
		}
		pdf.Cell(nil, "Givrés d'Oranges - Décembre 2017")
		pdf.Br(getHeight(&parser, fontSize) * 2)
		userName := strings.ToTitle(userData[0])
		pdf.Cell(nil, userName)
		pdf.Br(getHeight(&parser, fontSize) * 2)
		// items
		fontSize = 18
		err = pdf.SetFont(FontName, "", fontSize)
		if err != nil {
			log.Print(err.Error())
			return
		}
		totalPrice := 0.0
		for i, header := range headers {
			if i == 0 {
				continue
			}
			price, err := strconv.ParseFloat(prices[i], 64)
			if err != nil {
				log.Print(err.Error())
				return
			}
			var quantity int
			if userData[i] == "" {
				quantity = 0
			} else {
				quantity, err = strconv.Atoi(userData[i])
				if err != nil {
					log.Printf("Unable to convert `%s` to a valid quantity: %s", userData[i], err.Error())
					return
				}
			}
			fmt.Printf("Item `%s`: quantity=%d / price=%f\n", header, quantity, price)
			itemDescription := fmt.Sprintf("%s (%.2f€)", header, price)
			billItem := fmt.Sprintf("%s%s%d", itemDescription,
				strings.Repeat(".", 40-utf8.RuneCountInString(itemDescription)-utf8.RuneCountInString(strconv.Itoa(quantity))),
				quantity)
			pdf.Cell(nil, billItem)
			// next line
			pdf.Br(getHeight(&parser, fontSize))
			totalPrice += float64(quantity) * price
		}
		pdf.Br(getHeight(&parser, fontSize) * 2)
		totalPriceStr := fmt.Sprintf("Prix total: %.2f€", totalPrice)
		pdf.Cell(nil, fmt.Sprintf("%s%s", strings.Repeat(" ", 40-utf8.RuneCountInString(totalPriceStr)), totalPriceStr))

	}

	// write output
	pdf.WritePdf("tmp/result.pdf")
}

func getHeight(parser *core.TTFParser, fontSize int) float64 {
	//Measure Height
	//get  CapHeight (https://en.wikipedia.org/wiki/Cap_height)
	cap := float64(float64(parser.CapHeight()) * 1000.00 / float64(parser.UnitsPerEm()))
	//convert
	realHeight := cap * (float64(fontSize) / 1000.0)
	// fmt.Printf("realHeight = %f", realHeight)
	return realHeight * 2
}

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var fieldmsg = `comma separated order of fields in csv file. Valid fields:
    date, payee, amount, and note.`

var (
	fieldspec = flag.String("fields", "", fieldmsg)
	date      = flag.String("date", "1/2/2006", "format of dates in csv file")
	account   = flag.String("account", "Assets:", "account of all the transactions")
	category  = flag.String("category", "Expenses:", "category (expense/income) account")
	header    = flag.Bool("header", true, "true if the csv file has an initial header line")
)

func main() {
	flag.Parse()

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fields := strings.Split(*fieldspec, ",")
	for i, s := range fields {
		fields[i] = strings.TrimSpace(s)
	}

	if *header {
		records = records[1:]
	}

	for _, rec := range records {
		vals := map[string]string{}
		for i := range fields {
			vals[fields[i]] = rec[i]
		}

		t := time.Now()
		if vals["date"] != "" {
			t, err = time.Parse(*date, vals["date"])
			if err != nil {
				log.Fatal(err)
			}
		}

		amt, err := strconv.ParseFloat(vals["amount"], 64)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v %v ; \n", t.Format("2006/01/02"), vals["payee"])
		fmt.Printf("    %v        $%.2f\n", *account, amt)
		fmt.Printf("    ; %v\n", vals["note"])
		fmt.Printf("    %v\n\n", *category)

	}
}

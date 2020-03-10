package main

import (
	"log"
	"strconv"

	"github.com/gopherjs/jquery"
)

var jq = jquery.NewJQuery

func main() {
	totalEnts, _ := strconv.Atoi(jq("#saveTotal").Val())
	origMonthEnts, _ := strconv.Atoi(jq("#origMonth").Val())
	log.Print(totalEnts, origMonthEnts)
	jq("#ents").On(jquery.CHANGE, func(e jquery.Event) {
		monthEnts, _ := strconv.Atoi(jq(e.Target).Val())
		jq("#newents").SetVal(totalEnts + monthEnts - origMonthEnts)
	})
}

// vim:foldmethod=marker:

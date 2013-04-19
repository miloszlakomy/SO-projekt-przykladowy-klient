package main

import "flag"
import "ognisko/simple"
import "net/http"

var httpAddr = flag.String("http", "127.0.0.1:5555", "")

func main() {
	flag.Parse()
	simp := simple.NewSimple()
	go simp.Loop()

	http.Handle("/men", (*simple.MenView)(simp))
	http.Handle("/map", (*simple.MapView)(simp))
	http.Handle("/over", (*simple.OverviewView)(simp))
	panic(http.ListenAndServe(*httpAddr, nil))
}


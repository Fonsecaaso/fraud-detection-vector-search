package main

import (
	"fmt"
	"log"
	"net/http"

	v1 "github.com/Fonsecaaso/fraud-detection-vector-search/v1"
	v2 "github.com/Fonsecaaso/fraud-detection-vector-search/v2"
)

func main() {

	// fmt.Println("Initializing V1...")
	// v1.Initialize()
	// fmt.Println("V1 Initialized")

	fmt.Println("Initializing V2...")
	if err := v2.Initialize(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("V2 Initialized")

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/fraud-score", v1.Predict)
	mux.HandleFunc("/v2/fraud-score", v2.Predict)

	log.Fatal(http.ListenAndServe(":9999", mux))
}

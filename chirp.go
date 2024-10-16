package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func decode(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Content string `json:"body"`
	}
	type returnVals struct {
		Data string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	} else if len(params.Content) > 140 {
		type errorVals struct {
			Data string `json:"error"`
		}

		respBody := errorVals{
			Data: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
	} else {
		cleanedData := cleanText(params.Content)
		respBody := returnVals{
			Data: cleanedData,
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(dat)
	}
}

func cleanText(s string) string {
	splittedString := strings.Split(s, " ")
	for idx, element := range splittedString {
		switch strings.ToLower(element) {
		case
			"kerfuffle",
			"sharbert",
			"fornax":
			splittedString[idx] = "****"
		}
	}
	return strings.Join(splittedString, " ")
}

package main

import (
	"encoding/json"
	"net/http"
)

type ParsedCombination struct {
	Numbers    string `json:"field_winning_numbers"`
	Date       string `json:"field_draw_date"`
	Multiplier string `json:"field_multiplier"`
}

// GetRecentPBnumber fetches most recent powerball combination from powerball.com
// and returns ParsedCombination struct
func GetRecentPBnumber() (*ParsedCombination, error) {
	res, err := http.Get("https://www.powerball.com/api/v1/numbers/powerball/recent?_format=json")
	if err != nil {
		return nil, err
	}
	// We expect to get 3 recent powerball drawings
	// Read first one, since it is the most recent
	dec := json.NewDecoder(res.Body)
	dec.Token()
	var n parsedCombination
	err = dec.Decode(&n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

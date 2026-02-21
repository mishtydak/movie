package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func SearchOMDB(query string) ([]OMDBMovie, error) {
	apiKey := os.Getenv("OMDB_API_KEY")

	baseURL := "http://www.omdbapi.com/"

	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("s", query)

	fullURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResponse OMDBSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResponse)
	if err != nil {
		return nil, err
	}

	if searchResponse.Response == "False" {
		return nil, fmt.Errorf(searchResponse.Error)
	}

	return searchResponse.Search, nil
}
func GetMovieDetails(imdbID string) (*OMDBMovieDetail, error) {
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found")
	}

	baseURL := "http://www.omdbapi.com/"

	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("i", imdbID)
	params.Add("plot", "full")

	fullURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var movie OMDBMovieDetail
	err = json.NewDecoder(resp.Body).Decode(&movie)
	if err != nil {
		return nil, err
	}

	if movie.Response == "False" {
		return nil, fmt.Errorf(movie.Error)
	}

	return &movie, nil
}

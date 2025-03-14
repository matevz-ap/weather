package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Response struct {
	Results []Location
}

type Location struct {
	Latitude  float64
	Longitude float64
}

func geocode(location string) (float64, float64, error) {
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", location)
	res, err := http.Get(url)

	if err != nil {
		return 0, 0, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, 0, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, 0, err
	}

	if len(response.Results) == 0 {
		return 0, 0, err
	}

	return response.Results[0].Latitude, response.Results[0].Longitude, nil
}

func locationWeather(lat float64, long float64) (string, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m", lat, long)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func Weather(location string) (string, error) {
	long, lat, err := geocode(location)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	weather_data, err := locationWeather(lat, long)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return weather_data, nil
}

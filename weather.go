package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

const (
	url = "https://forecast.weather.gov/MapClick.php?lat=34.3651&lon=-89.5196&FcstType=digitalDWML"
)

type Result struct {
	XMLName xml.Name `xml:"dwml"`
	Data    Data
}

type Data struct {
	XMLName    xml.Name `xml:"data"`
	Parameters Parameters
	Times      Time
}

type Time struct {
	XMLName xml.Name `xml:"time-layout"`
	Values  []string `xml:"start-valid-time"`
}

type Parameters struct {
	XMLName xml.Name      `xml:"parameters"`
	Temps   []Temperature `xml:"temperature"`
	Winds   []Wind        `xml:"wind-speed"`
	// Percent cloud cover
	Clouds []int `xml:"cloud-amount>value"`
	// Percent chance of rain
	Rain     []int `xml:"probability-of-precipitation>value"`
	Humidity []int `xml:"humidity>value"`
	// Wind direction in degrees true. 360 is north, 90 east and
	// so on
	Direction []int `xml:"direction>value"`
}

type Temperature struct {
	// Type is either "dew point", "wind chill", or "hourly"
	Type  string   `xml:"type,attr"`
	Value []string `xml:"value"`
}

type Wind struct {
	// Type is either "sustained" or "gust"
	Type  string   `xml:"type,attr"`
	Value []string `xml:"value"`
}

func main() {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	var byts []byte
	for scanner := bufio.NewScanner(resp.Body); scanner.Scan(); {
		line := scanner.Text()
		if !strings.Contains(line, "encoding") {
			byts = append(byts, []byte(line+"\n")...)
		}
	}
	// fmt.Printf("%s\n", byts)
	res := new(Result)
	err = xml.Unmarshal(byts, res)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%s\n", res.Data.Parameters.Temps[0])
	for _, t := range res.Data.Parameters.Temps {
		fmt.Println(t.Type)
	}
	for _, t := range res.Data.Parameters.Winds {
		fmt.Println(t.Type)
	}
	fmt.Println(len(res.Data.Parameters.Clouds))
	fmt.Println(len(res.Data.Parameters.Rain))
	fmt.Println(len(res.Data.Times.Values))
	fmt.Println(len(res.Data.Parameters.Humidity))
	fmt.Println(len(res.Data.Parameters.Direction))
}

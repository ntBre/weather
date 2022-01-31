package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	url      = "https://forecast.weather.gov/MapClick.php?lat=34.3651&lon=-89.5196&FcstType=digitalDWML"
	longFmt  = "2006-01-02T15:04:05-07:00"
	shortFmt = "01-02/15:00"
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
	Type  string `xml:"type,attr"`
	Value []int  `xml:"value"`
}

type Wind struct {
	// Type is either "sustained" or "gust"
	Type  string `xml:"type,attr"`
	Value []int  `xml:"value"`
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
	res := new(Result)
	err = xml.Unmarshal(byts, res)
	if err != nil {
		panic(err)
	}
	temps := make(map[string][]int)
	for _, t := range res.Data.Parameters.Temps {
		temps[t.Type] = t.Value
	}
	winds := make(map[string][]int)
	for _, t := range res.Data.Parameters.Winds {
		winds[t.Type] = t.Value
	}
	// fmt.Println(len(res.Data.Parameters.Clouds))
	// fmt.Println(len(res.Data.Parameters.Rain))
	// fmt.Println(len(res.Data.Times.Values))
	// fmt.Println(len(res.Data.Parameters.Humidity))
	// fmt.Println(len(res.Data.Parameters.Direction))
	now := time.Now().Format(longFmt)
	var start int
	// make sure we start at the current time
	for i, t := range res.Data.Times.Values {
		if t == now {
			start = i
			break
		}
	}
	end := start + 24
	if l := len(res.Data.Times.Values); l < end {
		end = l
	}
	f, err := os.Create("/tmp/weather.dat")
	if err != nil {
		panic(err)
	}
	for i, t := range res.Data.Times.Values[start : end+1] {
		tt, _ := time.Parse(longFmt, t)
		fmt.Fprintf(f, "%s %d\n", tt.Format(shortFmt), temps["hourly"][i])
	}
	cmd := exec.Command("gnuplot", "--persist")
	// set xrange ["11/24/21":"12/01/21"]
	cmd.Stdin = strings.NewReader(`set terminal png medium size 640,480 font arial 12
set output 'out.png'
set bmargin 2.5
unset key
set xdata time
set timefmt "%m-%d/%H:%M"
set ylabel "Temperature (°F)"
plot "/tmp/weather.dat" using 1:2 with linespoints
`)
	cmd.Run()
}

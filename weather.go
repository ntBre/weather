package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	url      = "https://forecast.weather.gov/MapClick.php?lat=34.3651&lon=-89.5196&FcstType=digitalDWML"
	longFmt  = "2006-01-02T15:04:05-07:00"
	shortFmt = "01-02/15:00"
	badInt = -999
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
	for _, ts := range res.Data.Parameters.Temps {
		// this allows me to check for bad values, such as
		// missing wind chills
		for _, t := range ts.Value {
			v, err := strconv.Atoi(t)
			if err != nil {
				v = badInt
			}
			temps[ts.Type] = append(temps[ts.Type], v)
		}
	}
	winds := make(map[string][]int)
	for _, t := range res.Data.Parameters.Winds {
		winds[t.Type] = t.Value
	}
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
	tmax := temps["hourly"][start]
	tmin := tmax
	for i, t := range res.Data.Times.Values[start : end+1] {
		tt, _ := time.Parse(longFmt, t)
		h := temps["hourly"][i]
		d := temps["dew point"][i]
		w := temps["wind chill"][i]
		tmax = max(h, d, tmax)
		tmin = min(h, d, tmin)
		fmt.Fprintf(f, "%s %d %d %d\n", tt.Format(shortFmt),
			h, d, w,
		)
	}
	cmd := exec.Command("gnuplot", "--persist")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf(`set terminal png medium size 640,480 font arial 12
set output 'out.png'
set bmargin 2.5
set xdata time
set timefmt "%%m-%%d/%%H:%%M"
set ylabel "Temperature (°F)"
set yrange [%d:%d]
plot "/tmp/weather.dat" u 1:2 w linespoints lc rgb "red" title "Hourly", \
"/tmp/weather.dat" u 1:3 w linespoints lc rgb "green" title "Dew Point", \
"/tmp/weather.dat" u 1:($4 == %d ? NaN : $4) w linespoints lc rgb "blue" title "Wind Chill"
`, tmin-5, tmax+10, badInt))
	cmd.Run()
}

func max(ds ...int) int {
	max := ds[0]
	if len(ds) < 2 {
		return max
	}
	for _, d := range ds[1:] {
		if d > max {
			max = d
		}
	}
	return max
}

func min(ds ...int) int {
	min := ds[0]
	if len(ds) < 2 {
		return min
	}
	for _, d := range ds[1:] {
		if d < min && d != -999 {
			min = d
		}
	}
	return min
}

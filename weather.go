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
	badInt   = -999
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

func GetWeather() *Result {
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
	return res
}

func main() {
	res := GetWeather()
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
	hmax, hmin := tmax, tmax
	for i, t := range res.Data.Times.Values[start : end+1] {
		tt, _ := time.Parse(longFmt, t)
		h := temps["hourly"][i]
		d := temps["dew point"][i]
		w := temps["wind chill"][i]
		r := res.Data.Parameters.Rain[i]
		tmax = max(h, d, tmax)
		tmin = min(h, d, tmin)
		hmax = max(h, hmax)
		hmin = min(h, hmin)
		fmt.Fprintf(f, "%s %d %d %d %d\n", tt.Format(shortFmt),
			h, d, w, r,
		)
	}

	t1, _ := time.Parse(longFmt, res.Data.Times.Values[start])
	t2, _ := time.Parse(longFmt, res.Data.Times.Values[end])
	cmd := exec.Command("gnuplot", "--persist")
	cmd.Stdin = strings.NewReader(
		fmt.Sprintf(`set terminal pngcairo size 840,520 font "arial,12"
set output '/tmp/forecast.png'
fmin(x) = %d
fmax(x) = %d
set bmargin 2.5
set lmargin 8.0
set xdata time
set timefmt "%%m-%%d/%%H:%%M"
set xrange ["%s":"%s"]
set ylabel "Temperature (°F)"
set ytics nomirror
set y2label "Rain chance (%%)" rotate by 270
set y2tics 10
set yrange [%d:%d]
set y2range [0:105]
set mytics 5
plot fmax(x) dt 2 lc rgb "#ff8c00" title "", fmin(x) dt 2 lc rgb "#00bfff" title "", \
"/tmp/weather.dat" u 1:2 w linespoints lc rgb "red" title "Hourly", \
"/tmp/weather.dat" u 1:3 w linespoints lc rgb "green" title "Dew Point", \
"/tmp/weather.dat" u 1:($4 == %d ? NaN : $4) w linespoints lc rgb "blue" title "Wind Chill", \
"/tmp/weather.dat" u 1:5 w boxes lc rgb "#00bfff" title "Rain" axes x1y2
`,
			hmin, hmax,
			t1.Format(shortFmt), t2.Format(shortFmt),
			tmin-5, tmax+10, badInt))
	cmd.Stderr = os.Stderr
	cmd.Run()
	exec.Command("xwallpaper", "--center", "/tmp/forecast.png").Run()
}

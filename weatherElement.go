package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/donniet/mirror2/wunderground"
)

type weatherElement struct {
	visible    bool
	high       float64
	low        float64
	icon       string
	date       time.Time
	weatherURL string
	client     *http.Client
	changed    chan bool
	ticker     *time.Ticker
	err        error
	lock       *sync.Mutex
}

func newWeatherElement(weatherURL string, changed chan bool, frequency time.Duration) (e *weatherElement) {
	e = &weatherElement{
		visible:    false,
		high:       -100,
		low:        -100,
		icon:       "Sun",
		date:       time.Now(),
		weatherURL: weatherURL,
		changed:    changed,
		ticker:     time.NewTicker(frequency),
		lock:       &sync.Mutex{},
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
	go e.fetchWeatherThread()
	return e
}

func (e *weatherElement) ServeJSON(path []string, msg *json.RawMessage) (*json.RawMessage, error) {
	if len(path) == 0 {
		b, err := json.Marshal(e)
		return (*json.RawMessage)(&b), err
	}

	if len(path) > 1 {
		return nil, &NotFoundError{Path: path}
	}

	var v interface{}

	e.lock.Lock()
	defer e.lock.Unlock()

	switch path[0] {
	case "visible":
		v = e.visible
	case "high":
		v = e.high
	case "low":
		v = e.low
	case "icon":
		v = e.icon
	case "date":
		v = e.date
	default:
		return nil, &NotFoundError{Path: path}
	}

	b, err := json.Marshal(v)
	return (*json.RawMessage)(&b), err
}

func (e *weatherElement) UnmarshalJSON(b []byte) error {
	m := make(map[string]interface{})

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	if v, ok := m["visible"]; ok {
		vis := false
		if vis, ok = v.(bool); !ok {
			return fmt.Errorf("weather visible must be a boolean")
		}

		if vis && !e.Visible() {
			e.Show()
		} else if !vis && e.Visible() {
			e.Hide()
		}
	}
	return nil
}

func (e *weatherElement) MarshalJSON() ([]byte, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	r := map[string]interface{}{
		"high":    e.high,
		"low":     e.low,
		"icon":    e.icon,
		"visible": e.visible,
		"date":    e.date,
	}
	if e.err != nil {
		r["error"] = e.err.Error()
	}
	return json.Marshal(r)
}

func (e *weatherElement) fetchWeatherThread() {
	if e.err = e.fetchWeather(); e.err != nil {
		log.Printf("error fetching weather: %v", e.err)
	}

	for _ = range e.ticker.C {
		if e.err = e.fetchWeather(); e.err != nil {
			log.Printf("error fetching weather: %v", e.err)
		}
	}
}

func (e *weatherElement) fetchWeather() error {
	var f wunderground.ForecastResponse

	if res, err := e.client.Get(e.weatherURL); err != nil {
		return err
	} else if b, err := ioutil.ReadAll(res.Body); err != nil {
		return err
	} else if err = json.Unmarshal(b, &f); err != nil {
		return err
	} else if f.Forecast == nil ||
		f.Forecast.SimpleForecast == nil ||
		len(f.Forecast.SimpleForecast.ForecastDay) == 0 {
		return fmt.Errorf("weather response does not contain a valid forecast: %#v", res)
	} else {
		forecastDay := f.Forecast.SimpleForecast.ForecastDay[0]

		e.lock.Lock()

		if e.high, err = strconv.ParseFloat(forecastDay.High.Fahrenheit, 64); err != nil {
			e.high = -100
		}
		if e.low, err = strconv.ParseFloat(forecastDay.Low.Fahrenheit, 64); err != nil {
			e.low = -100
		}
		e.icon = iconMap[forecastDay.Icon]
		e.date = forecastDay.Date()
		e.lock.Unlock()
	}
	e.changed <- true
	return nil
}

func (e *weatherElement) Visible() bool {
	return e.visible
}

func (e *weatherElement) Name() string {
	return "Weather"
}

func (e *weatherElement) Show() {
	e.visible = true
	e.changed <- true
}

func (e *weatherElement) Hide() {
	e.visible = false
	e.changed <- true
}

func (e *weatherElement) High() float64 {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.high
}

func (e *weatherElement) Low() float64 {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.low
}

func (e *weatherElement) Icon() string {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.icon
}

var iconMap = map[string]string{
	"chanceflurries":    "Cloud-Snow-Sun-Alt",
	"chancerain":        "Cloud-Rain-Sun-Alt",
	"chancesleet":       "Cloud-Hail-Sun",
	"chancesnow":        "Cloud-Snow-Sun-Alt",
	"chancetstorms":     "Cloud-Lightning-Sun",
	"clear":             "Sun",
	"cloudy":            "Cloud",
	"flurries":          "Cloud-Snow",
	"fog":               "Cloud-Fog",
	"hazy":              "Cloud-Fog-Sun",
	"mostlycloudy":      "Cloud-Sun",
	"nt_chanceflurries": "Cloud-Snow-Moon-Alt",
	"nt_chancerain":     "Cloud-Rain-Moon-Alt",
	"nt_chancesleet":    "Cloud-Hail-Moon",
	"nt_chancesnow":     "Cloud-Snow-Moon-Alt",
	"nt_chancetstorms":  "Cloud-Lightning-Moon",
	"nt_clear":          "Moon",
	"nt_cloudy":         "Cloud-Moon",
	"nt_flurries":       "Cloud-Snow",
	"nt_fog":            "Cloud-Fog",
	"nt_hazy":           "CLoud-Fog-Moon",
	"nt_mostlysunny":    "Cloud-Moon",
	"nt_partlycloudy":   "Cloud-Moon",
	"nt_partlysunny":    "Cloud-Moon",
	"nt_rain":           "Cloud-Rain",
	"nt_sleet":          "Cloud-Hail",
	"nt_snow":           "Cloud-Snow",
	"nt_sunny":          "Cloud-Moon",
	"nt_tstorms":        "Cloud-Lightning",
	"partlycloudy":      "Cloud-Sun",
	"partlysunny":       "Cloud-Sun",
	"rain":              "Cloud-Rain",
	"sleet":             "Cloud-Hail",
	"snow":              "Cloud-Snow",
	"sunny":             "Sun",
	"tstorms":           "Cloud-Lightning",
}

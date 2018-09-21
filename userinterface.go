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

func NewMirrorInterface(weatherURL string) *mirrorInterface {
	c := make(chan UIElement)
	cec, _ := NewCECDisplay("", "Smart Mirror")
	return &mirrorInterface{
		changed: c,
		weather: newWeatherElement(weatherURL, c, time.Hour),
		display: cec,
		date: &dateTimeElement{
			visible: false,
			changed: c,
		},
	}
}

type mirrorInterface struct {
	changed chan<- UIElement
	weather *weatherElement
	date    *dateTimeElement
	display *CECDisplay
}

func (ui *mirrorInterface) Streams() []UIElement {
	return []UIElement{}
}

func (ui *mirrorInterface) Weather() WeatherElement {
	return ui.weather
}

func (ui *mirrorInterface) DateTime() UIElement {
	return ui.date
}

func (ui *mirrorInterface) Video() UIElement {
	return nil
}

func (ui *mirrorInterface) Display() Display {
	return ui.display
}

type dateTimeElement struct {
	visible bool
	changed chan<- UIElement
}

func (e *dateTimeElement) Visible() bool {
	return e.visible
}

func (e *dateTimeElement) Name() string {
	return "Date and Time"
}

func (e *dateTimeElement) Show() {
	e.visible = true
	e.changed <- e
}

func (e *dateTimeElement) Hide() {
	e.visible = false
	e.changed <- e
}

func (e *dateTimeElement) HasError() error {
	return nil
}

type weatherElement struct {
	visible    bool
	high       float64
	low        float64
	icon       string
	date       time.Time
	weatherURL string
	client     *http.Client
	changed    chan<- UIElement
	ticker     *time.Ticker
	err        error
	lock       *sync.Mutex
}

func newWeatherElement(weatherURL string, changed chan<- UIElement, frequency time.Duration) (e *weatherElement) {
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

func (e *weatherElement) UnmarshalJSON(b []byte) error {
	return nil
}

func (e *weatherElement) MarshalJSON() ([]byte, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	return json.Marshal(map[string]interface{}{
		"high": e.high,
		"low":  e.low,
		"icon": e.icon,
	})
}

func (e *weatherElement) HasError() error {
	return e.err
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
	e.changed <- e
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
	e.changed <- e
}

func (e *weatherElement) Hide() {
	e.visible = false
	e.changed <- e
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

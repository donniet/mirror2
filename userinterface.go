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
		streams: make(map[string]*streamElement),
	}
}

type streamElement struct {
	url     string
	name    string
	visible bool
	changed chan<- UIElement
}

func (e *streamElement) UnmarshalJSON([]byte) error {
	return nil
}

func (e *streamElement) MarshalJSON() ([]byte, error) {
	ret := make(map[string]interface{})
	ret["url"] = e.url
	ret["name"] = e.name
	ret["visible"] = e.visible
	return json.Marshal(ret)
}

func (e *streamElement) URL() string {
	return e.url
}
func (e *streamElement) Visible() bool {
	return e.visible
}
func (e *streamElement) Name() string {
	return e.name
}
func (e *streamElement) Show() {
	if !e.visible {
		e.visible = true
		e.changed <- e
	}
}
func (e *streamElement) Hide() {
	if e.visible {
		e.visible = false
		e.changed <- e
	}
}
func (e *streamElement) HasError() error {
	return nil
}

type mirrorInterface struct {
	changed chan<- UIElement
	weather *weatherElement
	date    *dateTimeElement
	display *CECDisplay
	streams map[string]*streamElement
	video   *videoElement
}

func (ui *mirrorInterface) UnmarshalJSON([]byte) error {
	return nil
}
func (ui *mirrorInterface) MarshalJSON() ([]byte, error) {
	ret := make(map[string]interface{})
	ret["streams"] = ui.Streams()
	ret["weather"] = ui.Weather()
	ret["dateTime"] = ui.DateTime()
	ret["video"] = ui.Video()
	ret["display"] = ui.Display()
	return json.Marshal(ret)
}

func (ui *mirrorInterface) Streams() (ret []StreamElement) {
	for _, e := range ui.streams {
		ret = append(ret, e)
	}
	return
}

func (ui *mirrorInterface) Weather() WeatherElement {
	return ui.weather
}

func (ui *mirrorInterface) DateTime() UIElement {
	return ui.date
}

func (ui *mirrorInterface) Video() VideoService {
	return ui.video
}

func (ui *mirrorInterface) Display() Display {
	return ui.display
}

func (ui *mirrorInterface) AddStream(name string, url string) {
	if s, ok := ui.streams[url]; ok {
		s.name = name
	} else {
		ui.streams[url] = &streamElement{
			name:    name,
			url:     url,
			visible: false,
		}
	}
}

func (ui *mirrorInterface) RemoveStream(url string) {
	delete(ui.streams, url)
}

type dateTimeElement struct {
	visible bool
	changed chan<- UIElement
}

func (e *dateTimeElement) UnmarshalJSON(b []byte) error {
	return nil
}
func (e *dateTimeElement) MarshalJSON() ([]byte, error) {
	ret := make(map[string]interface{})
	ret["visible"] = e.visible
	return json.Marshal(ret)
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
	r := map[string]interface{}{
		"high": e.high,
		"low":  e.low,
		"icon": e.icon,
	}
	if e.err != nil {
		r["error"] = e.err.Error()
	}
	return json.Marshal(r)
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

type videoStateCommand string

const (
	Play  videoStateCommand = "play"
	Pause                   = "pause"
	Stop                    = "stop"
)

type videoCommand struct {
	Mute    *bool              `json:"mute,omitempty"`
	Volume  *int               `json:"volume,omitempty"`
	Load    string             `json:"load,omitempty"`
	Command *videoStateCommand `json:"command,omitempty"`
}

type videoElement struct {
	visible  bool
	muted    bool
	volume   int
	state    VideoState
	commands chan<- videoCommand
	receiver <-chan VideoState
	lock     *sync.Mutex
}

func newVideoElement(commands chan<- videoCommand, receiver <-chan VideoState) *videoElement {
	ret := &videoElement{
		volume:   100,
		state:    Unstarted,
		commands: commands,
		receiver: receiver,
		lock:     &sync.Mutex{},
	}
	go ret.receiverState()
	return ret
}

func (v *videoElement) receiverState() {
	for s := range v.receiver {
		v.lock.Lock()
		v.state = s
		v.lock.Unlock()
	}
}

func (v *videoElement) LoadVideoByID(videoID string)   {}
func (v *videoElement) LoadVideoByURL(videoURL string) {}
func (v *videoElement) Play()                          {}
func (v *videoElement) Pause()                         {}
func (v *videoElement) Stop()                          {}
func (v *videoElement) SeekTo(seconds float32)         {}
func (v *videoElement) Mute()                          {}
func (v *videoElement) UnMute()                        {}
func (v *videoElement) IsMuted() bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.muted
}
func (v *videoElement) SetVolume(volume int) {}
func (v *videoElement) Volume() int {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.volume
}
func (v *videoElement) State() VideoState {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.state
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

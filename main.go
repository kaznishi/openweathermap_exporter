package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr = flag.String("listen-address", ":9999", "The address to listen on for HTTP requests.")
)

const (
	namespace = "openweathermap"
)

type weatherData struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Pressure float64 `json:"pressure"`
		Humidity float64 `json:"humidity"`
	}
}

type openWeatherMapCollector struct {
	location string
	apiKey   string

	temp     prometheus.Gauge
	pressure prometheus.Gauge
	humidity prometheus.Gauge
}

func (c *openWeatherMapCollector) fetchFromAPI() (weatherData, error) {
	var wd weatherData
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", c.location, c.apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return wd, err
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&wd)
	return wd, err
}

func newOpenWeatherMapCollector(location string, apiKey string) *openWeatherMapCollector {
	return &openWeatherMapCollector{
		location: location,
		apiKey:   apiKey,
		temp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "temperature_celsius",
			Help:      "Temperature in °C",
		}),
		pressure: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "pressure_hpa",
			Help:      "Atmospheric pressure in hPa",
		}),
		humidity: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "humidity_percent",
			Help:      "Humidity in Percent",
		}),
	}
}

func (c *openWeatherMapCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.temp.Desc()
	ch <- c.pressure.Desc()
	ch <- c.humidity.Desc()
}

func (c *openWeatherMapCollector) Collect(ch chan<- prometheus.Metric) {
	wd, err := c.fetchFromAPI()
	if err != nil {
		log.Printf("%v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.temp.Desc(),
		prometheus.GaugeValue,
		wd.Main.Temp,
	)
	ch <- prometheus.MustNewConstMetric(
		c.pressure.Desc(),
		prometheus.GaugeValue,
		wd.Main.Pressure,
	)
	ch <- prometheus.MustNewConstMetric(
		c.humidity.Desc(),
		prometheus.GaugeValue,
		wd.Main.Humidity,
	)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("api_key")
	if apiKey == "" {
		http.Error(w, fmt.Sprintf("api_key is not specified."), http.StatusBadRequest)
		return
	}
	location := r.URL.Query().Get("location")
	if location == "" {
		http.Error(w, fmt.Sprintf("location is not specified."), http.StatusBadRequest)
		return
	}

	c := newOpenWeatherMapCollector(location, apiKey)
	registry := prometheus.NewRegistry()
	registry.MustRegister(c)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r)
	})
	log.Fatal(http.ListenAndServe(*addr, nil))
}

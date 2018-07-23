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
	addr     = flag.String("listen-address", ":9999", "The address to listen on for HTTP requests.")
	location = flag.String("location", "Tokyo", "The city name which you want to get data of")
	apiKey   = flag.String("apiKey", "", "Your Key of OpenWeatherMap API")
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

func main() {
	flag.Parse()

	c := newOpenWeatherMapCollector(*location, *apiKey)
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}

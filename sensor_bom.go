package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorBOM struct {
	name        string
	scanner     *BOMScanner
	temperature prometheus.Gauge
	windSpeed   prometheus.Gauge
	windGust    prometheus.Gauge
	windDir     prometheus.Gauge
}

func NewSensorBOM(name string, debug bool) (s *SensorBOM, err error) {
	s = &SensorBOM{
		name: name,
	}

	if s.scanner, err = NewBOMScanner(debug); err != nil {
		return
	}

	s.temperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_temperature",
		Help: "The current temperature from the BOM",
	})

	s.windSpeed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_wind_speed",
		Help: "The current wind speed from the BOM",
	})

	s.windGust = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_wind_gust",
		Help: "The current wind gust speed from the BOM",
	})

	s.windDir = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_wind_dir",
		Help: "The current wind direction from the BOM",
	})

	return
}

func (s *SensorBOM) Update() {
	s.temperature.Set(s.scanner.AirTemp)
	s.windSpeed.Set(s.scanner.WindSpeed)
	s.windGust.Set(s.scanner.WindGust)
	s.windDir.Set(s.scanner.WindDir)
}

func (s *SensorBOM) Summary() string {
	return fmt.Sprintf("%s: %.2f C, %.0f-%.0f kph from %.1f", s.name, s.scanner.AirTemp, s.scanner.WindSpeed, s.scanner.WindGust, s.scanner.WindDir)
}

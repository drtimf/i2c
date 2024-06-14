package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorBME280 struct {
	name            string
	bme280          *BME280
	temperature     float64
	pressure        float64
	humidity        float64
	promTemperature prometheus.Gauge
	promPressure    prometheus.Gauge
	promHumidity    prometheus.Gauge
}

func NewSensorBME280(name string, i2cAddress uint8) (s *SensorBME280, err error) {
	s = &SensorBME280{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = 0x77
	}

	if s.bme280, err = NewBME280(i2cAddress); err != nil {
		return
	}

	s.promTemperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_temperature",
		Help: "Temperature from a BME280 sensor",
	})

	s.promPressure = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_pressure",
		Help: "The current pressure from the BME280",
	})

	s.promHumidity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_humidity",
		Help: "The current humidity from the BME280",
	})

	return
}

func (s *SensorBME280) Update() {
	if t, p, h, err := s.bme280.Read(); err != nil {
		fmt.Printf("ERROR: Failed to read temperature, pressure and humidity from BME280 \"%s\": %v\n", s.name, err)
	} else {
		if t < 100 && t > -100 {
			s.temperature = t
			s.pressure = p
			s.humidity = h
		}
	}

	s.promTemperature.Set(s.temperature)
	s.promPressure.Set(s.pressure)
	s.promHumidity.Set(s.humidity)
}

func (s *SensorBME280) Summary() string {
	return fmt.Sprintf("%s: %.2f C, %.2f hPa, %.2f rH", s.name, s.temperature, s.pressure, s.humidity)
}

func (s *SensorBME280) Details() string {
	return fmt.Sprintf("%s - BME280 temperature, pressume and humidity sensor: %.2f C, %.2f hPa, %.2f rH", s.name, s.temperature, s.pressure, s.humidity)
}

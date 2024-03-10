package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorAHT10 struct {
	name            string
	aht10           *piicodev.AHT10
	temperature     float64
	humidity        float64
	promTemperature prometheus.Gauge
	promHumidity    prometheus.Gauge
}

func NewSensorAHT10(name string, i2cAddress uint8) (s *SensorAHT10, err error) {
	s = &SensorAHT10{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.AHT10Address
	}

	if s.aht10, err = piicodev.NewAHT10(i2cAddress, 1); err != nil {
		return
	}

	s.promTemperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_temperature",
		Help: "Temperature from a AHT10 sensor",
	})

	s.promHumidity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_humidity",
		Help: "Humidity from a AHT10 sensor",
	})
	return
}

func (s *SensorAHT10) Update() {
	if temperature, humidity, err := s.aht10.ReadSensor(); err != nil {
		fmt.Printf("ERROR: Failed to read temperature and humidity from AHT10 \"%s\": %v\n", s.name, err)
	} else {
		if temperature < 100 && temperature > -100 {
			s.temperature = temperature
			s.humidity = humidity
		}
	}

	s.promTemperature.Set(s.temperature)
	s.promHumidity.Set(s.humidity)
}

func (s *SensorAHT10) Summary() string {
	return fmt.Sprintf("%s: %.2f C, %.2f rH", s.name, s.temperature, s.humidity)
}

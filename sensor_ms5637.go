package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorMS5637 struct {
	name            string
	ms5637          *piicodev.MS5637
	pressure        float64
	temperature     float64
	promPressure    prometheus.Gauge
	promTemperature prometheus.Gauge
}

func NewSensorMS5637(name string, i2cAddress uint8) (s *SensorMS5637, err error) {
	s = &SensorMS5637{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.MS5637Address
	}

	if s.ms5637, err = piicodev.NewMS5637(i2cAddress, 1); err != nil {
		return
	}

	s.promPressure = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_pressure",
		Help: "Pressure from a MS5637 sensor",
	})

	s.promTemperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_temperature",
		Help: "Temperature from a MS5637 sensor",
	})

	return
}

func (s *SensorMS5637) Update() {
	if p, t, err := s.ms5637.Read(); err != nil {
		fmt.Printf("ERROR: Failed to read pressure and temperature from MS5637 \"%s\": %v\n", s.name, err)
	} else {
		if t < 100 && t > -100 {
			s.pressure = p
			s.temperature = t
		}
	}

	s.promPressure.Set(s.pressure)
	s.promTemperature.Set(s.temperature)
}

func (s *SensorMS5637) Summary() string {
	return fmt.Sprintf("%s: %.2f hPa, %.2f C", s.name, s.pressure, s.temperature)
}

func (s *SensorMS5637) Details() string {
	return fmt.Sprintf("%s - MS5637 Pressure Sensor: %.2f hPa, %.2f C", s.name, s.pressure, s.temperature)
}

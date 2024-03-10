package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorTMP117 struct {
	name            string
	tmp117          *piicodev.TMP117
	temperature     float64
	promTemperature prometheus.Gauge
}

func NewSensorTMP117(name string, i2cAddress uint8) (s *SensorTMP117, err error) {
	s = &SensorTMP117{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.TMP117Address
	}

	if s.tmp117, err = piicodev.NewTMP117(i2cAddress, 1); err != nil {
		return
	}

	s.promTemperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_temperature",
		Help: "Temperature from a TMP117 sensor",
	})

	return
}

func (s *SensorTMP117) Update() {
	if t, err := s.tmp117.ReadTempC(); err != nil {
		fmt.Printf("ERROR: Failed to read temperature from TMP117 \"%s\": %v\n", s.name, err)
	} else {
		if t < 100 && t > -100 {
			s.temperature = t
		}
	}

	s.promTemperature.Set(s.temperature)
}

func (s *SensorTMP117) Summary() string {
	return fmt.Sprintf("%s: %.2f C", s.name, s.temperature)
}

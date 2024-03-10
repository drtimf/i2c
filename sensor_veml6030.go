package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorVEML6030 struct {
	name           string
	veml6030       *piicodev.VEML6030
	lightLevel     float64
	promLightLevel prometheus.Gauge
}

func NewSensorVEML6030(name string, i2cAddress uint8) (s *SensorVEML6030, err error) {
	s = &SensorVEML6030{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.VEML6030Address
	}

	if s.veml6030, err = piicodev.NewVEML6030(i2cAddress, 1); err != nil {
		return
	}

	s.promLightLevel = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_light_level",
		Help: "Light level from a VEML6030 sensor",
	})

	return
}

func (s *SensorVEML6030) Update() {
	if l, err := s.veml6030.Read(); err != nil {
		fmt.Printf("ERROR: Failed to read light level from VEML6040 \"%s\": %v\n", s.name, err)
	} else {
		s.lightLevel = l
	}

	s.promLightLevel.Set(s.lightLevel)
}

func (s *SensorVEML6030) Summary() string {
	return fmt.Sprintf("%s: %.2f lux", s.name, s.lightLevel)
}

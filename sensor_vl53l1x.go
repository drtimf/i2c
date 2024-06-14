package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorVL53L1X struct {
	name         string
	vl53l1x      *piicodev.VL53L1X
	distance     uint16
	promDistance prometheus.Gauge
}

func NewSensorVL53L1X(name string, i2cAddress uint8) (s *SensorVL53L1X, err error) {
	s = &SensorVL53L1X{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.VL53L1XAddress
	}

	if s.vl53l1x, err = piicodev.NewVL53L1X(i2cAddress, 1); err != nil {
		return
	}

	s.promDistance = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_distance",
		Help: "The current distance from a VL53L1X sensor",
	})

	return
}

func (s *SensorVL53L1X) Update() {
	if d, err := s.vl53l1x.Read(); err != nil {
		fmt.Printf("ERROR: Failed to read distance from VL53L1X \"%s\": %v\n", s.name, err)
	} else {
		s.distance = d
	}

	s.promDistance.Set(float64(s.distance))
}

func (s *SensorVL53L1X) Summary() string {
	return fmt.Sprintf("%s: %d mm", s.name, s.distance)
}

func (s *SensorVL53L1X) Details() string {
	return fmt.Sprintf("%s - VL53L1X Distance Sensor: %d mm", s.name, s.distance)
}

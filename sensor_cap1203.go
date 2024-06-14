package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
)

type SensorCAP1203 struct {
	name    string
	cap1203 *piicodev.CAP1203
	status  [3]bool
}

func NewSensorCAP1203(name string, i2cAddress uint8) (s *SensorCAP1203, err error) {
	s = &SensorCAP1203{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.CAP1203Address
	}

	if s.cap1203, err = piicodev.NewCAP1203(i2cAddress, 1); err != nil {
		return
	}

	s.cap1203.SetSensitivity(5)
	return
}

func (s *SensorCAP1203) Update() {
	var err error
	if s.status[0], s.status[1], s.status[2], err = s.cap1203.Read(); err != nil {
		fmt.Printf("ERROR: Failed to read capacitive status from CAP1203 \"%s\": %v\n", s.name, err)
	}
}

func (s *SensorCAP1203) Summary() string {
	return fmt.Sprintf("%s: %t,%t,%t", s.name, s.status[0], s.status[1], s.status[2])
}

func (s *SensorCAP1203) Details() string {
	return fmt.Sprintf("%s - CAP1203 Capacitive Touch Sensor: %t,%t,%t", s.name, s.status[0], s.status[1], s.status[2])
}

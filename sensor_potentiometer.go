package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
)

type SensorPotentiometer struct {
	name    string
	pot     *piicodev.Potentiometer
	value   uint16
	changed bool
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func NewSensorPotentiometer(name string, i2cAddress uint8) (s *SensorPotentiometer, err error) {
	s = &SensorPotentiometer{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.PotentiometerAddress
	}

	if s.pot, err = piicodev.NewPotentiometer(i2cAddress, 1); err != nil {
		return
	}

	s.value, err = s.pot.ReadRawValue()

	return
}

func (s *SensorPotentiometer) Update() {
	var err error
	var newValue uint16
	if newValue, err = s.pot.ReadRawValue(); err != nil {
		fmt.Printf("ERROR: Failed to read raw value from Potentiometer \"%s\": %v\n", s.name, err)
		return
	}

	s.changed = false
	if abs(int(newValue)-int(s.value)) > 5 {
		s.changed = true
		s.value = newValue
	}
}

func (s *SensorPotentiometer) Summary() string {
	return fmt.Sprintf("%s: %t,%d", s.name, s.changed, s.value)
}

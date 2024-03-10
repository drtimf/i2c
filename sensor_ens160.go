package main

import (
	"fmt"

	"github.com/drtimf/go-piicodev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorENS160 struct {
	name       string
	ens160     *piicodev.ENS160
	operation  string
	aqi        byte
	aqiRating  string
	tvoc       uint16
	eco2       uint16
	eco2Rating string

	promAQI  prometheus.Gauge
	promTVOC prometheus.Gauge
	promECO2 prometheus.Gauge
}

func NewSensorENS160(name string, i2cAddress uint8) (s *SensorENS160, err error) {
	s = &SensorENS160{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.ENS160Address
	}

	if s.ens160, err = piicodev.NewENS160(i2cAddress, 1); err != nil {
		return
	}

	s.promAQI = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_aqi",
		Help: "Air quality index from a ENS160 sensor",
	})

	s.promTVOC = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_tvoc",
		Help: "True volatile organic compounds from a ENS160 sensor",
	})

	s.promECO2 = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_eco2",
		Help: "CO2-equivalents from a ENS160 sensor",
	})
	return
}

func (s *SensorENS160) Update() {
	if operation, err := s.ens160.GetOperation(); err != nil {
		fmt.Printf("ERROR: Failed to read the operation from ENS160 \"%s\": %v\n", s.name, err)
	} else {
		s.operation = operation
	}

	if aqi, aqiRating, err := s.ens160.ReadAQI(); err != nil {
		fmt.Printf("ERROR: Failed to read the AQI from ENS160 \"%s\": %v\n", s.name, err)
	} else {
		s.aqi = aqi
		s.aqiRating = aqiRating
	}

	if tvoc, err := s.ens160.ReadTVOC(); err != nil {
		fmt.Printf("ERROR: Failed to read the TVOC from ENS160 \"%s\": %v\n", s.name, err)
	} else {
		s.tvoc = tvoc
	}

	if eco2, eco2Rating, err := s.ens160.ReadECO2(); err != nil {
		fmt.Printf("ERROR: Failed to read the ECO2 from ENS160 \"%s\": %v\n", s.name, err)
	} else {
		s.eco2 = eco2
		s.eco2Rating = eco2Rating
	}

	s.promAQI.Set(float64(s.aqi))
	s.promTVOC.Set(float64(s.tvoc))
	s.promECO2.Set(float64(s.eco2))
}

func (s *SensorENS160) Summary() string {
	return fmt.Sprintf("%s: [%s], %d (%s) AQI, %d TVOC, %d ppm (%s) eCO2", s.name, s.operation, s.aqi, s.aqiRating, s.tvoc, s.eco2, s.eco2Rating)
}

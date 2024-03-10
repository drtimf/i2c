package main

import (
	"fmt"
	"time"

	"github.com/drtimf/go-piicodev"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorPIR struct {
	name              string
	movementTriggered bool
	movement          bool
	promMovement      prometheus.Counter
}

func NewSensorPIR(name string, i2cAddress uint8) (s *SensorPIR, err error) {
	s = &SensorPIR{
		name: name,
	}

	if i2cAddress == 0 {
		i2cAddress = piicodev.QwiicPIRAddress
	}

	s.promMovement = promauto.NewCounter(prometheus.CounterOpts{
		Name: name + "_movement_count",
		Help: "Number of movement events recorded by the PIR sensor",
	})

	go func() {
		for {
			var pir *piicodev.QwiicPIR = nil

			if pir, err = piicodev.NewQwiicPIR(i2cAddress, 1); err != nil {
				fmt.Printf("ERROR: Failed to open PIR sensor \"%s\": %v\n", s.name, err)
			} else {
				var nv, s1, s2, s3 bool

				for {
					if nv, err = pir.GetRawReading(); err != nil {
						fmt.Printf("ERROR: Failed to read movement from PIR sensor \"%s\": %v\n", s.name, err)
						break
					} else {
						s1 = s2
						s2 = s3
						s3 = nv

						if s1 && s2 && s3 {
							s.movementTriggered = true
						}
					}

					time.Sleep(250 * time.Millisecond)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return
}

func (s *SensorPIR) Update() {
	s.movement = s.movementTriggered
	s.movementTriggered = false
	if s.movement {
		s.promMovement.Add(1)
	}
}

func (s *SensorPIR) Summary() string {
	return fmt.Sprintf("%s: %t", s.name, s.movement)
}

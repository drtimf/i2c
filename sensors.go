package main

type Sensor interface {
	Update()
	Summary() string
}

type SensorManagement struct {
	sensors []Sensor
}

func NewSensorManagement() (sm *SensorManagement) {
	sm = &SensorManagement{
		sensors: make([]Sensor, 0),
	}

	return
}

func (sm *SensorManagement) AddSensor(s Sensor) {
	sm.sensors = append(sm.sensors, s)
}

func (sm *SensorManagement) UpdateSensors() {
	for _, s := range sm.sensors {
		s.Update()
	}
}

func (sm *SensorManagement) Summary() (summary string) {
	summary = ""

	for i, s := range sm.sensors {
		if i > 0 {
			summary += " | "
		}
		summary += s.Summary()
	}

	return
}

func (sm *SensorManagement) GetTemperature() (temperature float64, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorTMP117:
			ok = true
			temperature = ts.temperature
			return
		case *SensorAHT10:
			ok = true
			temperature = ts.temperature
			return
		case *SensorBME280:
			ok = true
			temperature = ts.temperature
			return
		case *SensorBOM:
			ok = true
			temperature = ts.scanner.AirTemp
			return
		}
	}
	return
}

func (sm *SensorManagement) GetLightLevel() (lightLevel float64, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorVEML6030:
			ok = true
			lightLevel = ts.lightLevel
			return
		}
	}
	return
}

func (sm *SensorManagement) GetPressure() (pressure float64, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorMS5637:
			ok = true
			pressure = ts.pressure
			return
		case *SensorBME280:
			ok = true
			pressure = ts.pressure
			return
		}
	}
	return
}

func (sm *SensorManagement) GetHumidity() (humidity float64, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorBME280:
			ok = true
			humidity = ts.humidity
			return
		case *SensorAHT10:
			ok = true
			humidity = ts.humidity
			return
		}
	}
	return
}

func (sm *SensorManagement) GetCapSensorStatus() (status [3]bool, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorCAP1203:
			ok = true
			status = ts.status
			return
		}
	}
	return
}

func (sm *SensorManagement) GetOccupancy() (occupied, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorVL53L1X:
			ok = true
			if ts.distance < 1000 {
				occupied = true
			} else {
				occupied = false
			}
			return
		}
	}
	return
}

func (sm *SensorManagement) GetSwitchPress() (pressType int, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorSwitch:
			ok = true
			pressType = 0
			if ts.wasDoublePressed {
				pressType = 2
			} else if ts.wasPressed {
				pressType = 1
			}
			return
		}
	}
	return
}

func (sm *SensorManagement) GetPotentiometer() (changed bool, value uint16, ok bool) {
	ok = false
	for _, s := range sm.sensors {
		switch ts := s.(type) {
		case *SensorPotentiometer:
			ok = true
			changed = ts.changed
			value = ts.value
			return
		}
	}
	return
}

package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusSensors struct {
	temperature    prometheus.Gauge
	temperatureBME prometheus.Gauge
	temperatureMS  prometheus.Gauge
	lightLevel     prometheus.Gauge
	pressure       prometheus.Gauge
	pressureBME    prometheus.Gauge
	humidity       prometheus.Gauge
	distance       prometheus.Gauge
}

func PrometheusStart() (p *PrometheusSensors) {
	p = &PrometheusSensors{}
	p.temperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_temperature",
		Help: "The current temperature from the TMP117",
	})

	p.temperatureBME = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_temperature_bme",
		Help: "The current temperature from the BME280",
	})

	p.temperatureMS = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_temperature_ms",
		Help: "The current temperature from the MS5637",
	})

	p.lightLevel = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_light_level",
		Help: "The current light level",
	})

	p.pressure = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_pressure",
		Help: "The current pressure from the MS5637",
	})

	p.pressureBME = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_pressure_bme",
		Help: "The current pressure from the BME280",
	})

	p.humidity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_humidity",
		Help: "The current humidity from the BME280",
	})

	p.distance = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_distance",
		Help: "The current distance",
	})

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		http.ListenAndServe(":2112", nil)
	}()

	return
}

func (p *PrometheusSensors) SetTemperature(temp float64) {
	p.temperature.Set(temp)
}

func (p *PrometheusSensors) SetTemperatureBME(temp float64) {
	p.temperatureBME.Set(temp)
}

func (p *PrometheusSensors) SetTemperatureMS(temp float64) {
	p.temperatureMS.Set(temp)
}

func (p *PrometheusSensors) SetLightLevel(lightLevel float64) {
	p.lightLevel.Set(lightLevel)
}

func (p *PrometheusSensors) SetPressure(pressure float64) {
	p.pressure.Set(pressure)
}

func (p *PrometheusSensors) SetPressureBME(pressure float64) {
	p.pressureBME.Set(pressure)
}

func (p *PrometheusSensors) SetHumidity(humidity float64) {
	p.humidity.Set(humidity)
}

func (p *PrometheusSensors) SetRangeSensor(dist uint16) {
	p.distance.Set(float64(dist))
}

package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const DYSON_MESSAGE_CURRENT_STATE = "CURRENT-STATE"
const DYSON_MESSAGE_ENVIRONMENTAL_CURRENT_SENSOR_DATA = "ENVIRONMENTAL-CURRENT-SENSOR-DATA"
const DYSON_MESSAGE_STATE_CHANGE = "STATE-CHANGE"

type DysonMessageEnvironmentalSensorData struct {
	Message string            `json:"msg"`
	Time    time.Time         `json:"time"`
	Data    map[string]string `json:"data"`
}

func (m *DysonMessageEnvironmentalSensorData) GetDataFloat(n string) (f float64) {
	if v, ok := m.Data[n]; ok {
		if fv, err := strconv.ParseFloat(v, 64); err == nil {
			f = fv
		}
	}

	return
}

func (m *DysonMessageEnvironmentalSensorData) GetDataInt(n string) (i int) {
	if v, ok := m.Data[n]; ok {
		if iv, err := strconv.ParseInt(v, 10, 32); err == nil {
			i = int(iv)
		}
	}

	return
}

func (m *DysonMessageEnvironmentalSensorData) Temperature() (temperature float64) {
	return (m.GetDataFloat("tact") / 10.0) - 273.15
}

func (m *DysonMessageEnvironmentalSensorData) Humidity() (temperature int) {
	return m.GetDataInt("hact")
}

// Return particulate matter smaller than 2.5 microns in micro grams per cubic meter.
func (m *DysonMessageEnvironmentalSensorData) ParticulateMatter25() (pm25 int) {
	return m.GetDataInt("pm25")
}

// Return particulate matter smaller than 10 microns in micro grams per cubic meter.
func (m *DysonMessageEnvironmentalSensorData) ParticulateMatter10() (pm10 int) {
	return m.GetDataInt("pm10")
}

// Return VOCs in micro grams per cubic meter.
func (m *DysonMessageEnvironmentalSensorData) VolatileOrganicCompounds() (va10 int) {
	return m.GetDataInt("va10")
}

// Return nitrogen dioxide level in micro grams per cubic meter.
func (m *DysonMessageEnvironmentalSensorData) NitrogenDioxide() (noxl int) {
	return m.GetDataInt("noxl")
}

type SensorDysonHotCool struct {
	name            string
	client          mqtt.Client
	statusTopic     string
	cmdTopic        string
	temperature     float64
	humidity        float64
	pm25            int
	pm10            int
	va10            int
	noxl            int
	promTemperature prometheus.Gauge
	promHumidity    prometheus.Gauge
	promPM25        prometheus.Gauge
	promPM10        prometheus.Gauge
	promVA10        prometheus.Gauge
	promNOXL        prometheus.Gauge
}

func (s *SensorDysonHotCool) onMessageReceived(client mqtt.Client, message mqtt.Message) {
	var err error
	var msg DysonMessageEnvironmentalSensorData

	if err = json.Unmarshal(message.Payload(), &msg); err != nil {
		fmt.Println("ERROR: Bad JSON MQTT payload", err)
		return
	}

	if msg.Message == DYSON_MESSAGE_ENVIRONMENTAL_CURRENT_SENSOR_DATA && msg.Temperature() > 0 {
		s.temperature = msg.Temperature()
		s.humidity = float64(msg.Humidity())
		s.pm25 = msg.ParticulateMatter25()
		s.pm10 = msg.ParticulateMatter10()
		s.va10 = msg.VolatileOrganicCompounds()
		s.noxl = msg.NitrogenDioxide()
	}
}

func (s *SensorDysonHotCool) publishUpdateRequest() {
	s.client.Publish(s.cmdTopic, 0, false,
		fmt.Sprintf(`{ "msg": "REQUEST-PRODUCT-ENVIRONMENT-CURRENT-SENSOR-DATA", "time": "%s" }`,
			time.Now().UTC().Format("2006-01-02T15:04:05.000Z")))
}

func NewSensorDysonHotCool(name string, server string, deviceType string, serial string, password string, debug bool) (s *SensorDysonHotCool, err error) {
	s = &SensorDysonHotCool{
		name:        name,
		statusTopic: fmt.Sprintf("%s/%s/status/current", deviceType, serial),
		cmdTopic:    fmt.Sprintf("%s/%s/command", deviceType, serial),
	}

	hostname, _ := os.Hostname()
	clientid := hostname + strconv.Itoa(time.Now().Second())
	connOpts := mqtt.NewClientOptions().AddBroker(server).SetClientID(clientid).SetCleanSession(true).
		SetUsername(serial).SetPassword(password).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert})

	connOpts.OnConnect = func(c mqtt.Client) {
		if token := c.Subscribe(s.statusTopic, 0, s.onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		time.Sleep(1 * time.Second)
		s.publishUpdateRequest()
	}

	s.client = mqtt.NewClient(connOpts)
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to %s\n", server)
	}

	s.promTemperature = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_temperature",
		Help: "Temperature from a Dyson Hot-Cool sensor",
	})

	s.promHumidity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_humidity",
		Help: "Humidity from a Dyson Hot-Cool sensor",
	})

	s.promPM25 = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_pm25",
		Help: "Particulate Matter 2.5 from a Dyson Hot-Cool sensor",
	})

	s.promPM10 = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_pm10",
		Help: "Particulate Matter 10 from a Dyson Hot-Cool sensor",
	})

	s.promVA10 = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_va10",
		Help: "Volatile Organic Compounds from a Dyson Hot-Cool sensor",
	})

	s.promNOXL = promauto.NewGauge(prometheus.GaugeOpts{
		Name: name + "_noxl",
		Help: "Nitrogen Dioxide from a Dyson Hot-Cool sensor",
	})

	go func() {
		for {
			s.publishUpdateRequest()
			time.Sleep(5 * time.Second)
		}
	}()

	time.Sleep(2 * time.Second)
	return
}

func (s *SensorDysonHotCool) Update() {
	s.promTemperature.Set(s.temperature)
	s.promHumidity.Set(s.humidity)
	s.promPM25.Set(float64(s.pm25))
	s.promPM10.Set(float64(s.pm10))
	s.promVA10.Set(float64(s.va10))
	s.promNOXL.Set(float64(s.noxl))
}

func (s *SensorDysonHotCool) Summary() string {
	return fmt.Sprintf("%s: %.2f C, %.2f rH", s.name, s.temperature, s.humidity)
}

func (s *SensorDysonHotCool) Details() string {
	return fmt.Sprintf("%s - Dyson Hot-Cool temperature and humidity: %.2f C, %.2f rH", s.name, s.temperature, s.humidity)
}

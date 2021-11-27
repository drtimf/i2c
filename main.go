package main

import (
	"fmt"
	"os"
	"time"

	// "i2c/go-piicodev.local"
	"github.com/2tvenom/golifx"
	"github.com/drtimf/go-piicodev"
	"gopkg.in/yaml.v2"
)

type I2cConfiguration struct {
	HomeKitDeviceID  string
	HomeKitDevicePin uint32
	EnableLifx       bool
	EnableTMP117     bool
	EnableVEML6030   bool
	EnableVL53L1X    bool
	EnableMS5637     bool
	EnableBME280     bool
	EnableOLED       bool
}

func LoadI2cConfiguration(fileName string) (cfg *I2cConfiguration, err error) {
	var f *os.File
	if f, err = os.Open(fileName); err != nil {
		return
	}
	defer f.Close()

	cfg = &I2cConfiguration{}
	err = yaml.NewDecoder(f).Decode(&cfg)
	return
}

func PrintState(device string, status bool) {
	if status {
		fmt.Printf("Enabled: %s\n", device)
	} else {
		fmt.Printf("Disabled: %s\n", device)
	}
}

func main() {
	var err error

	var config *I2cConfiguration
	if config, err = LoadI2cConfiguration("config/config.yaml"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("HomeKit bridge device ID:", config.HomeKitDeviceID)
	var hkb *HomeKitBridge
	if hkb, err = HomeKitBridgeStart(config.HomeKitDeviceID, config.HomeKitDevicePin); err != nil {
		fmt.Println(err)
		return
	}

	prom := PrometheusStart()

	var bulb *golifx.Bulb = nil

	PrintState("Lifx bulb", config.EnableLifx)
	if config.EnableLifx {
		var bulbs []*golifx.Bulb
		if bulbs, err = golifx.LookupBulbs(); err == nil && len(bulbs) > 0 {
			bulb = bulbs[0]
			fmt.Println(bulb.String())
		}
	}

	var tmp117 *piicodev.TMP117 = nil
	PrintState("TMP117 temperature sensor", config.EnableLifx)
	if config.EnableTMP117 {
		if tmp117, err = piicodev.NewTMP117(piicodev.TMP117Address, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var veml6030 *piicodev.VEML6030 = nil
	PrintState("VEML6030 light sensor", config.EnableVEML6030)
	if config.EnableVEML6030 {
		if veml6030, err = piicodev.NewVEML6030(piicodev.VEML6030Address, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var vl53l1x *piicodev.VL53L1X = nil
	PrintState("VL53L1X distance sensor", config.EnableVL53L1X)
	if config.EnableVL53L1X {
		if vl53l1x, err = piicodev.NewVL53L1X(piicodev.VL53L1XAddress, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var ms5637 *piicodev.MS5637 = nil
	PrintState("MS5637 pressure sensor", config.EnableMS5637)
	if config.EnableMS5637 {
		if ms5637, err = piicodev.NewMS5637(piicodev.MS5637Address, 1); err != nil {
			fmt.Println(err)
			return
		}
	}

	var bme280 *BME280 = nil
	PrintState("BME280 pressure,temperature and humidity sensor", config.EnableBME280)
	if config.EnableBME280 {
		if bme280, err = NewBME280(); err != nil {
			fmt.Println(err)
			return
		}
	}

	var oled *OLEDDisplay = nil
	PrintState("OLED display", config.EnableOLED)
	if config.EnableOLED {
		if oled, err = NewOLEDDisplay(); err != nil {
			fmt.Println(err)
		}
	}

	for {
		var tempC float64
		if tmp117 != nil {
			if tempC, err = tmp117.ReadTempC(); err != nil {
				fmt.Println(err)
				return
			}
		}

		var light float64
		if veml6030 != nil {
			if light, err = veml6030.Read(); err != nil {
				fmt.Println(err)
				return
			}
		}

		var rng uint16
		if vl53l1x != nil {
			if rng, err = vl53l1x.Read(); err != nil {
				fmt.Println(err)
				return
			}
		}

		var pressure, temperature float64
		if ms5637 != nil {
			if pressure, temperature, err = ms5637.Read(); err != nil {
				fmt.Println(err)
				return
			}
		}

		var bme280temp, bme280pressure, bme280humidity float64
		if bme280 != nil {
			if bme280temp, bme280pressure, bme280humidity, err = bme280.Read(); err != nil {
				fmt.Println(err)
				return
			}
			// REVISIT: apply weird offsets relative to other sensors
			bme280temp += 1.0
			bme280pressure -= 1.4
		}

		prom.SetHumidity(bme280humidity)

		if config.EnableBME280 {
			hkb.SetTemperature(bme280temp)
		} else {
			hkb.SetTemperature(tempC)
		}
		hkb.SetLightLevel(light)
		hkb.SetRangeSensor(rng)

		prom.SetTemperature(tempC)
		prom.SetTemperatureBME(bme280temp)
		prom.SetTemperatureMS(temperature)
		prom.SetLightLevel(light)
		prom.SetRangeSensor(rng)
		prom.SetPressure(pressure)
		prom.SetPressureBME(bme280pressure)
		if oled != nil {
			// oled.WriteOLED(fmt.Sprintf("%.2f C\n%.2f hPa\n%.2f lux", tempC, pressure, light))
			oled.DisplayTemperature(tempC, fmt.Sprintf("%.2f hPa\n%.2f lux", pressure, light))
		}

		powerState := "n/a"
		if bulb != nil {
			var bulbState *golifx.BulbState
			if bulbState, err = bulb.GetColorState(); err == nil {
				if bulbState.Power {
					powerState = "on "
				} else {
					powerState = "off "
				}

				powerState += fmt.Sprintf("(H: %d, S: %d B: %d, K: %d)", bulbState.Color.Hue, bulbState.Color.Saturation, bulbState.Color.Brightness, bulbState.Color.Kelvin)
			}
		}

		fmt.Printf("%.2f hPa, %.2f C | %.2f C | %.2f lux | %d mm | %.2f C, %.2f hPa, %.2f rH | bulb: %s\n", pressure, temperature, tempC, light, rng, bme280temp, bme280pressure, bme280humidity, powerState)
		time.Sleep(5 * time.Second)
	}
}

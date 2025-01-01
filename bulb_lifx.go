package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/2tvenom/golifx"
	"github.com/gin-gonic/gin"
)

var lifxBulbNames = map[string]string{
	"d0:73:d5:6e:33:86": "Kitchen",
	"d0:73:d5:6d:ff:fd": "Kitchen Lamp",
	"d0:73:d5:64:72:09": "Study",
	"d0:73:d5:6c:df:be": "Living Room Lamp",
	"d0:73:d5:64:77:57": "Living Room TV",
	"d0:73:d5:67:23:88": "Bedroom",
	"d0:73:d5:6e:59:ce": "Bathroom",
}

type LifxBulb struct {
	Name       string
	MACAddress string
	Color      golifx.HSBK
	Power      bool
	Label      string
	bulb       *golifx.Bulb
}

type LifxBulbState struct {
	bulbs []*LifxBulb
}

func NewLifxBulbState() (lbs *LifxBulbState) {
	var err error

	lbs = new(LifxBulbState)

	if err = lbs.GetBulbs(); err != nil {
		return
	}

	return
}

func (b *LifxBulb) Refresh() (err error) {
	var bulbState *golifx.BulbState
	if bulbState, err = b.bulb.GetColorState(); err != nil {
		return
	}

	b.Power = bulbState.Power
	b.Color = *bulbState.Color
	b.Label = bulbState.Label
	return
}

func (lbs *LifxBulbState) GetBulbs() (err error) {
	lbs.bulbs = make([]*LifxBulb, 0)

	var bulbs []*golifx.Bulb
	if bulbs, err = golifx.LookupBulbs(); err == nil && len(bulbs) > 0 {
		time.Sleep(1 * time.Second)

		for _, lb := range bulbs {
			var name string
			if n, found := lifxBulbNames[lb.MacAddress()]; found {
				name = n
			} else {
				name = fmt.Sprintf("Unknown %s", lb.MacAddress())
			}

			b := &LifxBulb{
				Name:       name,
				MACAddress: lb.MacAddress(),
				bulb:       lb,
			}

			if err = b.Refresh(); err == nil {
				lbs.bulbs = append(lbs.bulbs, b)
			}
		}
	}

	return
}

func (lbs *LifxBulbState) FindBulbByMAC(mac string) (bulb *LifxBulb, err error) {
	for _, b := range lbs.bulbs {
		if b.MACAddress == mac {
			bulb = b
			return
		}
	}

	err = fmt.Errorf("failed to find bulb with MAC address '%s'", mac)
	return
}

func (lbs *LifxBulbState) SetBulbPower(mac string, power string) (err error) {
	var bulb *LifxBulb
	if bulb, err = lbs.FindBulbByMAC(mac); err != nil {
		return
	}

	if power == "on" || power == "true" {
		if err = bulb.bulb.SetPowerState(true); err == nil {
			bulb.Power = true
		}
	} else if power == "off" || power == "false" {
		if err = bulb.bulb.SetPowerState(false); err == nil {
			bulb.Power = false
		}
	}

	return
}

func NewBulbLifxRouter(lbs *LifxBulbState) {
	httpRouter := gin.New()
	httpRouter.GET("/lifx/bulbs", func(c *gin.Context) {
		c.JSON(http.StatusOK, lbs.bulbs)
	})

	httpRouter.GET("/lifx/bulb/:mac", func(c *gin.Context) {
		mac := c.Param("mac")
		var err error
		var bulb *LifxBulb

		if bulb, err = lbs.FindBulbByMAC(mac); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": err.Error()})
		} else {
			c.JSON(http.StatusOK, bulb)
		}
	})

	httpRouter.POST("/lifx/bulb/:mac", func(c *gin.Context) {
		var err error

		mac := c.Param("mac")
		power := c.PostForm("power")

		if err = lbs.SetBulbPower(mac, power); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		}
	})

	httpRouter.POST("/lifx/refresh", func(c *gin.Context) {
		var err error
		if err = lbs.GetBulbs(); err != nil {
			return
		}
	})

	http.Handle("/lifx/", httpRouter.Handler())
}

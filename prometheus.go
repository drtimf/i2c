package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusSensors struct {
	wdHDPrice [hdTypeCapacitySize]prometheus.Gauge
}

func PrometheusStart(ctx context.Context) (p *PrometheusSensors) {
	p = &PrometheusSensors{}

	p.wdHDPrice[external14TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_14tb",
		Help: "The current Western Digital HD Price for 14TB",
	})
	p.wdHDPrice[external16TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_16tb",
		Help: "The current Western Digital HD Price for 16TB",
	})
	p.wdHDPrice[external18TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_18tb",
		Help: "The current Western Digital HD Price for 18TB",
	})
	p.wdHDPrice[external20TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_20tb",
		Help: "The current Western Digital HD Price for 20TB",
	})
	p.wdHDPrice[external22TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_22tb",
		Help: "The current Western Digital HD Price for 22TB",
	})

	p.wdHDPrice[red14TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_red_14tb",
		Help: "The current Western Digital HD Price for Red 14TB",
	})
	p.wdHDPrice[red16TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_red_16tb",
		Help: "The current Western Digital HD Price for Red 16TB",
	})
	p.wdHDPrice[red18TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_red_18tb",
		Help: "The current Western Digital HD Price for Red 18TB",
	})
	p.wdHDPrice[red20TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_red_20tb",
		Help: "The current Western Digital HD Price for Red 20TB",
	})
	p.wdHDPrice[red22TB] = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monitor_wd_hd_price_red_22tb",
		Help: "The current Western Digital HD Price for Red 22TB",
	})

	server := &http.Server{
		Addr: ":2112",
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Prometheus HTTP server error: ", err)
		}
		fmt.Println("Prometheus HTTP server shutdown")
	}()

	go func() {
		select {
		case <-ctx.Done():
			if err := server.Shutdown(ctx); err != nil {
				fmt.Println("Prometheus HTTP server shutdown error: ", err)
			}
			return
		}
	}()

	return
}

func (p *PrometheusSensors) SetWesternDigitalHDPrice(wd *WesternDigitalDiskPrices) {
	if wd != nil {
		var hdtc hdTypeCapacity = 0
		for ; hdtc < hdTypeCapacitySize; hdtc++ {
			if wd.pricePerTB(hdtc) > 0 {
				p.wdHDPrice[hdtc].Set(wd.pricePerTB(hdtc))
			}
		}
	}
}

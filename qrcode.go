package main

import (
	"strconv"
	"strings"

	"github.com/skip2/go-qrcode"
)

const (
	HAP_TYPE_IP                = 0
	HAP_TYPE_BLE               = 1
	HAP_TYPE_IP_WAC            = 2
	ACCESSORY_CATEGORY_BRIDGES = 2
)

func CreateXHMUrl(category byte, hapType int, pin uint32, setupID string) (xhm string) {
	var payload uint64 = ((uint64(category) << 31) | uint64(pin))

	if hapType == HAP_TYPE_IP_WAC {
		payload |= (1 << 30)
	}

	if hapType == HAP_TYPE_BLE {
		payload |= (1 << 29)
	}

	if hapType == HAP_TYPE_IP || hapType == HAP_TYPE_IP_WAC {
		payload |= (1 << 28)
	}

	b36Payload := strings.ToUpper(strconv.FormatUint(payload, 36))

	for len(b36Payload) < 9 {
		b36Payload = "0" + b36Payload
	}

	xhm = "X-HM://" + b36Payload + setupID
	return
}

func GenCLIQRCode(xhm string) (qr string, err error) {
	var q *qrcode.QRCode
	if q, err = qrcode.New(xhm, qrcode.Highest); err != nil {
		return
	}

	qr = q.ToString(false)
	return
}

package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/platforms/firmata"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

const (
	GETSENSOR = 5
)

func init() {

}

type Client struct {
	Bot *gobot.Robot
}

func NewClient() *Client {
	// ラズパイ
	r := raspi.NewAdaptor()
	relay := gpio.NewGroveRelayDriver(r, "7")

	// 湿度センサー
	firmataAdaptor := firmata.NewAdaptor()
	sensor := aio.NewAnalogSensorDriver(firmataAdaptor, "0")

	// tick worker
	toggle := func() {
		gobot.Every(time.Duration(GETSENSOR)*time.Second, func() {
			if !isWater(sensor) {
				return
			}
			if err := relay.Toggle(); err != nil {
				logrus.Error(err)
				return
			}
			logrus.Infof("toggle switch")
		})
	}

	robot := gobot.NewRobot("onOff",
		[]gobot.Connection{r, firmataAdaptor},
		[]gobot.Device{relay, sensor},
		toggle,
	)

	return &Client{
		Bot: robot,
	}
}

func main() {
	done := make(chan bool)

	client := NewClient()
	client.Bot.Start()
	defer client.Bot.Stop()

	<-done
}

// isWater is
// センサーのデータを読み取る
// - 温度 & 湿度
func isWater(sensor *aio.AnalogSensorDriver) bool {
	val, err := sensor.Read()
	if err != nil {
		logrus.Println("Failed to read", err)
		return false
	}
	cel := (5.0 * float64(val) * 100.0) / 1024
	logrus.Infof("Raw-value:%d Celsius:%.2f", val, cel)

	// 湿度が高ければreturn false
	if 30 < cel {
		return false
	}

	return true
}

package main

import (
	"fmt"
	"github.com/immesys/spawnpoint/spawnable"
	"github.com/PuerkitoBio/goquery"
	"github.com/satori/go.uuid"
	bw2 "gopkg.in/immesys/bw2bind.v5"
	"time"
)

var NAMESPACE_UUID uuid.UUID

func init() {
	NAMESPACE_UUID = uuid.FromStringOrNil("d8b61708-2797-11e6-836b-0cc47a0f7eea")
}

type TimeseriesReading struct {
	UUID  string
	Time  int64
	Value float64
}

func (msg TimeseriesReading) ToMsgPackBW() (po bw2.PayloadObject) {
	po, _ = bw2.CreateMsgPackPayloadObject(bw2.FromDotForm("2.0.9.1"), msg)
	return
}

//needs lots of work
func main() {
	bw := bw2.ConnectOrExit("")

	params := spawnable.GetParamsOrExit()
	apikey := params.MustString("API_KEY")
	city := params.MustString("city")
	baseuri := params.MustString("svc_base_uri")
	read_rate := params.MustString("read_rate")

	bw.OverrideAutoChainTo(true)
	bw.SetEntityFromEnvironOrExit()
	svc := bw.RegisterService(baseuri, "s.caiso")
	iface := svc.RegisterInterface(city, "i.weather")

	params.MergeMetadata(bw)

	fmt.Println(iface.FullURI())
	fmt.Println(iface.SignalURI("fahrenheit"))

	// generate UUIDs from city + metric name
	temp_f_uuid := uuid.NewV3(NAMESPACE_UUID, city+"fahrenheit").String()
	temp_c_uuid := uuid.NewV3(NAMESPACE_UUID, city+"celsius").String()
	relative_humidity_uuid := uuid.NewV3(NAMESPACE_UUID, city+"relative_humidity").String()

	src := NewCaisoEnergySource(read_rate)
	data := src.Start()
	for point := range data {
		fmt.Println(point)
		temp_f := TimeseriesReading{UUID: temp_f_uuid, Time: time.Now().Unix(), Value: point.solarProd}
		iface.PublishSignal("fahrenheit", temp_f.ToMsgPackBW())

		temp_c := TimeseriesReading{UUID: temp_c_uuid, Time: time.Now().Unix(), Value: point.windProd}
		iface.PublishSignal("celsius", temp_c.ToMsgPackBW())

		rh := TimeseriesReading{UUID: relative_humidity_uuid, Time: time.Now().Unix(), Value: point.date}
		iface.PublishSignal("relative_humidity", rh.ToMsgPackBW())
	}
}
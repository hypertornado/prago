package prago

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type Beacon struct {
	ID int64 `prago-order-desc:"true"`

	Name    string
	PageURL string

	Value1Str string `prago-type:"text"`
	Value2Str string `prago-type:"text"`
	Value3Str string `prago-type:"text"`

	Value1Int int64

	UserUUID  string
	UserAgent string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type BeaconRequest struct {
	Name    string
	PageURL string

	Value1Str string
	Value2Str string
	Value3Str string
}

func (beacon *Beacon) UpdateInt() {
	val, err := strconv.Atoi(beacon.Value1Str)
	if err == nil {
		beacon.Value1Int = int64(val)
	}
}

func (app *App) initBeacons() {
	NewResource[Beacon](app).Name(unlocalized("Beacon"), unlocalized("Beacony")).Board(app.optionsBoard)

	app.SetLogHandler(func(typ, message string) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic in custom log handler: %s\n", err)
			}
		}()

		if typ == "error" || typ == "panic" || typ == "info" {
			if strings.Contains(message, "message=context canceled") {
				return
			}
			app.InternalBeacon("logger", typ, message, "")
		}
	})

	app.Handle("POST", "/beacon", func(request *Request) {
		data, err := io.ReadAll(request.Request().Body)
		must(err)

		var beaconRequest BeaconRequest
		err = json.Unmarshal(data, &beaconRequest)
		must(err)

		beacon := &Beacon{
			Name:    beaconRequest.Name,
			PageURL: beaconRequest.PageURL,

			Value1Str: beaconRequest.Value1Str,
			Value2Str: beaconRequest.Value2Str,
			Value3Str: beaconRequest.Value3Str,

			UserUUID:  app.GetTrackingUUIDFunc(request),
			UserAgent: request.Request().UserAgent(),
		}

		beacon.UpdateInt()
		must(CreateItem(app, beacon))
	})

	go deleteOldBeacons(app)

}

func (app *App) InternalBeacon(typ, val1, val2, val3 string) {
	beacon := &Beacon{
		Name:      typ,
		Value1Str: val1,
		Value2Str: val2,
		Value3Str: val3,
	}
	beacon.UpdateInt()
	must(CreateItem(app, beacon))
}

func deleteOldBeacons(app *App) {
	for {
		time.Sleep(5 * time.Minute)
		doDeleteOldBeacons(app)
	}
}

func doDeleteOldBeacons(app *App) {

	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovering from doDeleteOldBeacons panic: %v", err)
		}
	}()

	olderThen := time.Now().AddDate(0, 0, -30)
	beaconsToDelete := Query[Beacon](app).Where("createdat < ?", olderThen).Limit(1000).List()
	for _, v := range beaconsToDelete {
		DeleteItem[Beacon](app, v.ID)
	}

}

func (app *App) GetBeaconCount(beaconName, pageURL, val1, val2, val3 string) int64 {
	q := Query[Beacon](app)
	newerThen := time.Now().AddDate(0, 0, -7)
	q.Where("createdat > ?", newerThen)

	if beaconName != "" {
		q.Is("Name", beaconName)
	}
	if pageURL != "" {
		q.Is("PageURL", pageURL)
	}
	if val1 != "" {
		q.Is("Value1Str", val1)
	}
	if val2 != "" {
		q.Is("Value2Str", val2)
	}
	if val3 != "" {
		q.Is("Value3Str", val3)
	}

	res, err := q.Count()
	if err != nil {
		panic(err)
	}

	return res
}

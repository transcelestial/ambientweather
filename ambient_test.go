package ambient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewKey(t *testing.T) {
	key := NewKey("", "")
	assert.Empty(t, key.APIKey())
	assert.Empty(t, key.ApplicationKey())

	key.SetAPIKey("apikey")
	assert.Equal(t, "apikey", key.APIKey())

	key.SetApplicationKey("appkey")
	assert.Equal(t, "appkey", key.ApplicationKey())
}

func TestGetDevice(t *testing.T) {
	key := NewKey("appkey", "apikey")

	now := time.Now().UTC()
	date := now.Format(time.RFC3339Nano)
	dateunix := now.Unix()

	data := []byte(fmt.Sprintf(`[{
    "macAddress": "00:AA:BB:CC:DD:EE",
    "lastData": {
      "dateutc": %d,
      "winddir": 122,
      "windspeedmph": 5.37,
      "windgustmph": 8.05,
      "maxdailygust": 18.34,
      "tempf": 86.7,
      "battout": 1,
      "humidity": 66,
      "hourlyrainin": 100,
      "eventrainin": 0.78,
      "dailyrainin": 2000,
      "weeklyrainin": 0.78,
      "monthlyrainin": 5.13,
      "yearlyrainin": 26.54,
      "totalrainin": 26.54,
      "uv": 4,
      "solarradiation": 457.17,
      "feelsLike": 94.97,
      "dewPoint": 73.95,
      "lastRain": "%s",
      "tz": "",
      "date": "%s"
    },
    "info": {
      "name": "Amundsen-Scott South Pole Station",
      "coords": {
        "coords": {
          "lon": 139.2728900,
          "lat": -89.9975500
        },
        "address": "Amundsen-Scott South Pole Station, Antarctica",
        "location": "Antarctica",
        "elevation": 0,
        "geo": {
          "type": "Point",
          "coordinates": [-89.9975500, 139.2728900]
        }
      }
    }
  }]`, dateunix, date, date))

	rc := make(chan *http.Request, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc <- r.Clone(context.Background())
		_, err := w.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if !assert.NoError(t, err) {
		return
	}

	// check request
	req := <-rc
	assert.Equal(t, http.MethodGet, req.Method)
	assert.Equal(t, "/devices?applicationKey=appkey&apiKey=apikey", req.URL.String())

	// check response
	assert.Equal(t, http.StatusOK, resp.HTTPResponseCode)
	assert.True(t, bytes.Equal(data, resp.JSONResponse))
	if assert.Len(t, resp.DeviceRecords, 1) {
		dev := resp.DeviceRecords[0]
		assert.Equal(t, "00:AA:BB:CC:DD:EE", dev.MacAddress)
		// last data
		assert.Equal(t, dateunix, dev.LastData.DateUTC)
		assert.Equal(t, 122, dev.LastData.WindDir)
		assert.Equal(t, 5.37, dev.LastData.WindSpeedMph)
		assert.Equal(t, 8.05, dev.LastData.WindGustMph)
		assert.Equal(t, 18.34, dev.LastData.MaxDailyGust)
		assert.Equal(t, 86.7, dev.LastData.Temp)
		assert.Equal(t, 1, dev.LastData.BatteryOut)
		assert.Equal(t, 66, dev.LastData.Humidity)
		assert.Equal(t, 100.0, dev.LastData.HourlyRain)
		assert.Equal(t, 0.78, dev.LastData.EventRain)
		assert.Equal(t, 2000.0, dev.LastData.DailyRain)
		assert.Equal(t, 0.78, dev.LastData.WeeklyRain)
		assert.Equal(t, 5.13, dev.LastData.MonthlyRain)
		assert.Equal(t, 26.54, dev.LastData.YearlyRain)
		assert.Equal(t, 26.54, dev.LastData.TotalRain)
		assert.Equal(t, 4, dev.LastData.Uv)
		assert.Equal(t, 457.17, dev.LastData.SolarRadiation)
		assert.Equal(t, 94.97, dev.LastData.FeelsLike)
		assert.Equal(t, 73.95, dev.LastData.DewPoint)
		assert.True(t, now.Equal(dev.LastData.LastRain))
		assert.Empty(t, dev.LastData.Tz)
		assert.True(t, now.Equal(dev.LastData.Date))
		// info
		assert.Equal(t, "Amundsen-Scott South Pole Station", dev.Info.Name)
		assert.Equal(t, 139.2728900, dev.Info.Coords.Coords.Lon)
		assert.Equal(t, -89.9975500, dev.Info.Coords.Coords.Lat)
		assert.Equal(t, "Amundsen-Scott South Pole Station, Antarctica", dev.Info.Coords.Address)
		assert.Equal(t, "Antarctica", dev.Info.Coords.Location)
		assert.Equal(t, 0.0, dev.Info.Coords.Elevation)
		assert.Equal(t, "Point", dev.Info.Coords.Geo.Type)
		if assert.Len(t, dev.Info.Coords.Geo.Coordinates, 2) {
			assert.Equal(t, -89.9975500, dev.Info.Coords.Geo.Coordinates[0])
			assert.Equal(t, 139.2728900, dev.Info.Coords.Geo.Coordinates[1])
		}
	}
}

func TestBadAddressParamError(t *testing.T) {
	key := NewKey("appkey", "apikey")
	data, err := GetDevice(key, SetAddress("bad address"))
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unsupported protocol scheme")
	}
	assert.Nil(t, data.DeviceRecords)
	assert.Nil(t, data.JSONResponse)
	assert.Equal(t, 0, data.HTTPResponseCode)
}

func TestUnauthorizedError(t *testing.T) {
	key := NewKey("appkey", "apikey")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusUnauthorized)
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if assert.Error(t, err) {
		assert.Equal(t, "API/app key not authorized", err.Error())
	}
	assert.Equal(t, http.StatusUnauthorized, resp.HTTPResponseCode)
}

func TestTooManyRequestsError(t *testing.T) {
	key := NewKey("appkey", "apikey")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusTooManyRequests)
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if assert.Error(t, err) {
		assert.Equal(t, "too many requests. rate limit exceeded", err.Error())
	}
	assert.Equal(t, http.StatusTooManyRequests, resp.HTTPResponseCode)
}

func TestBadGatewayError(t *testing.T) {
	key := NewKey("appkey", "apikey")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadGateway)
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if assert.Error(t, err) {
		assert.Equal(t, "ambientweather service is unreachable", err.Error())
	}
	assert.Equal(t, http.StatusBadGateway, resp.HTTPResponseCode)
}

func TestServiceUnavailableError(t *testing.T) {
	key := NewKey("appkey", "apikey")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if assert.Error(t, err) {
		assert.Equal(t, "ambientweather service is unreachable", err.Error())
	}
	assert.Equal(t, http.StatusServiceUnavailable, resp.HTTPResponseCode)
}

func TestOtherStatusCodeErrors(t *testing.T) {
	key := NewKey("appkey", "apikey")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadRequest)
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if assert.Error(t, err) {
		assert.Equal(t, "request failed with response code: 400", err.Error())
	}
	assert.Equal(t, http.StatusBadRequest, resp.HTTPResponseCode)
}

func TestBadBodyError(t *testing.T) {
	key := NewKey("appkey", "apikey")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if assert.Error(t, err) {
		assert.Equal(t, io.ErrUnexpectedEOF.Error(), err.Error())
	}
	assert.Equal(t, http.StatusOK, resp.HTTPResponseCode)
}

func TestBadJSONError(t *testing.T) {
	key := NewKey("appkey", "apikey")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`[{"macAddress": 0}]`))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	resp, err := GetDevice(key, SetAddress(ts.URL))
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "cannot parse body")
	}
	assert.Equal(t, http.StatusOK, resp.HTTPResponseCode)
}

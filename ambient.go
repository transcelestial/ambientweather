package ambient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const Version = "v1"

const URL = "https://api.ambientweather.net/" + Version

type Record struct {
	DateUTC        int64     `json:"dateutc"`
	WindDir        int       `json:"winddir"`
	WindSpeedMph   float64   `json:"windspeedmph"`
	WindGustMph    float64   `json:"windgustmph"`
	MaxDailyGust   float64   `json:"maxdailygust"`
	Temp           float64   `json:"tempf"`
	BatteryOut     int       `json:"battout"`
	Humidity       int       `json:"humidity"`
	HourlyRain     float64   `json:"hourlyrainin"`
	EventRain      float64   `json:"eventrainin"`
	DailyRain      float64   `json:"dailyrainin"`
	WeeklyRain     float64   `json:"weeklyrainin"`
	MonthlyRain    float64   `json:"monthlyrainin"`
	YearlyRain     float64   `json:"yearlyrainin"`
	TotalRain      float64   `json:"totalrainin"`
	Uv             int       `json:"uv"`
	SolarRadiation float64   `json:"solarradiation"`
	FeelsLike      float64   `json:"feelslike"`
	DewPoint       float64   `json:"dewpoint"`
	LastRain       time.Time `json:"lastrain"`
	Tz             string    `json:"tz"`
	Date           time.Time `json:"date"`
}

type DeviceGeo struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type DeviceLatLong struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type DeviceCoords struct {
	Coords    DeviceLatLong `json:"coords"`
	Address   string        `json:"address"`
	Location  string        `json:"location"`
	Elevation float64       `json:"elevation"`
	Geo       DeviceGeo     `json:"geo"`
}

type DeviceInfo struct {
	Name   string       `json:"name"`
	Coords DeviceCoords `json:"coords"`
}

type DeviceRecord struct {
	MacAddress string     `json:"macaddress"`
	Info       DeviceInfo `json:"info"`
	LastData   Record     `json:"lastdata"`
}

type APIDeviceResponse struct {
	DeviceRecords    []DeviceRecord
	JSONResponse     []byte
	HTTPResponseCode int
	ResponseTime     time.Duration
}

type Key struct {
	applicationKey string
	apiKey         string
}

// NewKey creates a new key to use for authentication w/ the Ambientweather service.
func NewKey(applicationKey string, apiKey string) Key {
	return Key{applicationKey: applicationKey, apiKey: apiKey}
}

func (Key Key) APIKey() string {
	return Key.apiKey
}

func (Key Key) ApplicationKey() string {
	return Key.applicationKey
}

func (Key *Key) SetApplicationKey(applicationKey string) {
	Key.applicationKey = applicationKey
}

func (Key *Key) SetAPIKey(apiKey string) {
	Key.apiKey = apiKey
}

type DeviceParam func(*DeviceParams)

type DeviceParams struct {
	Address string
}

// SetAddress sets the address of the API.
// Can be used for tests with httptest.
func SetAddress(address string) DeviceParam {
	return func(params *DeviceParams) {
		params.Address = address
	}
}

// GetDevice queries device data from Ambientweather.
func GetDevice(key Key, params ...DeviceParam) (APIDeviceResponse, error) {
	var ar APIDeviceResponse

	path := "/devices?applicationKey=" + key.applicationKey + "&apiKey=" + key.apiKey
	dp := &DeviceParams{
		Address: URL,
	}
	for _, p := range params {
		p(dp)
	}

	startTime := time.Now()
	resp, err := http.Get(dp.Address + path)
	ar.ResponseTime = time.Since(startTime)
	if err != nil {
		return ar, err
	}

	ar.HTTPResponseCode = resp.StatusCode
	ar.JSONResponse, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return ar, err
	}

	switch resp.StatusCode {
	case 200:
	case 401:
		return ar, errors.New("API/app key not authorized")
	case 429:
		// https://ambientweather.docs.apiary.io/#introduction/rate-limiting
		return ar, errors.New("too many requests. rate limit exceeded")
	case 502, 503:
		return ar, errors.New("ambientweather service is unreachable")
	default:
		return ar, fmt.Errorf("request failed with response code: %d", resp.StatusCode)
	}

	err = json.Unmarshal(ar.JSONResponse, &ar.DeviceRecords)
	if err != nil {
		return ar, fmt.Errorf("cannot parse body: %v", err)
	}

	return ar, nil
}

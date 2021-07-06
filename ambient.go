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
	Dateutc        int64     `json:"dateutc"`
	Winddir        int       `json:"winddir"`
	Windspeedmph   float64   `json:"windspeedmph"`
	Windgustmph    float64   `json:"windgustmph"`
	Maxdailygust   float64   `json:"maxdailygust"`
	Tempf          float64   `json:"tempf"`
	Battout        int       `json:"battout"`
	Humidity       int       `json:"humidity"`
	Hourlyrainin   float64   `json:"hourlyrainin"`
	Eventrainin    float64   `json:"eventrainin"`
	Dailyrainin    float64   `json:"dailyrainin"`
	Weeklyrainin   float64   `json:"weeklyrainin"`
	Monthlyrainin  float64   `json:"monthlyrainin"`
	Yearlyrainin   float64   `json:"yearlyrainin"`
	Totalrainin    float64   `json:"totalrainin"`
	Uv             int       `json:"uv"`
	Solarradiation float64   `json:"solarradiation"`
	Feelslike      float64   `json:"feelslike"`
	Dewpoint       float64   `json:"dewpoint"`
	Lastrain       time.Time `json:"lastrain"`
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
	Macaddress string     `json:"macaddress"`
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
	URL string
}

func SetURL(url string) DeviceParam {
	return func(params *DeviceParams) {
		params.URL = url
	}
}

func GetDevice(key Key, params ...DeviceParam) (APIDeviceResponse, error) {
	var ar APIDeviceResponse

	dp := &DeviceParams{
		URL: URL + "/devices?applicationKey=" + key.applicationKey + "&apiKey=" + key.apiKey,
	}
	for _, p := range params {
		p(dp)
	}

	startTime := time.Now()
	resp, err := http.Get(dp.URL)
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
		{
			return ar, errors.New("API/APP key not authorized")
		}
	case 429, 502, 503:
		{
			if resp.StatusCode >= 500 {
				ar.JSONResponse, _ = json.Marshal(fmt.Sprintf("{\"errormessage\": \"HTTP Error Code: %d\"}", resp.StatusCode))
			}
			return ar, nil
		}
	default:
		{
			return ar, errors.New("bad non-200/429/502/503 Response Code")
		}
	}
	err = json.Unmarshal(ar.JSONResponse, &ar.DeviceRecords)
	if err != nil {
		return ar, err
	}
	return ar, nil
}

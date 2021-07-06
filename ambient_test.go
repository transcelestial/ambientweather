package ambient

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var awsApiKey = flag.String("AMBIENT_APIKEY", os.Getenv("AMBIENT_APIKEY"), "Ambientweather API key")
var awsAppKey = flag.String("AMBIENT_APPKEY", os.Getenv("AMBIENT_APPKEY"), "Ambientweather APP Key")

func TestAmbientNewKeys(t *testing.T) {
	newKey := NewKey("", "")
	assert.Empty(t, newKey.APIKey())
	assert.Empty(t, newKey.ApplicationKey())

	newKey.SetAPIKey(*awsApiKey)
	assert.EqualValues(t, *awsApiKey, newKey.APIKey())

	newKey.SetApplicationKey(*awsAppKey)
	assert.EqualValues(t, *awsAppKey, newKey.ApplicationKey())
}

func TestErrInvalidURL(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)
	data, err := GetDevice(newKey, SetURL("invalid url"))

	assert.Nil(t, data.DeviceRecords)
	assert.Nil(t, data.JSONResponse)
	assert.Equal(t, 0, data.HTTPResponseCode)
	assert.Equal(t, "Get \"invalid%20url\": unsupported protocol scheme \"\"", err.Error())
}

func TestErrUnauthorized(t *testing.T) {
	newKey := NewKey("", *awsApiKey)

	data, err := GetDevice(newKey)
	assert.Equal(t, http.StatusUnauthorized, data.HTTPResponseCode)
	assert.Equal(t, "API/APP key not authorized", err.Error())

	newKey = NewKey(*awsAppKey, "")

	data, err = GetDevice(newKey)
	assert.Equal(t, http.StatusUnauthorized, data.HTTPResponseCode)
	assert.Equal(t, "API/APP key not authorized", err.Error())
}

func TestErrTooManyRequest(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusTooManyRequests)
	}))
	defer ts.Close()

	data, err := GetDevice(newKey, SetURL(ts.URL))
	assert.Equal(t, http.StatusTooManyRequests, data.HTTPResponseCode)
	assert.Nil(t, err)
}

func TestErrBadGateWay(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusBadGateway)
	}))
	defer ts.Close()

	data, err := GetDevice(newKey, SetURL(ts.URL))
	assert.Equal(t, http.StatusBadGateway, data.HTTPResponseCode)
	assert.Nil(t, err)
}

func TestErrServiceUnavailable(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	data, err := GetDevice(newKey, SetURL(ts.URL))
	assert.Equal(t, http.StatusServiceUnavailable, data.HTTPResponseCode)
	assert.Nil(t, err)
}

func TestJSONBodyEOF(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
	}))
	defer ts.Close()

	data, err := GetDevice(newKey, SetURL(ts.URL))
	assert.Equal(t, http.StatusOK, data.HTTPResponseCode)
	assert.EqualError(t, io.ErrUnexpectedEOF, err.Error())
}

func TestErrUnchecked(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "", http.StatusNotFound)
	}))
	defer ts.Close()

	data, err := GetDevice(newKey, SetURL(ts.URL))
	assert.Equal(t, http.StatusNotFound, data.HTTPResponseCode)
	assert.Equal(t, "bad non-200/429/502/503 Response Code", err.Error())
}

func TestErrJsonMarshalWrongDataType(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)

	jsonData := `[{"macAddress":100,"lastData":{"dateutc":1624263420000,"winddir":122,"windspeedmph":5.37,"windgustmph":8.05,"maxdailygust":18.34,"tempf":86.7,"battout":1,"humidity":66,"hourlyrainin":0,"eventrainin":0.78,"dailyrainin":0,"weeklyrainin":0.78,"monthlyrainin":5.13,"yearlyrainin":26.54,"totalrainin":26.54,"uv":4,"solarradiation":457.17,"feelsLike":94.97,"dewPoint":73.95,"lastRain":"2021-06-20T15:03:00.000Z","tz":"Asia/Singapore","date":"2021-06-21T08:17:00.000Z"},"info":{"name":"PPC","coords":{"coords":{"lon":103.8440929,"lat":1.285631},"address":"101 Upper Cross St, Singapore 058357","location":"Singapore","elevation":2.285359382629395,"geo":{"type":"Point","coordinates":[103.8440929,1.285631]}}}}]`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(jsonData)) // nolint: errcheck
	}))
	defer ts.Close()

	data, err := GetDevice(newKey, SetURL(ts.URL))
	assert.Equal(t, http.StatusOK, data.HTTPResponseCode)
	assert.Equal(t, "json: cannot unmarshal number into Go struct field DeviceRecord.macaddress of type string", err.Error())
}

func TestNoErr(t *testing.T) {
	newKey := NewKey(*awsAppKey, *awsApiKey)

	jsonData := `[{"macAddress":"00:0E:C6:30:1A:8B","lastData":{"dateutc":1624263420000,"winddir":122,"windspeedmph":5.37,"windgustmph":8.05,"maxdailygust":18.34,"tempf":86.7,"battout":1,"humidity":66,"hourlyrainin":0,"eventrainin":0.78,"dailyrainin":0,"weeklyrainin":0.78,"monthlyrainin":5.13,"yearlyrainin":26.54,"totalrainin":26.54,"uv":4,"solarradiation":457.17,"feelsLike":94.97,"dewPoint":73.95,"lastRain":"2021-06-20T15:03:00.000Z","tz":"Asia/Singapore","date":"2021-06-21T08:17:00.000Z"},"info":{"name":"PPC","coords":{"coords":{"lon":103.8440929,"lat":1.285631},"address":"101 Upper Cross St, Singapore 058357","location":"Singapore","elevation":2.285359382629395,"geo":{"type":"Point","coordinates":[103.8440929,1.285631]}}}}]`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(jsonData)) // nolint: errcheck
	}))
	defer ts.Close()

	data, err := GetDevice(newKey, SetURL(ts.URL))
	assert.Equal(t, http.StatusOK, data.HTTPResponseCode)
	assert.NoError(t, err)
}

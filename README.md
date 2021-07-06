# Ambientweather
> Go client for the [Ambientweather](https://ambientweather.docs.apiary.io/) service.

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/transcelestial/ambientweather/Test?label=test&style=flat-square)](https://github.com/transcelestial/ambientweather/actions?query=workflow%3ATest)

## Usage
1. Generate an app and API key from your [account](https://ambientweather.net/account)

2. Use the API:
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	ambient "github.com/transcelestial/ambientweather"
)

func main() {
	key := ambient.NewKey("appkey", "apikey")

	devices, err := ambient.GetDevice(key)
	if err != nil {
		log.Fatal(err)
	}

	if devices.HTTPResponseCode == http.StatusOK {
		for _, dev := range devices.DeviceRecords {
			fmt.Println(dev)
		}
	}
}
```


## Test
To run tests run:
```bash
go test .
```

To avoid caching during tests use:
```bash
go test -count=1 .
```

To get coverage reports use the `-cover` flag:
```bash
go test -coverprofile=coverage.out .
```

And to view the profile run:
```bash
go tool cover -html=coverage.out
```

To run static analysis on the code run:
```bash
go vet .
```

## Contribute
If you wish to contribute, please use the following guidelines:
* Use [conventional commits](https://conventionalcommits.org/)
* Use [effective Go](https://golang.org/doc/effective_go)

## Credits/Contributors
* [lrosenman/ambient](https://github.com/lrosenman/ambient)
* [Darryl Tan](https://github.com/moobshake)

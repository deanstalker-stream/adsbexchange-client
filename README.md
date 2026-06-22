# adsbexchange-client

[![Lint](https://github.com/deanstalker-stream/adsbexchange-client/actions/workflows/lint.yml/badge.svg)](https://github.com/deanstalker-stream/adsbexchange-client/actions/workflows/lint.yml) [![Test](https://github.com/deanstalker-stream/adsbexchange-client/actions/workflows/test.yml/badge.svg)](https://github.com/deanstalker-stream/adsbexchange-client/actions/workflows/test.yml)

Provides a client for the Community API provided by ADSB Exchange. This will not work for their Commercial API.

https://rapidapi.com/adsbx/api/adsbexchange-com1

## Usage

### Configuration

The client is configured via [Viper](https://github.com/spf13/viper), which supports both a `config.json` file and environment variables.

#### config.json

```json
{
  "feed": {
    "adsbexchange": {
      "host": "adsbexchange-com1.p.rapidapi.com",
      "key": "your-rapidapi-key"
    }
  }
}
```

#### Environment Variables

| Variable                  | Description                |
|---------------------------|----------------------------|
| `FEED_ADSBEXCHANGE_HOST`  | RapidAPI host for the feed |
| `FEED_ADSBEXCHANGE_KEY`   | RapidAPI key               |

### Example

```go
package main

import (
    "fmt"
    "log"

    adsbexchange "github.com/deanstalker-stream/adsbexchange-client"
    "github.com/spf13/viper"
    "go.uber.org/zap"
)

func main() {
    v := viper.New()
    v.SetConfigName("config")
    v.SetConfigType("json")
    v.AddConfigPath(".")

    // Register environment variable bindings
    adsbexchange.BindEnvs(v)

    // Read config.json if present; environment variables are always checked
    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            log.Fatalf("error reading config: %v", err)
        }
    }

    var cfg adsbexchange.Config
    if err := v.UnmarshalKey("feed.adsbexchange", &cfg); err != nil {
        log.Fatalf("error unmarshalling config: %v", err)
    }

    logger, _ := zap.NewProduction()
    client, err := adsbexchange.NewClient(logger, &cfg)
    if err != nil {
        log.Fatalf("error creating client: %v", err)
    }

    aircraft, err := client.GetMilitaryAircraft()
    if err != nil {
        log.Fatalf("error fetching aircraft: %v", err)
    }

    for _, ac := range aircraft {
        fmt.Printf("Flight: %s, Hex: %s\n", ac.Flight, ac.Hex)
    }
}
```

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## License

Copyright 2026 Dean Stalker

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
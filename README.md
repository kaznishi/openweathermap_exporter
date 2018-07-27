## インストール

```
go get github.com/kaznishi/openweathermap_exporter
```

## 使い方

```
openweathermap_exporter
```

を実行するとWEBサーバプロセスが立ち上がります。

```
http://[HOST IP]:9999/metrics?location=[CITY NAME]&api_key=[API KEY]
```

に適宜HOST IPとCITY NAMEとAPI_KEYを埋めてアクセスすると以下のような表示のページが出力されます。

```
# HELP openweathermap_humidity_percent Humidity in Percent
# TYPE openweathermap_humidity_percent gauge
openweathermap_humidity_percent 65
# HELP openweathermap_pressure_hpa Atmospheric pressure in hPa
# TYPE openweathermap_pressure_hpa gauge
openweathermap_pressure_hpa 1009
# HELP openweathermap_temperature_celsius Temperature in °C
# TYPE openweathermap_temperature_celsius gauge
openweathermap_temperature_celsius 27.4
```

Prometheus本体側の設定ファイル(prometheus.yml)のscrape_configsには以下のような記述を追加してください。

```
- job_name: 'openweathermap'
  scrape_interval: 60s
  metrics_path: /metrics
  params:
    api_key:
      - xxxxxxxxxxxxxxxxxxxxxxxxx
  static_configs:
  - targets:
    - Tokyo
    - Osaka-shi
    - Naha-shi
    - Sapporo-shi
  relabel_configs:
    - source_labels: [__address__]
      target_label: __param_location
    - source_labels: [__param_location]
      target_label: location
    - target_label: __address__
      replacement: 127.0.0.1:9999
```


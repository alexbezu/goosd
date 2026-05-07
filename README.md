# goosd
Golang OSD HUD overlay GUI

## Documentation
- The HUD is currently driven by fake/simulated data so rendering can be verified before MAVLink input exists.
- The Ebitengine window is transparent, frameless, and always on top.
- Mouse click-through overlay mode is available with the `-click-through` flag.
- The first HUD skeleton draws a center reticle, pitch ladder, roll scale, heading text, altitude, and speed.
- The HUD DTO/state package defines attitude, heading, altitude, speed, GPS status, and health flags for MAVLink command mapping from [betaflight](https://github.com/betaflight/betaflight/blob/master/src/main/telemetry/mavlink.c).
- MAVLink UDP input is available with `-mavlink-udp`, for example `go run ./cmd -mavlink-udp :16000`.
- The MAVLink reader maps `ATTITUDE`, `VFR_HUD`, `GLOBAL_POSITION_INT`, `GPS_RAW_INT`, `HEARTBEAT`, and `SYS_STATUS` messages into the HUD DTO.

## TODO
- Later: add config for window size, position, always-on-top, click-through, and data source.

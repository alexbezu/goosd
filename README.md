# goosd
Golang OSD HUD overlay GUI

## Documentation
- The HUD is currently driven by fake/simulated data so rendering can be verified before MAVLink input exists.
- The Ebitengine window is transparent, frameless, and always on top.
- Mouse click-through overlay mode is available with the `-click-through` flag.
- The first HUD skeleton draws a center reticle, pitch ladder, roll scale, heading text, altitude, and speed.
- The HUD DTO/state package defines attitude, heading, altitude, speed, GPS status, and health flags for future MAVLink command mapping from [/Users/oleksii/betaflight/src/main/telemetry/mavlink.c](https://github.com/betaflight/betaflight/blob/master/src/main/telemetry/mavlink.c).

## TODO
- add a gomavlib UDP listener (reader only) and convert incoming MAVLink data into the HUD DTO
- Later: add config for window size, position, always-on-top, click-through, and data source.

package telemetry

import (
	"math"
	"testing"
	"time"

	"github.com/alexbezu/goosd/internal/hud"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

func TestMAVLinkSourceApplyMessages(t *testing.T) {
	source := &MAVLinkSource{
		now: time.Now,
		state: hud.State{
			Health: hud.HealthTelemetryLost,
		},
	}
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)

	source.apply(&common.MessageHeartbeat{
		BaseMode:     common.MAV_MODE_FLAG_SAFETY_ARMED,
		SystemStatus: common.MAV_STATE_ACTIVE,
	}, now)
	source.apply(&common.MessageAttitude{
		Roll:  float32(math.Pi / 6),
		Pitch: float32(-math.Pi / 12),
		Yaw:   float32(math.Pi),
	}, now)
	source.apply(&common.MessageVfrHud{
		Heading:     181,
		Alt:         123.4,
		Groundspeed: 22.5,
	}, now)
	source.apply(&common.MessageGpsRawInt{
		FixType:           common.GPS_FIX_TYPE_3D_FIX,
		Eph:               90,
		SatellitesVisible: 11,
	}, now)
	source.apply(&common.MessageBatteryStatus{
		Voltages:         [10]uint16{3700, 3710, 3690, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16},
		CurrentBattery:   842,
		BatteryRemaining: 73,
	}, now)

	state := source.State(now)

	if !state.Health.Has(hud.HealthArmed) {
		t.Fatal("state should be armed")
	}
	if state.Health.Has(hud.HealthTelemetryLost) {
		t.Fatal("telemetry should be marked present after a message")
	}
	if !closeEnough(state.Attitude.RollDeg, 30) {
		t.Fatalf("roll = %v, want 30", state.Attitude.RollDeg)
	}
	if !closeEnough(state.Attitude.PitchDeg, -15) {
		t.Fatalf("pitch = %v, want -15", state.Attitude.PitchDeg)
	}
	if state.Heading.Deg != 181 {
		t.Fatalf("heading = %v, want 181", state.Heading.Deg)
	}
	if !closeEnough(state.AltitudeM, 123.4) {
		t.Fatalf("altitude = %v, want 123.4", state.AltitudeM)
	}
	if !closeEnough(state.SpeedMS, 22.5) {
		t.Fatalf("speed = %v, want 22.5", state.SpeedMS)
	}
	if state.GPS.FixType != hud.GPSFix3D {
		t.Fatalf("gps fix = %v, want %v", state.GPS.FixType, hud.GPSFix3D)
	}
	if state.GPS.Satellites != 11 {
		t.Fatalf("satellites = %v, want 11", state.GPS.Satellites)
	}
	if state.GPS.HDOP != 0.9 {
		t.Fatalf("hdop = %v, want 0.9", state.GPS.HDOP)
	}
	if !state.Battery.RemainingPctValid || state.Battery.RemainingPct != 73 {
		t.Fatalf("battery pct = %v/%v, want valid 73", state.Battery.RemainingPctValid, state.Battery.RemainingPct)
	}
	if !state.Battery.VoltageValid || !closeEnough(state.Battery.VoltageV, 11.1) {
		t.Fatalf("battery voltage = %v/%v, want valid 11.1", state.Battery.VoltageValid, state.Battery.VoltageV)
	}
	if !state.Battery.CurrentValid || !closeEnough(state.Battery.CurrentA, 8.42) {
		t.Fatalf("battery current = %v/%v, want valid 8.42", state.Battery.CurrentValid, state.Battery.CurrentA)
	}
}

func TestMAVLinkSourceTelemetryTimeout(t *testing.T) {
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	source := &MAVLinkSource{
		state: hud.State{
			Health:    hud.HealthArmed,
			UpdatedAt: now,
		},
	}

	state := source.State(now.Add(4 * time.Second))

	if !state.Health.Has(hud.HealthTelemetryLost) {
		t.Fatal("telemetry timeout should set HealthTelemetryLost")
	}
}

func TestGPSFixType(t *testing.T) {
	tests := []struct {
		in   common.GPS_FIX_TYPE
		want hud.GPSFixType
	}{
		{common.GPS_FIX_TYPE_NO_FIX, hud.GPSFixNone},
		{common.GPS_FIX_TYPE_2D_FIX, hud.GPSFix2D},
		{common.GPS_FIX_TYPE_3D_FIX, hud.GPSFix3D},
		{common.GPS_FIX_TYPE_RTK_FIXED, hud.GPSFix3D},
	}

	for _, test := range tests {
		if got := gpsFixType(test.in); got != test.want {
			t.Fatalf("gpsFixType(%v) = %v, want %v", test.in, got, test.want)
		}
	}
}

func TestBatteryStatusLowAndUnknownValues(t *testing.T) {
	source := &MAVLinkSource{}
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)

	source.apply(&common.MessageBatteryStatus{
		Voltages:         [10]uint16{math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16, math.MaxUint16},
		CurrentBattery:   -1,
		BatteryRemaining: 20,
	}, now)

	state := source.State(now)

	if !state.Health.Has(hud.HealthLowBattery) {
		t.Fatal("battery remaining <= 20 should set low battery health")
	}
	if !state.Battery.RemainingPctValid || state.Battery.RemainingPct != 20 {
		t.Fatalf("battery pct = %v/%v, want valid 20", state.Battery.RemainingPctValid, state.Battery.RemainingPct)
	}
	if state.Battery.VoltageValid {
		t.Fatalf("battery voltage should be invalid, got %v", state.Battery.VoltageV)
	}
	if state.Battery.CurrentValid {
		t.Fatalf("battery current should be invalid, got %v", state.Battery.CurrentA)
	}
}

func closeEnough(got, want float64) bool {
	return math.Abs(got-want) < 0.0001
}

package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/alexbezu/goosd/internal/hud"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 960
	screenHeight = 540
)

var (
	hudGreen = color.RGBA{R: 96, G: 255, B: 142, A: 230}
	hudDim   = color.RGBA{R: 96, G: 255, B: 142, A: 120}
)

type Game struct {
	sim *hud.Simulator
	now func() time.Time
}

func NewGame() *Game {
	return &Game{
		sim: hud.NewSimulator(time.Now()),
		now: time.Now,
	}
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	state := g.sim.State(g.now())
	drawHUD(screen, state)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func drawHUD(screen *ebiten.Image, state hud.State) {
	cx := float32(screenWidth) / 2
	cy := float32(screenHeight) / 2

	drawPitchLadder(screen, cx, cy, state.Attitude.RollDeg, state.Attitude.PitchDeg)
	drawRollScale(screen, cx, cy, state.Attitude.RollDeg)
	drawReticle(screen, cx, cy)
	drawTapeText(screen, state)
}

func drawReticle(screen *ebiten.Image, cx, cy float32) {
	vector.StrokeLine(screen, cx-58, cy, cx-14, cy, 2, hudGreen, true)
	vector.StrokeLine(screen, cx+14, cy, cx+58, cy, 2, hudGreen, true)
	vector.StrokeLine(screen, cx, cy-10, cx, cy+10, 2, hudGreen, true)
	vector.FillCircle(screen, cx, cy, 2.5, hudGreen, true)
}

func drawPitchLadder(screen *ebiten.Image, cx, cy float32, rollDeg, pitchDeg float64) {
	roll := rollDeg * math.Pi / 180
	sinRoll, cosRoll := math.Sin(roll), math.Cos(roll)

	for mark := -40; mark <= 40; mark += 10 {
		offsetY := float64(mark)*8 + pitchDeg*8
		if math.Abs(offsetY) > float64(screenHeight)/2 {
			continue
		}

		width := float32(96)
		if mark == 0 {
			width = 156
		}
		y := float32(offsetY)
		x0, y0 := rotate(-width/2, y, sinRoll, cosRoll)
		x1, y1 := rotate(width/2, y, sinRoll, cosRoll)
		clr := hudGreen
		if mark != 0 {
			clr = hudDim
		}
		vector.StrokeLine(screen, cx+x0, cy+y0, cx+x1, cy+y1, 2, clr, true)

		if mark != 0 {
			label := fmt.Sprintf("%d", int(math.Abs(float64(mark))))
			lx, ly := rotate(width/2+12, y-8, sinRoll, cosRoll)
			ebitenutil.DebugPrintAt(screen, label, int(cx+lx), int(cy+ly))
			lx, ly = rotate(-width/2-28, y-8, sinRoll, cosRoll)
			ebitenutil.DebugPrintAt(screen, label, int(cx+lx), int(cy+ly))
		}
	}
}

func drawRollScale(screen *ebiten.Image, cx, cy float32, rollDeg float64) {
	radius := float32(176)
	for _, mark := range []float64{-60, -45, -30, -20, -10, 0, 10, 20, 30, 45, 60} {
		angle := (mark - 90) * math.Pi / 180
		inner := radius
		if int(math.Abs(mark))%30 == 0 {
			inner -= 18
		} else {
			inner -= 10
		}
		x0 := cx + inner*float32(math.Cos(angle))
		y0 := cy + inner*float32(math.Sin(angle))
		x1 := cx + radius*float32(math.Cos(angle))
		y1 := cy + radius*float32(math.Sin(angle))
		vector.StrokeLine(screen, x0, y0, x1, y1, 2, hudDim, true)
	}

	roll := (rollDeg - 90) * math.Pi / 180
	x := cx + (radius+14)*float32(math.Cos(roll))
	y := cy + (radius+14)*float32(math.Sin(roll))
	vector.StrokeLine(screen, x, y, x-9, y+18, 2, hudGreen, true)
	vector.StrokeLine(screen, x, y, x+9, y+18, 2, hudGreen, true)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ROLL %+03.0f", rollDeg), int(cx)-39, int(cy-radius)-34)
}

func drawTapeText(screen *ebiten.Image, state hud.State) {
	heading := normalize360(state.Heading.Deg)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HDG %03.0f", heading), screenWidth/2-30, 28)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SPD %03.0f m/s", state.SpeedMS), 44, screenHeight/2-8)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ALT %04.0f m", state.AltitudeM), screenWidth-128, screenHeight/2-8)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("GPS %s %02d %.1f", state.GPS.FixType, state.GPS.Satellites, state.GPS.HDOP), 44, screenHeight-42)
	ebitenutil.DebugPrintAt(screen, healthText(state.Health), screenWidth-152, screenHeight-42)
}

func healthText(health hud.Health) string {
	switch {
	case health.Has(hud.HealthFailsafe):
		return "FAILSAFE"
	case health.Has(hud.HealthTelemetryLost):
		return "NO TELEMETRY"
	case health.Has(hud.HealthLowBattery):
		return "LOW BAT"
	case health.Has(hud.HealthArmed):
		return "ARMED"
	default:
		return "DISARMED"
	}
}

func normalize360(deg float64) float64 {
	deg = math.Mod(deg, 360)
	if deg < 0 {
		deg += 360
	}
	return deg
}

func rotate(x, y float32, sin, cos float64) (float32, float32) {
	rx := float64(x)*cos - float64(y)*sin
	ry := float64(x)*sin + float64(y)*cos
	return float32(rx), float32(ry)
}

func main() {
	clickThrough := flag.Bool("click-through", false, "pass mouse input through the HUD window")
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("goosd")
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowMousePassthrough(*clickThrough)

	options := &ebiten.RunGameOptions{
		ScreenTransparent: true,
	}
	if err := ebiten.RunGameWithOptions(NewGame(), options); err != nil {
		log.Fatal(err)
	}
}

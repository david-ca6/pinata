// comments for lemons

package main

import (
	"embed"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strings"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// embed the assets folder in the binary, required for the web build
//
//go:embed assets/*
var assets embed.FS

const (
	screenWidth  = 1280
	screenHeight = 720
	maxLife      = 100
	channel      = "kanekolumi"
)

// represent a simple 2D entity
type Entity2D struct {
	x        float64
	y        float64
	rotation float64
	width    float64
	height   float64
	image    ebiten.Image
}

// represent a projectile (a lemon)
type Projectile struct {
	Entity2D
	speed   float64
	targetX float64
	targetY float64
	moving  bool
}

// represent a candy
type Candy struct {
	Entity2D
	speed   float64
	targetX float64
	targetY float64
	moving  bool
}

// the game struct store the game states
type Game struct {
	life int

	candy       Entity2D
	pinataFront Entity2D
	pinataBack  Entity2D
	projectile  Projectile

	nbCandy      int
	throwChannel chan bool

	audioContext *audio.Context
	impactSound  *audio.Player
}

// Update is called every frame, it updates the game state
func (g *Game) Update() error {

	// handle the throw command from the twitch chat using a go channel
	// only start a new throw if the projectile is not moving and life > 0
	if !g.projectile.moving && g.life > 0 {
		select {
		case <-g.throwChannel:
			side := rand.Intn(2)
			if side == 0 {
				g.projectile.x = -50
			} else {
				g.projectile.x = screenWidth + 50
			}
			g.projectile.y = float64(rand.Intn(screenHeight))
			g.projectile.moving = true
		default:
		}
	}

	// if the life is greater than 0, move the projectile toward the target, otherwise the pinata is broken and the candy falls out
	if g.life > 0 {
		if g.projectile.moveToward() {
			g.life -= 1
			if g.impactSound != nil {
				g.impactSound.Rewind()
				g.impactSound.Play()
			}
		}
	} else {
		g.projectile.x = -50
		g.projectile.y = -50
		g.projectile.moving = false
		g.pinataBack.moveToward(3*(screenWidth/4), screenHeight+g.pinataBack.height, 15)
		g.candy.moveToward(screenWidth/2, screenHeight+50, 10)
		if g.candy.y >= screenHeight+50 && g.nbCandy < 5 {
			g.candy.y = screenHeight / 2
			g.nbCandy++
		}
	}

	return nil
}

// Draw is called every frame, it draws the game state to the screen
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 0})
	g.candy.Draw(screen)
	g.pinataFront.Draw(screen)
	g.pinataBack.Draw(screen)
	g.projectile.Draw(screen)

	drawLifeBar(screen, g.life)
}

// Layout is called to get the renderer size, not to be confused with the window size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// main function, entry point of the program
// here we init everything and start the game loop
func main() {
	game := &Game{}
	game.life = maxLife
	game.nbCandy = 0
	game.throwChannel = make(chan bool, 100)

	// init the audio context and load the sound effect
	game.audioContext = audio.NewContext(48000)
	soundFile, err := assets.Open("assets/nya.wav")
	if err != nil {
		log.Printf("Failed to load sound: %v", err)
	} else {
		soundData, err := wav.DecodeWithoutResampling(soundFile)
		if err != nil {
			log.Printf("Failed to decode sound: %v", err)
		} else {
			game.impactSound, err = game.audioContext.NewPlayer(soundData)
			if err != nil {
				log.Printf("Failed to create audio player: %v", err)
			}
		}
		soundFile.Close()
	}

	// init the candy
	game.candy.init("assets/candy.png")
	game.candy.x = screenWidth / 2
	game.candy.y = screenHeight / 2
	game.candy.height = 100
	game.candy.width = 100

	// init the pinata front entity
	game.pinataFront.init("assets/front.png")
	game.pinataFront.x = screenWidth / 2
	game.pinataFront.y = screenHeight / 2
	game.pinataFront.height = 720
	game.pinataFront.width = 720

	// init the pinata back entity
	game.pinataBack.init("assets/back.png")
	game.pinataBack.x = screenWidth / 2
	game.pinataBack.y = screenHeight / 2
	game.pinataBack.height = 720
	game.pinataBack.width = 720

	// init the projectile entity
	game.projectile.init("assets/projectile.png")
	game.projectile.x = -50
	game.projectile.y = -50
	game.projectile.height = 100
	game.projectile.width = 100
	game.projectile.speed = 20
	game.projectile.targetX = screenWidth / 2
	game.projectile.targetY = screenHeight / 2
	game.projectile.moving = false

	// init the twitch chat integration
	twitch := &Twitch{
		channel:      channel,
		username:     "justinfan12345",
		password:     "oauth:1234567890",
		throwChannel: game.throwChannel,
	}
	go twitch.Connect()

	// init the game options
	gopt := &ebiten.RunGameOptions{}
	gopt.ScreenTransparent = true
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("lumi pinata")
	if err := ebiten.RunGameWithOptions(game, gopt); err != nil {
		log.Fatal(err)
	}
}

func (e *Entity2D) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	scaleX := e.width / float64(e.image.Bounds().Dx())
	scaleY := e.height / float64(e.image.Bounds().Dy())
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(-e.width/2, -e.height/2)
	op.GeoM.Rotate(e.rotation)
	op.GeoM.Translate(e.x, e.y)
	screen.DrawImage(&e.image, op)
}

// this init an entity from an image file
func (e *Entity2D) init(path string) {
	file, err := assets.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	img, _, err := ebitenutil.NewImageFromReader(file)
	if err != nil {
		log.Fatal(err)
	}
	e.width = float64(img.Bounds().Dx())
	e.height = float64(img.Bounds().Dy())
	e.x = 0
	e.y = 0
	e.rotation = 0
	e.image = *img
}

// this move the entity toward a target
func (e *Entity2D) moveToward(x, y float64, speed float64) {
	dx := x - e.x
	dy := y - e.y
	distance := math.Sqrt(dx*dx + dy*dy)
	if distance > speed {
		e.x += dx * speed / distance
		e.y += dy * speed / distance
	} else {
		e.x = x
		e.y = y
	}
}

// this move the projectile toward the target
func (e *Projectile) moveToward() bool {
	if !e.moving {
		return false
	}
	dx := e.targetX - e.x
	dy := e.targetY - e.y
	distance := math.Sqrt(dx*dx + dy*dy)
	if distance > e.speed {
		e.x += dx * e.speed / distance
		e.y += dy * e.speed / distance
		return false
	} else {
		e.x = -50
		e.y = -50
		e.moving = false
		return true
	}
}

// this draw the life bar
func drawLifeBar(screen *ebiten.Image, life int) {
	// Draw life bar background
	barWidth := 500.0
	barHeight := 30.0
	barX := 20.0
	barY := 20.0

	// Draw background (gray)
	bgRect := ebiten.NewImage(int(barWidth), int(barHeight))
	bgRect.Fill(color.RGBA{100, 100, 100, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(barX, barY)
	screen.DrawImage(bgRect, op)

	// Draw life bar (green)
	lifeWidth := barWidth * float64(life) / maxLife
	if lifeWidth > 0 {
		lifeRect := ebiten.NewImage(int(lifeWidth), int(barHeight))
		lifeRect.Fill(color.RGBA{127, 255, 127, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(barX, barY)
		screen.DrawImage(lifeRect, op)
	}
}

// ------------------------------------------------------------------------------------------------
// twitch chat integration
// we use websocket to allow to run in a browser
//
// ------------------------------------------------------------------------------------------------

type Twitch struct {
	channel      string
	username     string
	password     string
	throwChannel chan bool
}

func (t *Twitch) Connect() {
	// create the websocket connection to twitch websocket server
	ws := js.Global().Get("WebSocket").New("wss://irc-ws.chat.twitch.tv:443")
	onOpen := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Printf("WebSocket connected")
		ws.Call("send", "PASS "+t.password)
		ws.Call("send", "NICK "+t.username)
		ws.Call("send", "JOIN #"+t.channel)
		log.Printf("Connected to Twitch channel: %s", t.channel)
		return nil
	})
	defer onOpen.Release()

	// handle the messages from the twitch chat
	onMessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		message := args[0].Get("data").String()

		if strings.Contains(message, "PING") {
			ws.Call("send", "PONG :tmi.twitch.tv")
			return nil
		}

		if strings.Contains(message, "PRIVMSG") && strings.Contains(strings.ToLower(message), "!throw") {
			log.Printf("Detected !throw command: %s", message)
			select {
			case t.throwChannel <- true:
			default:
			}
		}
		return nil
	})
	defer onMessage.Release()

	// handle the errors from the websocket connection
	onError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Printf("WebSocket error: %v", args[0])
		return nil
	})
	defer onError.Release()

	// handle the closing of the websocket connection
	onClose := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Printf("WebSocket closed")
		return nil
	})
	defer onClose.Release()

	// set the event listeners for the websocket connection
	ws.Set("onopen", onOpen)
	ws.Set("onmessage", onMessage)
	ws.Set("onerror", onError)
	ws.Set("onclose", onClose)

	select {}
}

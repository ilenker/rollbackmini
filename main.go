package main

import (
	"fmt"
	"os"
	"time"
	"sync"
	"math"
	//"unsafe"
	"github.com/gdamore/tcell/v2"
)

// Globals
var (
	ROLLBACK bool
	BREAK    bool
	SYNC     bool

	localScore int
	peerScore  int

	scr tcell.Screen
	err error

	inputFromPeerCh chan PeerPacket
	replyFromPeerCh chan PeerPacket
	packetsToPeerCh chan PeerPacket

	player1 Snake
	player2 Snake

	localPlayerPtr *Snake
	peerPlayerPtr  *Snake

	_simSamples int64

	mu sync.Mutex
	condLighting = sync.NewCond(&mu)

	_graphZoom float64 = 1
	s []*Slider
)

func _main() {
	test()
}


func main() {
	lightPoints = make([]Vec2, 10)
							/*#### INIT ####*/
	FrameDiffBuffer = makeAverageIntBuffer(20)
	player1 = snakeMake(Vec2{(MapW/2) ,  11 + MapH/2}, R, P1Head)
	player2 = snakeMake(Vec2{(MapW/2) , -12 + MapH/2}, R, P2Head)
	loadConfig("config.json")


	if online {
		inputFromPeerCh = make(chan PeerPacket, 128)
		replyFromPeerCh = make(chan PeerPacket, 128)
		packetsToPeerCh = make(chan PeerPacket,  32)
		go multiplayer(inputFromPeerCh, replyFromPeerCh, packetsToPeerCh)
		<-inputFromPeerCh
	}

	stylesInit()
	setCOLORTERM()
	defer restoreCOLORTERM()
	scr, err = tcell.NewScreen()	;F(err, "")
	err = scr.Init()				;F(err, "")
	scr.EnableMouse()

	s = make([]*Slider, 4)
	slidersInit(MapW + 28, 2)

	s[0].Value = γ 
	s[0].Percent = iLerp(int(s[0].Min), int(s[0].Max), γ)

	s[1].Value = ω 
	s[1].Percent = iLerp(int(s[1].Min), int(s[1].Max), ω)

	s[2].Value = β 
	s[2].Percent = iLerp(int(s[2].Min), int(s[2].Max), β)

	boardInit()
	textBoxesInit()

	startCol := newVecRGB(tcell.ColorSteelBlue.RGB())
	rgbOsc := newRGBOscillator(startCol)

	localInputCh := make(chan signal, 8)
	go readLocalInputs(localInputCh)

	frameDiffGraph = newBarGraph(MapW + MapY + 4, 21)

	render(scr, MapX, MapY)

	simTick := time.NewTicker(SIM_TIME)

	mainLightTicker := time.NewTicker(time.Second * 3)


	
/* ············································································· Sync Loop       */
	if online {
		x, y := debugBox("Connecting")
		for {
			<-simTick.C

			x, y = loadingInfo(x, y)

			SIM_FRAME++
			render(scr, MapX, MapY)
			if SIM_FRAME > 600 {
				break
			}
		}
		if LOCAL == 1 {
			// Send Start Signal as player 1
			packetsToPeerCh <-PeerPacket{}
			time.Sleep(time.Duration(
				float64(avgRTTuSec) / float64(2)) * time.Microsecond)
		}
		if LOCAL == 2 {
			// Block here for start signal as player 2
			<-inputFromPeerCh
		}
		SIM_FRAME = START_FRAME
	}


/* ············································································· Main Loop       */
	// qwfp
	for {

		select {
		case <-mainLightTicker.C:
			go flash(mainLightArgs.unpack())
		default:
		}


		condLighting.Broadcast()

		simStart := time.Now()
		<-simTick.C


		if !online { goto SkipRollback }
/* ············································································· Network Inbound */
		select {
		case pPacket := <-inputFromPeerCh:
			if pPacket.frameID < 5 {
				errorBox("skip", 0, 0)
				goto SkipRollback
			}

			// Case of "reporting no inputs"
			if pPacket.content[0] == iNone ||
			pPacket.content[0] == 0 {
				goto SkipRollback
			}

/* ·····································································┬·············· Rollback  ·
·                                                                       └──Net Out - Hit Confirm */
			prePos := getPeerPlayerPtr().pos
			ROLLBACK = true
			rollbackBuffer.resimFramesWithNewInputs(pPacket)
			ROLLBACK = false
			postPos := getPeerPlayerPtr().pos

			col := getPeerPlayerPtr().stateID
			dir := Vec2{-1, 0}
			switch getPeerPlayerPtr().dir {
			case R:
				dir = Vec2{1, 0}
			}
			dist := AbsInt(prePos.x - postPos.x)

			go rollbackStreak(prePos, dist, dir, colorID(col))

		default:
		// Don't block
		}

		SkipRollback:

/* ································································┬···········Stage Local Input  ·
·                                                                  └──Net Out - Send Local Input */
		drainLocalInputCh(localInputCh)

/* ············································································· Push Save State */
		rollbackBuffer.pushFrame(copyCurrentFrameData(SIM_FRAME))

/* ···················································································· Simulate */
		variableDisplay()
		simulate()

/* ············································································· Network Inbound  ·
·                                                                                    Hit Confirm */
		select {
		case reply := <-replyFromPeerCh:
			if reply.content[0] == iHit {
				other := getPeerPlayerPtr()
				dir := 1.5
				if other.stateID == P1Head { dir = 0.5 }
				for range 4 { go hitEffect(other.pos, dir, hitCols[other.stateID]) }
				localScore++
			}
			if reply.content[0] == iCrit {
				other := getPeerPlayerPtr()
				dir := 1.5
				if other.stateID == P1Head { dir = 0.5 }
				for range 6 { go hitEffectCrit(other.pos, dir, hitCols[other.stateID]) }
				localScore += 5
			}
		default:
		}

/* ······················································································ Render */
		drawPixelBox(scr, 2, 2, MapW - 1, MapH/2 - 1, rgbOsc())

		scoreBox(fmt.Sprintf("[%02d]:[%02d]", localScore, peerScore), 0, 0)
		setColor(13,1, cols[getLocalPlayerPtr().stateID])
		setColor(14,1, cols[getLocalPlayerPtr().stateID])
		setColor(18,1, cols[getPeerPlayerPtr().stateID])
		setColor(19,1, cols[getPeerPlayerPtr().stateID])

		frameBox(fmt.Sprintf(" [%05d] ", SIM_FRAME), 0, 0)

		_simSamples += int64(float64((time.Duration(SIM_TIME) - time.Since(simStart)).Microseconds()) * float64(_graphZoom/1000))
		//errorBox(fmt.Sprintf("zoom: %f", _graphZoom), 0, 0)

		if SIM_FRAME % 9 == 0 {
			frameDiffGraph(
				int(_simSamples / 9) + 10 - int(_graphZoom * 2),
				)
			_simSamples = 0

		}

		//eB(fmt.Sprintf("tc.col:%d", unsafe.Sizeof(tcell.ColorBlack)), 0, 0)
		//(fmt.Sprintf("vecRGB:%d", unsafe.Sizeof(VecRGB{})), 0, 1)
		//(fmt.Sprintf("board :%d", unsafe.Sizeof(board)), 0, 2)
		//(fmt.Sprintf("vfxLay:%d", unsafe.Sizeof(vfxLayer)), 0, 3)
		//(fmt.Sprintf("ligLay:%d", unsafe.Sizeof(lightLayer)), 0, 4)
		sliders()

		γ = s[0].Value
		ω = s[1].Value
		
		render(scr, MapX, MapY)

		SIM_FRAME++
	}

}


func Break() {
	
	BREAK = true
	for {
		ev := scr.PollEvent()
		if key, ok := ev.(*tcell.EventKey); ok {

			if key.Key() == tcell.KeyESC {
				scr.Fini()
				os.Exit(0)
			} 

			// Keymap
			switch key.Rune() {

			case 'p':
				BREAK = false
				return

			case '1':
				variablePage = 1
			case '2':
				variablePage = 2
			case '3':
				variablePage = 3
			case '4':
				variablePage = 4
			}

			variableDisplay()
			render(scr, MapX, MapY)
		} 
	}

}


// Collect all (local) input and send down a single channel
func readLocalInputs(inputCh chan signal) {
	for {
		//if SIM_FRAME < START_FRAME + RB_BUFFER_LEN * 3 {
		//	continue
		//}
		ev := scr.PollEvent()

		if mev, ok := ev.(*tcell.EventMouse); ok {
			for _, slider := range s {
				slider.HandleEvent(mev)
			}
		}

		if key, ok := ev.(*tcell.EventKey); ok {

			if key.Key() == tcell.KeyESC {
				scr.Fini()
				os.Exit(0)
			} 

			// Special Keys
			switch key.Key() {
			case tcell.KeyLeft:
				inputCh <-iLeft
			case tcell.KeyRight:
				inputCh <-iRight
			default:
			}

			// Keymap
			switch key.Rune() {
			case 'p':
				for range 6 { go hitEffectCrit(player1.pos, 1.5, hitCols[player2.stateID]) }

			case 'b':
				for range 4 { go hitEffect(player1.pos, 1.5, hitCols[player2.stateID]) }

			case 'r':
				
			case 'x':
				inputCh <-iLeft
			case 'd':
				inputCh <-iRight
			case ' ':
				inputCh <-iShot

			case '1':
				_graphZoom--

			case '2':
				_graphZoom++

			case '!':
				variablePage = 1
			case '@':
				variablePage = 2
			case '#':
				variablePage = 3
			case '$':
				variablePage = 4
			case '%':
				variablePage = 5
			case '^':
				variablePage = 6
			case '0':
				callsBox.Clear()
				errorBox.Clear()
			}

		} 
	}

}


func drainLocalInputCh(inputCh chan signal) {

	player := getLocalPlayerPtr()

	select {
	case input := <-inputCh:
		if player.shotCD == 0 {
			sendCurrentFrameInputs(input)
			player.tryInput(input)
		}
		return
	default:
		sendCurrentFrameInputs(iNone)
		return
	}

}


func getLocalPlayerCopy() Snake {
	if LOCAL == 2 {
		return player2
	}
	return player1
}

func getPeerPlayerCopy() Snake {
	if LOCAL == 2 {
		return player1
	}
	return player2
}

func getLocalPlayerPtr() *Snake {
	if LOCAL == 2 {
		return &player2
	}
	return &player1
}

func getPeerPlayerPtr() *Snake {
	if LOCAL == 2 {
		return &player1
	}
	return &player2
}

func raycast(p1 Vec2, dir float64) {

	p2 := p1.translate(dir, 20)

	f := 0.0

	for {
		if f > 2 {
			return
		}
		x := math.Round(lerp(p1.x, p2.x, f))
		y := math.Round(lerp(p1.y, p2.y, f))

		if x > MapW || x < 0 {
			return
		}

		if y > MapH || y < 0 {
			return
		}

		f += 0.001
	}
}

func angleBetween(a, b Vec2) float64 {
	ax := float64(a.x)
	ay := float64(a.y)
	bx := float64(b.x)
	by := float64(b.y)

	dot := ax * bx + ay * by
	mag := math.Hypot(ax, ay) * math.Hypot(bx, by)

	if mag == 0 {
		return 0
	}
	return math.Acos(dot / mag) // radians
}

func angleTo(a, b Vec2) float64 {
	d := Vec2{b.x - a.x, b.y - a.y}
	return math.Atan2(float64(d.y), float64(d.x)) // radians, from -π to +π
}

package main

import (
	"fmt"
	"os"
	"time"
	"github.com/gdamore/tcell/v2"
)

// Globals
var ROLLBACK bool = false
var BREAK    bool = false
var SYNC     bool = false

var localScore int = 0
var peerScore  int = 0

var scr tcell.Screen
var err error

var inputFromPeerCh chan PeerPacket
var replyFromPeerCh chan PeerPacket
var packetsToPeerCh chan PeerPacket

var player1 Snake
var player2 Snake

var localPlayerPtr *Snake
var peerPlayerPtr  *Snake

var _simSamples int64 = 0

func main() {
							/*#### INIT ####*/
	FrameDiffBuffer = makeAverageIntBuffer(20)
	player1 = snakeMake(Vec2{(MapW/2) ,  6 + MapH/2}, R, P1Head)
	player2 = snakeMake(Vec2{(MapW/2) , -7 + MapH/2}, R, P2Head)
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
	//scr.SetStyle(ColEmpty)

	boardInit()
	textBoxesInit()

	startCol := newVecRGB(tcell.ColorSteelBlue.RGB())
	rgbOsc := newRGBOscillator(startCol)

	localInputCh := make(chan signal, 8)
	go readLocalInputs(localInputCh)

	frameDiffGraph = newBarGraph(2, 19)

	render(scr, MapX, MapY)

	simTick := time.NewTicker(SIM_TIME)

	
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
		setColor(18,1, cols[getLocalPlayerPtr().stateID])
		setColor(19,1, cols[getLocalPlayerPtr().stateID])
		setColor(23,1, cols[getPeerPlayerPtr().stateID])
		setColor(24,1, cols[getPeerPlayerPtr().stateID])

		frameBox(fmt.Sprintf(" [%05d] ", SIM_FRAME), 0, 0)

		_simSamples += (time.Duration(SIM_TIME) - time.Since(simStart)).Microseconds() / 250

		if SIM_FRAME % 5 == 0 {
			frameDiffGraph(
				int(_simSamples / 5) + 6,
				)
			_simSamples = 0
		}

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
			case 'x':
				inputCh <-iLeft
			case 'd':
				inputCh <-iRight
			case ' ':
				inputCh <-iShot


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

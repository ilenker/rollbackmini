package main

import "fmt"

type Snake struct {
	pos Vec2
	dir direction
	shootDir Vec2

	scpt int16
	subcellDebt int16

	inputBuffer [4]signal
	inputIndex int8

	stateID cellState
	isLocal bool
	shooting bool

	shotCD int
}


func (s *Snake) popInput() signal {

	if s.inputIndex < 0 {
		return iNone
	}

	input := s.inputBuffer[s.inputIndex]

	s.inputBuffer[s.inputIndex] = iNone

	s.inputIndex--

	return input
} 


func (s *Snake) tryInput(input signal) bool {
	if s.inputIndex + 1 >= INPUT_BUFFER_LEN {
		return false
	}
	if input == iNone || input == 0 {
		return false
	}
	if input == iRight && s.dir == R {
		return false
	}
	if input == iLeft && s.dir == L {
		return false
	}
	
	s.inputIndex++
	s.inputBuffer[s.inputIndex] = input
	return true
}


func (s *Snake) move() {
	if s.dir == R {
		//s.pos.x = wrapInt(s.pos.x + 1, MapW)
		if s.pos.x == MapW - 1 {
			s.dir = L
			s.move()
			return
		}
		s.pos.x++
	}
	if s.dir == L {
		//s.pos.x = wrapInt(s.pos.x - 1, MapW)
		if s.pos.x == 0 {
			s.dir = R
			s.move()
			return
		}
		s.pos.x--
	}
}


func (s *Snake) shoot() {
	if !s.shooting {
		return
	}
	s.shooting = false

	distance := 32

	other := player2
	if s.stateID == P2Head {
		other = player1
	}

	callsBox(fmt.Sprintf("A:%2d|B:%2d, rb:%v\n", s.pos.x, other.pos.x, ROLLBACK))
	dx := AbsInt(other.pos.x - s.pos.x)
	if dx < 2 {
		distance = AbsInt(other.pos.y - s.pos.y) - 1

		if !s.isLocal {
			dir := 1.5
			if other.stateID == P1Head { dir = 0.5 }

			if other.shotCD != 0 {
				peerScore++
				packetsToPeerCh <-PeerPacket{
					0, [4]signal{iHit},
				}

				for range 4 { go hitEffect(other.pos, dir, hitCols[other.stateID]) }

			} else {
				peerScore += 5
				packetsToPeerCh <-PeerPacket{
					0, [4]signal{iCrit},
				}

				for range 5 { go hitEffectCrit(other.pos, dir, hitCols[other.stateID]) }

			}
			
		}


	} else {
		if ROLLBACK && !s.isLocal {
			packetsToPeerCh <-PeerPacket{
				0, [4]signal{iMiss},
			}
		}

	}
	go beamEffect(
		s.pos.add(Vec2{-1, 0}).add(s.shootDir.scale(2)),
		distance-1,
		s.shootDir,
		beamCols[s.stateID],
		)

	go beamEffect(
		s.pos,
		distance+2,
		s.shootDir,
		beamCols[s.stateID],
		)

	go beamEffect(
		s.pos.add(Vec2{ 1, 0}).add(s.shootDir.scale(2)),
		distance-1,
		s.shootDir,
		beamCols[s.stateID],
		)
}


func (s *Snake) control() {

	input := s.popInput()
	//callsBox(fmt.Sprintf("popInput()->%c\n", input))

	switch input {
	case iRight:
		s.dir = R

	case iLeft:
		s.dir = L

	case iShot:
		s.shotCD = 60
		s.shooting = true

	}
}

func (s *Snake) cooldown() {
	col := colorID(P1Head)
	pos := Vec2{MapX + MapW/2 - 8, MapH - 2}

	if s.stateID == P2Head {
		col = colorID(P2Head)
		pos = Vec2{MapX + MapW/2 - 12, 1}
	}

	length := int(iLerp(0, 60, float64(s.shotCD)) * 16)
	cooldownBar(pos, length, col)

	if s.shotCD > 0 {
		if s.scpt != 2 {s.scpt = 2}
		s.shotCD--
		return
	}

	if s.scpt != 8 {
		s.scpt = 8
	}

}

func snakeMake(start Vec2, d direction, stateID cellState) Snake{

	isLocal := true
	if LOCAL == 2 {
		isLocal = false
	}

	shootDir := Vec2{0, -1}
	if stateID == P2Head {
		shootDir = Vec2{0, 1}
	}

	inputBuffer := [4]signal{'_', '_', '_', '_',}

	snake := Snake{
		pos: start,
		dir: d,
		subcellDebt: 0,
		inputBuffer: inputBuffer,
		inputIndex: -1,
		stateID: stateID,
		shootDir: shootDir,
		shooting: false,
		isLocal: isLocal,
		shotCD: 0,
	}

	return snake
}

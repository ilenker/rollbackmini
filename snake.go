package main

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
}


func (s *Snake) popInput() signal {

	if s.inputIndex < 0 {
		return iNone
	}

	input := s.inputBuffer[s.inputIndex]

	s.inputBuffer[s.inputIndex] = '_'

	s.inputIndex--

	return input
} 


func (s *Snake) tryInput(input signal) bool {
	if s.inputIndex + 1 >= INPUT_BUFFER_LEN {
		return true
	}
	if input == iNone || input == 0 {
		return true
	}
	
	s.inputIndex++
	s.inputBuffer[s.inputIndex] = input
	return false
}


func (s *Snake) move() {
	if s.dir == R {
		s.pos.x = wrapInt(s.pos.x + 1, MapW)
	}
	if s.dir == L {
		s.pos.x = wrapInt(s.pos.x - 1, MapW)
	}
}


func (s *Snake) shoot() {
	if !s.shooting {
		return
	}
	distance := 20

	other := player2
	if s.stateID == P2Head {
		other = player1
	}

	if other.pos.x == s.pos.x {
		distance = AbsInt(other.pos.y - s.pos.y) - 1

		if ROLLBACK && !s.isLocal {
			//packetsToPeerCh <- PeerPacket{
				//0, [4]signal{'H'},
			//}
			//go beamEffect(s.pos, distance, s.shootDir, beamCols[s.stateID])
		}

	} 
	go beamEffect(s.pos, distance, s.shootDir, beamCols[s.stateID])
	s.shooting = false
}


func (s *Snake) control() {

	input := s.popInput()

	switch input {
	case iRight:
		s.dir = R

	case iLeft:
		s.dir = L

	case iShot:
		s.shooting = true
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
	}

	return snake
}

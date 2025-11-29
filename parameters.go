package main

type lightFlashArgs struct {
	Location   Vec2
    Size       int
    Luminosity float64
    Tint       Vec3[float64]
}

var mainLightArgs = lightFlashArgs{
	Location  : Vec2{MapW/2, MapH/2},
	Size      : 40,
	Luminosity: 3,
	Tint      : Vec3[float64]{1, 0.7, 0.7},
}

const size     = 25
const lum      = 3

var ω = 255.0     // Whitepoint
var β = 100.0     // Blackpoint
var γ = .6        // Gamma
var γRec = 1/γ    // Gamma

var FOFactor = 0.03

func (l *lightFlashArgs) unpack() (Vec2, int, float64, Vec3[float64]){
	return l.Location, l.Size, l.Luminosity, l.Tint
}


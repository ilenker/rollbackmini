package main

type lightFlashArgs struct {
	Location   Vec2
    Size       int
    Luminosity float64
    Linger     float64
    Tint       Vec3[float64]
    CastShadow bool
}

var mainLightArgs = lightFlashArgs{
	Location  : Vec2{MapW/2, MapH/2},
	Size      : 90,
	Luminosity: 100,
	Linger    : 0.05,
	Tint      : Vec3[float64]{1, 0.3, 0.2},
	CastShadow: true,
}

var ω = 255.0 // Whitepoint
var β = 100.0 // Blackpoint
var γ = .5    // Gamma

const size     = 50
const lum      = 50.0
const linger   = 0.09
const FOFactor = 0.1

func (l *lightFlashArgs) unpack() (Vec2, int, float64, float64, Vec3[float64], bool){
	return l.Location, l.Size, l.Luminosity, l.Linger, l.Tint, l.CastShadow
}


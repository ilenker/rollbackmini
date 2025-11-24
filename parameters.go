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
	Size      : 500,
	Luminosity: 100,
	Linger    : 4,
	Tint      : Vec3[float64]{1, .8, .8},
	CastShadow: true,
}

const gamma = 2.0
const size = 100
const lum = 10.0
const linger = 2.0
const FOFactor = 1

func (l *lightFlashArgs) unpack() (Vec2, int, float64, float64, Vec3[float64], bool){
	return l.Location, l.Size, l.Luminosity, l.Linger, l.Tint, l.CastShadow
}


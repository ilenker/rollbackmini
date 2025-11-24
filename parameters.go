package main

type lightFlashArgs struct {
	Location   Vec2
    Size       int
    Luminosity float32
    Linger     int
    Tint       Vec3[float32]
    CastShadow bool
}

var mainLightArgs = lightFlashArgs{
	Location  : Vec2{MapW/2, MapH/2},
	Size      : 400,
	Luminosity: 100,
	Linger    : 15,
	Tint      : Vec3[float32]{1, .8, .8},
	CastShadow: true,
}

var hitLightArgs = lightFlashArgs{
	Location  : Vec2{MapW/2, MapH/2},
	Size      : 400,
	Luminosity: 100,
	Linger    : 15,
	Tint      : Vec3[float32]{1, .8, .8},
	CastShadow: true,
}

func (l *lightFlashArgs) unpack() (Vec2, int, float32, int, Vec3[float32], bool){
	return l.Location, l.Size, l.Luminosity, l.Linger, l.Tint, l.CastShadow
}


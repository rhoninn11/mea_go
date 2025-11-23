package internal

import (
	"image"
	"image/color"

	mea_gen_d "mea_go/src/api/mea.gen.d"
)

func ImgProtoToGo(protoImg *mea_gen_d.Image) *image.RGBA {
	w := int(protoImg.Info.Width)
	h := int(protoImg.Info.Height)
	total := w * h
	goImg := image.NewRGBA(image.Rect(0, 0, w, h))

	idxCalc := func(idx int) (int, int) {
		y := idx / w
		x := idx - w*y
		return x, y
	}
	var y int
	var x int
	var pixel []byte
	for i := range total {
		rgb_idx := i * 3
		pixel = protoImg.Pixels[rgb_idx : rgb_idx+3]
		c := color.RGBA{
			R: pixel[0],
			G: pixel[1],
			B: pixel[2],
			A: 255,
		}
		x, y = idxCalc(i)
		goImg.SetRGBA(x, y, c)
	}
	return goImg
}

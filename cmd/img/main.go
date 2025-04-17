package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

func addLabel(img *image.RGBA) {
	x := 100
	y := 10
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: inconsolata.Bold8x16,
		Dot:  point,
	}

	drawer.DrawString("jem chleb")
}

func main() {
	loSize := 256
	blanc := image.NewRGBA(image.Rect(0, 0, loSize, loSize))
	for x := 0; x < loSize; x++ {
		for y := 0; y < loSize; y++ {
			blanc.Set(x, y, color.White)
		}
	}

	hiSize := 1024
	num := 4
	canvas := image.NewRGBA(image.Rect(0, 0, hiSize*num, hiSize))
	for i := 0; i < num; i++ {
		prog := image.NewRGBA(image.Rect(0, 0, hiSize, hiSize))
		draw.Draw(prog, prog.Rect, blanc, image.Point{X: -512, Y: -i * 128}, draw.Over)
		addLabel(prog)
		draw.Draw(canvas, canvas.Rect, prog, image.Point{X: -hiSize * i, Y: 0}, draw.Over)
	}

	f, err := os.Create("tmp/obraz.png")
	if err != nil {
		log.Fatal(err.Error())
	}

	err = png.Encode(f, canvas)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("imge saved")
}

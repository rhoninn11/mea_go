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

const imgSize = 1024
const num = 4
const rows = 1

// wystrtuj kanwas w oparciu o stałe
func spawnCanvas() *image.RGBA {
	xSpace := imgSize * num * rows
	ySpace := imgSize
	canvas := image.NewRGBA(image.Rect(0, 0, xSpace, ySpace))
	red := color.RGBA{0xff, 0x00, 0x00, 0x00}
	// color.Black;
	uniMask := image.NewUniform(red)
	draw.Draw(canvas, canvas.Rect, uniMask, slot('a'), draw.Over)
	return canvas
}

const slots = "abcd"

// wybierz slota z pamięci graficznej
func slot(hmm rune) image.Point {
	var idx = 0
	for i, r := range slots {
		if r == hmm {
			idx = i
		}
	}
	return image.Point{X: -imgSize * idx, Y: 0}
}

const lbX = 100
const lbY = 32

// dodaj obrazek do zdjęcia
func addLabel(img *image.RGBA, pnt fixed.Point26_6) {
	const myStr = "jem chleb"
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: inconsolata.Bold8x16,
		Dot:  pnt,
	}
	drawer.DrawString(myStr)

	drawer.Src = image.White
	drawer.Face = inconsolata.Regular8x16
	drawer.Dot = pnt
	drawer.DrawString(myStr)
}

//TODO: mogę dodać tu czyszczenie tła

// otwórz parę obrazków dodaj do nich napisy
func main() {
	loSize := 256
	blanc := image.NewRGBA(image.Rect(0, 0, loSize, loSize))
	for x := 0; x < loSize; x++ {
		for y := 0; y < loSize; y++ {
			blanc.Set(x, y, color.White)
		}
	}
	subSlots := slots[0:num]
	canvas := spawnCanvas()
	stencil := image.NewRGBA(image.Rect(0, 0, imgSize, imgSize))

	for _, slt := range subSlots {
		// draw.Draw(sample, sample.Rect, blanc, slot(slt), draw.Over)
		// somehow need to clear
		// stencil.clear()
		addLabel(stencil, fixed.P(lbX, lbY))
		draw.Draw(canvas, canvas.Rect, stencil, slot(slt), draw.Over)
		// _ = slt
	}
	if err := savePng(canvas, "tmp/out.png"); err != nil {
		log.Fatal(err)
	}
}

// save img data as png image
func savePng(img *image.RGBA, path string) error {
	f, err := os.Create("tmp/obraz.png")
	if err != nil {
		return fmt.Errorf("creat failed for %s, %w", path, err)
	}
	err = png.Encode(f, img)
	if err != nil {
		return fmt.Errorf("enode faile for %s, %w", path, err)
	}
	fmt.Println("image saved")
	return nil
}

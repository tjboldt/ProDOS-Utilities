// Copyright Terence J. Boldt (c)2021-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to read, write and delete
// files on a ProDOS drive image

package prodos

import (
	"bytes"
	"fmt"
	"image"
	"image/color"

	// force import jpeg support by init only
	_ "image/jpeg"
	// force import png support by init only
	_ "image/png"

	"golang.org/x/image/draw"
)

// 10  PRINT  CHR$ (4)"open offsets"
// 20  PRINT  CHR$ (4)"write offsets"
// 30  PRINT "offsets := [192]int{";
// 40  FOR Y = 0 TO 191
// 50  HPLOT 0,Y
// 60  PRINT ( PEEK (39) * 256 +  PEEK (38)) - 8192;
// 70  IF Y < 191 THEN  PRINT ", ";
// 80  NEXT
// 90  PRINT "}"
// 100  PRINT  CHR$ (4)"close offsets"
var offsets = []int{0, 1024, 2048, 3072, 4096, 5120, 6144, 7168, 128, 1152, 2176, 3200, 4224, 5248, 6272, 7296, 256, 1280, 2304, 3328, 4352, 5376, 6400, 7424, 384, 1408, 2432, 3456, 4480, 5504, 6528, 7552, 512, 1536, 2560, 3584, 4608, 5632, 6656, 7680, 640, 1664, 2688, 3712, 4736, 5760, 6784, 7808, 768, 1792, 2816, 3840, 4864, 5888, 6912, 7936, 896, 1920, 2944, 3968, 4992, 6016, 7040, 8064, 40, 1064, 2088, 3112, 4136, 5160, 6184, 7208, 168, 1192, 2216, 3240, 4264, 5288, 6312, 7336, 296, 1320, 2344, 3368, 4392, 5416, 6440, 7464, 424, 1448, 2472, 3496, 4520, 5544, 6568, 7592, 552, 1576, 2600, 3624, 4648, 5672, 6696, 7720, 680, 1704, 2728, 3752, 4776, 5800, 6824, 7848, 808, 1832, 2856, 3880, 4904, 5928, 6952, 7976, 936, 1960, 2984, 4008, 5032, 6056, 7080, 8104, 80, 1104, 2128, 3152, 4176, 5200, 6224, 7248, 208, 1232, 2256, 3280, 4304, 5328, 6352, 7376, 336, 1360, 2384, 3408, 4432, 5456, 6480, 7504, 464, 1488, 2512, 3536, 4560, 5584, 6608, 7632, 592, 1616, 2640, 3664, 4688, 5712, 6736, 7760, 720, 1744, 2768, 3792, 4816, 5840, 6864, 7888, 848, 1872, 2896, 3920, 4944, 5968, 6992, 8016, 976, 2000, 3024, 4048, 5072, 6096, 7120, 8144}
var pixel = []byte{1, 2, 4, 8, 16, 32, 64}

// ConvertImageToHiResMonochrome converts jpeg and png images to Apple II hi-res monochrome
func ConvertImageToHiResMonochrome(imageBytes []byte) []byte {

	img, _, err := image.Decode(bytes.NewReader(imageBytes))

	if err != nil {
		fmt.Printf("%s\n", err)
		return nil
	}

	a2imgSize := image.Rect(0, 0, 280, 192)

	a2img := image.NewPaletted(a2imgSize, []color.Color{
		color.Black,
		color.White,
	})

	a2monoImgSize := image.Rect(0, 0, 280, 192)

	scaledImg := image.NewRGBA(a2monoImgSize)
	draw.BiLinear.Scale(scaledImg, a2monoImgSize, img, img.Bounds(), draw.Over, nil)
	draw.FloydSteinberg.Draw(a2img, a2monoImgSize, scaledImg, image.Point{})

	hires := make([]byte, 8192)

	for y := a2img.Bounds().Min.Y; y < a2img.Bounds().Max.Y; y++ {
		for x := a2img.Bounds().Min.X; x < a2img.Bounds().Max.X; x++ {
			if a2img.At(x, y) == color.White {
				hires[offsets[y]+x/7] |= pixel[x%7] | 128
			}
		}
	}

	return hires
}

// ConvertImageToHiResColour converts jpeg and png images to Apple II hi-res colour
func ConvertImageToHiResColour(imageBytes []byte) []byte {

	img, _, err := image.Decode(bytes.NewReader(imageBytes))

	if err != nil {
		fmt.Printf("%s\n", err)
		return nil
	}

	a2colourImgSize := image.Rect(0, 0, 140, 192)

	black := color.NRGBA{0, 0, 0, 255}
	green := color.NRGBA{20, 245, 60, 255}
	purple := color.NRGBA{255, 68, 253, 255}
	white := color.NRGBA{255, 255, 255, 255}
	orange := color.NRGBA{255, 106, 60, 255}
	blue := color.NRGBA{20, 207, 253, 255}

	a2img := image.NewPaletted(a2colourImgSize, []color.Color{
		black,
		green,
		purple,
		white,
		orange,
		blue,
	})

	scaledImg := image.NewRGBA(a2colourImgSize)
	draw.BiLinear.Scale(scaledImg, a2colourImgSize, img, img.Bounds(), draw.Over, nil)
	draw.FloydSteinberg.Draw(a2img, a2colourImgSize, scaledImg, image.Point{})

	hires := make([]byte, 8192)

	for y := a2img.Bounds().Min.Y; y < a2img.Bounds().Max.Y; y++ {
		for x7 := a2img.Bounds().Min.X; x7 < a2img.Bounds().Max.X; x7 += 7 {
			switch a2img.At(x7, y) {
			case green:
				hires[offsets[y]+x7*2/7] = 2
			case purple:
				hires[offsets[y]+x7*2/7] = 1
			case orange:
				hires[offsets[y]+x7*2/7] = 2
				hires[offsets[y]+x7*2/7] |= 0x80
			case blue:
				hires[offsets[y]+x7*2/7] = 1
				hires[offsets[y]+x7*2/7] |= 0x80
			case white:
				hires[offsets[y]+x7*2/7] = 3
			}
			switch a2img.At(x7+1, y) {
			case green:
				hires[offsets[y]+x7*2/7] |= 8
				hires[offsets[y]+x7*2/7] &= 0x7F
			case purple:
				hires[offsets[y]+x7*2/7] |= 4
				hires[offsets[y]+x7*2/7] &= 0x7F
			case orange:
				hires[offsets[y]+x7*2/7] |= 8
				hires[offsets[y]+x7*2/7] |= 0x80
			case blue:
				hires[offsets[y]+x7*2/7] |= 4
				hires[offsets[y]+x7*2/7] |= 0x80
			case white:
				hires[offsets[y]+x7*2/7] |= 12
			}
			switch a2img.At(x7+2, y) {
			case green:
				hires[offsets[y]+x7*2/7] |= 32
				hires[offsets[y]+x7*2/7] &= 0x7F
			case purple:
				hires[offsets[y]+x7*2/7] |= 16
				hires[offsets[y]+x7*2/7] &= 0x7F
			case orange:
				hires[offsets[y]+x7*2/7] |= 32
				hires[offsets[y]+x7*2/7] |= 0x80
			case blue:
				hires[offsets[y]+x7*2/7] |= 16
				hires[offsets[y]+x7*2/7] |= 0x80
			case white:
				hires[offsets[y]+x7*2/7] |= 48
			}
			switch a2img.At(x7+3, y) {
			case green:
				hires[offsets[y]+x7*2/7+1] |= 1
				hires[offsets[y]+x7*2/7+1] &= 0x7F
			case purple:
				hires[offsets[y]+x7*2/7] |= 64
				hires[offsets[y]+x7*2/7] &= 0x7F
			case orange:
				hires[offsets[y]+x7*2/7+1] |= 1
				hires[offsets[y]+x7*2/7+1] |= 0x80
			case blue:
				hires[offsets[y]+x7*2/7] |= 64
				hires[offsets[y]+x7*2/7] |= 0x80
			case white:
				hires[offsets[y]+x7*2/7] |= 64
				hires[offsets[y]+x7*2/7+1] |= 1
			}
			switch a2img.At(x7+4, y) {
			case green:
				hires[offsets[y]+x7*2/7+1] |= 4
				hires[offsets[y]+x7*2/7+1] &= 0x7F
			case purple:
				hires[offsets[y]+x7*2/7+1] |= 2
				hires[offsets[y]+x7*2/7+1] &= 0x7F
			case orange:
				hires[offsets[y]+x7*2/7+1] |= 4
				hires[offsets[y]+x7*2/7+1] |= 0x80
			case blue:
				hires[offsets[y]+x7*2/7+1] |= 2
				hires[offsets[y]+x7*2/7+1] |= 0x80
			case white:
				hires[offsets[y]+x7*2/7+1] |= 6
			}
			switch a2img.At(x7+5, y) {
			case green:
				hires[offsets[y]+x7*2/7+1] |= 16
				hires[offsets[y]+x7*2/7+1] &= 0x7F
			case purple:
				hires[offsets[y]+x7*2/7+1] |= 8
				hires[offsets[y]+x7*2/7+1] &= 0x7F
			case orange:
				hires[offsets[y]+x7*2/7+1] |= 16
				hires[offsets[y]+x7*2/7+1] |= 0x80
			case blue:
				hires[offsets[y]+x7*2/7+1] |= 8
				hires[offsets[y]+x7*2/7+1] |= 0x80
			case white:
				hires[offsets[y]+x7*2/7+1] |= 24
			}
			switch a2img.At(x7+6, y) {
			case green:
				hires[offsets[y]+x7*2/7+1] |= 64
				hires[offsets[y]+x7*2/7+1] &= 0x7F
			case purple:
				hires[offsets[y]+x7*2/7+1] |= 32
				hires[offsets[y]+x7*2/7+1] &= 0x7F
			case orange:
				hires[offsets[y]+x7*2/7+1] |= 64
				hires[offsets[y]+x7*2/7+1] |= 0x80
			case blue:
				hires[offsets[y]+x7*2/7+1] |= 32
				hires[offsets[y]+x7*2/7+1] |= 0x80
			case white:
				hires[offsets[y]+x7*2/7+1] |= 96
			}
		}
	}

	return hires
}

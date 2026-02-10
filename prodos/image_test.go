// Copyright Terence J. Boldt (c)2026
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides tests for conversion between Apple II hi-res and standard images

// This file was actually generated entirely by Claude Opus 4.6
// It was part of the image export that was created in about 7 minutes of AI time

package prodos

import (
	"image/color"
	"testing"
)

func TestConvertHiResToMonochromeImage(t *testing.T) {
	t.Run("RejectsTooLarge", func(t *testing.T) {
		_, err := ConvertHiResToMonochromeImage(make([]byte, 8193))
		if err == nil {
			t.Error("expected error for oversized data")
		}
	})

	t.Run("AcceptsShorterData", func(t *testing.T) {
		hires := make([]byte, 8184)
		img, err := ConvertHiResToMonochromeImage(hires)
		if err != nil {
			t.Fatalf("unexpected error for 8184 bytes: %s", err)
		}
		r, g, b, _ := img.At(0, 0).RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("expected black pixel, got r=%d g=%d b=%d", r, g, b)
		}
	})

	t.Run("AllBlack", func(t *testing.T) {
		hires := make([]byte, 8192)
		img, err := ConvertHiResToMonochromeImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		r, g, b, _ := img.At(0, 0).RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("expected black pixel, got r=%d g=%d b=%d", r, g, b)
		}
	})

	t.Run("SinglePixel", func(t *testing.T) {
		hires := make([]byte, 8192)
		// Set pixel at (0,0): byte at offsets[0]=0, bit 0 (pixel[0]=1), with high bit
		hires[0] = 0x81
		img, err := ConvertHiResToMonochromeImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		r, _, _, _ := img.At(0, 0).RGBA()
		if r == 0 {
			t.Error("expected white pixel at (0,0)")
		}
		r, _, _, _ = img.At(1, 0).RGBA()
		if r != 0 {
			t.Error("expected black pixel at (1,0)")
		}
	})
}

func TestConvertHiResToColourImage(t *testing.T) {
	t.Run("RejectsTooLarge", func(t *testing.T) {
		_, err := ConvertHiResToColourImage(make([]byte, 8193))
		if err == nil {
			t.Error("expected error for oversized data")
		}
	})

	t.Run("AllBlack", func(t *testing.T) {
		hires := make([]byte, 8192)
		img, err := ConvertHiResToColourImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		r, g, b, _ := img.At(0, 0).RGBA()
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("expected black pixel, got r=%d g=%d b=%d", r, g, b)
		}
	})

	t.Run("WhitePixelPair", func(t *testing.T) {
		hires := make([]byte, 8192)
		// Set bits 0 and 1 at byte 0 (columns 0,1) -> white
		hires[0] = 0x03
		img, err := ConvertHiResToColourImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		r, g, b, _ := img.At(0, 0).RGBA()
		if r != 0xFFFF || g != 0xFFFF || b != 0xFFFF {
			t.Errorf("expected white pixel, got r=%04X g=%04X b=%04X", r, g, b)
		}
	})

	t.Run("PurplePixel", func(t *testing.T) {
		hires := make([]byte, 8192)
		// Set bit 0 only (even column only, high bit 0) -> purple
		hires[0] = 0x01
		img, err := ConvertHiResToColourImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		expected := color.NRGBA{255, 68, 253, 255}
		got := img.At(0, 0)
		if got != expected {
			t.Errorf("expected purple %v, got %v", expected, got)
		}
	})

	t.Run("GreenPixel", func(t *testing.T) {
		hires := make([]byte, 8192)
		// Set bit 1 only (odd column only, high bit 0) -> green
		hires[0] = 0x02
		img, err := ConvertHiResToColourImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		expected := color.NRGBA{20, 245, 60, 255}
		got := img.At(0, 0)
		if got != expected {
			t.Errorf("expected green %v, got %v", expected, got)
		}
	})

	t.Run("BluePixel", func(t *testing.T) {
		hires := make([]byte, 8192)
		// Set bit 0 only with high bit (even column, high bit 1) -> blue
		hires[0] = 0x81
		img, err := ConvertHiResToColourImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		expected := color.NRGBA{20, 207, 253, 255}
		got := img.At(0, 0)
		if got != expected {
			t.Errorf("expected blue %v, got %v", expected, got)
		}
	})

	t.Run("OrangePixel", func(t *testing.T) {
		hires := make([]byte, 8192)
		// Set bit 1 only with high bit (odd column, high bit 1) -> orange
		hires[0] = 0x82
		img, err := ConvertHiResToColourImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		expected := color.NRGBA{255, 106, 60, 255}
		got := img.At(0, 0)
		if got != expected {
			t.Errorf("expected orange %v, got %v", expected, got)
		}
	})
}

func TestConvertHiResToCRTImage(t *testing.T) {
	t.Run("RejectsTooLarge", func(t *testing.T) {
		_, err := ConvertHiResToCRTImage(make([]byte, 8193))
		if err == nil {
			t.Error("expected error for oversized data")
		}
	})

	t.Run("OutputSize", func(t *testing.T) {
		hires := make([]byte, 8192)
		img, err := ConvertHiResToCRTImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		bounds := img.Bounds()
		if bounds.Dx() != 1120 || bounds.Dy() != 768 {
			t.Errorf("expected 1120x768, got %dx%d", bounds.Dx(), bounds.Dy())
		}
	})

	t.Run("ScanLinesDarker", func(t *testing.T) {
		hires := make([]byte, 8192)
		// Fill with white pixels
		for i := range hires {
			hires[i] = 0xFF
		}
		img, err := ConvertHiResToCRTImage(hires)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		// Row 1 (full brightness) should be brighter than row 3 (scan line)
		_, g1, _, _ := img.At(560, 1).RGBA()
		_, g3, _, _ := img.At(560, 3).RGBA()
		if g3 >= g1 {
			t.Errorf("scan line row should be darker: row1 g=%d, row3 g=%d", g1, g3)
		}
	})
}

func TestMonochromeRoundTrip(t *testing.T) {
	// Create a simple test pattern: alternating white and black columns
	hires := make([]byte, 8192)
	// Set every other pixel in the first row
	for byteIdx := 0; byteIdx < 40; byteIdx++ {
		hires[offsets[0]+byteIdx] = 0xAA // bits 1,3,5 set = pixels at columns 1,3,5 per byte
	}

	img, err := ConvertHiResToMonochromeImage(hires)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Verify pixel at column 1 of first byte (should be white, bit 1 = pixel[1] = 2, 0xAA & 2 = 2)
	r, _, _, _ := img.At(1, 0).RGBA()
	if r == 0 {
		t.Error("expected white pixel at (1,0)")
	}

	// Verify pixel at column 0 (should be black, bit 0 = pixel[0] = 1, 0xAA & 1 = 0)
	r, _, _, _ = img.At(0, 0).RGBA()
	if r != 0 {
		t.Error("expected black pixel at (0,0)")
	}
}

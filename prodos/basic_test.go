// Copyright Terence J. Boldt (c)2022-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides tests for conversion between BASIC and text

package prodos

import (
	"bytes"
	"testing"
)

func TestConvertBasicToText(t *testing.T) {
	var tests = []struct {
		name  string
		basic []byte
		want  string
	}{
		{
			"Simple",
			[]byte{
				0x14, 0x08, 0x0A, 0x00, 0xBA, 0x22, 0x48, 0x45, 0x4C, 0x4C, 0x4F, 0x20, 0x57, 0x4F, 0x52, 0x4C, 0x44, 0x22, 0x00,
				0x1A, 0x08, 0x14, 0x00, 0x80, 0x00,
				0x00, 0x00},
			"10  PRINT \"HELLO WORLD\"\n20  END \n"},
	}

	for _, tt := range tests {
		testname := tt.name
		t.Run(testname, func(t *testing.T) {
			text := ConvertBasicToText(tt.basic)
			if text != tt.want {
				t.Errorf("%s\ngot '%#v'\nwant '%#v'\n", testname, []byte(text), []byte(tt.want))
			}
		})
	}
}

func TestConvertTextToBasic(t *testing.T) {
	var tests = []struct {
		name      string
		basicText string
		want      []byte
	}{
		{
			"Hello world",
			"10  PRINT \"HELLO WORLD\"\n20  END \n",
			[]byte{
				0x14, 0x08, 0x0A, 0x00, 0xBA, 0x22, 0x48, 0x45, 0x4C, 0x4C, 0x4F, 0x20, 0x57, 0x4F, 0x52, 0x4C, 0x44, 0x22, 0x00,
				0x1A, 0x08, 0x14, 0x00, 0x80, 0x00,
				0x00, 0x00}},
		{
			"Variables",
			"10 A = 1\n",
			[]byte{
				0x09, 0x08, 0x0A, 0x00, 0x41, 0xD0, 0x31, 0x00,
				0x00, 0x00}},
		{
			"Rem",
			"10  REM x y z\n",
			[]byte{
				0x0C, 0x08, 0x0A, 0x00, 0xB2, 0x78, 0x20, 0x79, 0x20, 0x7A, 0x00,
				0x00, 0x00}},
		{
			"Punctuation",
			"1  PRINT ;: PRINT\n",
			[]byte{
				0x0A, 0x08, 0x01, 0x00, 0xBA, 0x3B, 0x3A, 0xBA, 0x00,
				0x00, 0x00}},
	}

	for _, tt := range tests {
		testname := tt.name
		t.Run(testname, func(t *testing.T) {
			basic, _ := ConvertTextToBasic(tt.basicText)
			if bytes.Compare(basic, tt.want) != 0 {
				t.Errorf("%s\ngot '%#v'\nwant '%#v'\n", testname, basic, tt.want)
			}
		})
	}
}

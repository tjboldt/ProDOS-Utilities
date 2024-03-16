// Copyright Terence J. Boldt (c)2021-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides conversion between BASIC and text

package prodos

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var tokens = map[byte]string{
	0x80: "END",
	0x81: "FOR",
	0x82: "NEXT",
	0x83: "DATA",
	0x84: "INPUT",
	0x85: "DEL",
	0x86: "DIM",
	0x87: "READ",
	0x88: "GR",
	0x89: "TEXT",
	0x8A: "PR#",
	0x8B: "IN#",
	0x8C: "CALL",
	0x8D: "PLOT",
	0x8E: "HLIN",
	0x8F: "VLIN",
	0x90: "HGR2",
	0x91: "HGR",
	0x92: "HCOLOR=",
	0x93: "HPLOT",
	0x94: "DRAW",
	0x95: "XDRAW",
	0x96: "HTAB",
	0x97: "HOME",
	0x98: "ROT=",
	0x99: "SCALE=",
	0x9A: "SHLOAD",
	0x9B: "TRACE",
	0x9C: "NOTRACE",
	0x9D: "NORMAL",
	0x9E: "INVERSE",
	0x9F: "FLASH",
	0xA0: "COLOR=",
	0xA1: "POP",
	0xA2: "VTAB",
	0xA3: "HIMEM:",
	0xA4: "LOMEM:",
	0xA5: "ONERR",
	0xA6: "RESUME",
	0xA7: "RECALL",
	0xA8: "STORE",
	0xA9: "SPEED=",
	0xAA: "LET",
	0xAB: "GOTO",
	0xAC: "RUN",
	0xAD: "IF",
	0xAE: "RESTORE",
	0xAF: "&",
	0xB0: "GOSUB",
	0xB1: "RETURN",
	0xB2: "REM",
	0xB3: "STOP",
	0xB4: "ON",
	0xB5: "WAIT",
	0xB6: "LOAD",
	0xB7: "SAVE",
	0xB8: "DEF FN",
	0xB9: "POKE",
	0xBA: "PRINT",
	0xBB: "CONT",
	0xBC: "LIST",
	0xBD: "CLEAR",
	0xBE: "GET",
	0xBF: "NEW",
	0xC0: "TAB",
	0xC1: "TO",
	0xC2: "FN",
	0xC3: "SPC(",
	0xC4: "THEN",
	0xC5: "AT",
	0xC6: "NOT",
	0xC7: "STEP",
	0xC8: "+",
	0xC9: "-",
	0xCA: "*",
	0xCB: "/",
	//0xCC: ";", // fails if this is there
	0xCD: "AND",
	0xCE: "OR",
	0xCF: ">",
	0xD0: "=",
	0xD1: "<",
	0xD2: "SGN",
	0xD3: "INT",
	0xD4: "ABS",
	0xD5: "USR",
	0xD6: "FRE",
	0xD7: "SCRN(",
	0xD8: "PDL",
	0xD9: "POS",
	0xDA: "SQR",
	0xDB: "RND",
	0xDC: "LOG",
	0xDD: "EXP",
	0xDE: "COS",
	0xDF: "SIN",
	0xE0: "TAN",
	0xE1: "ATN",
	0xE2: "PEEK",
	0xE3: "LEN",
	0xE4: "STR$",
	0xE5: "VAL",
	0xE6: "ASC",
	0xE7: "CHR$",
	0xE8: "LEFT$",
	0xE9: "RIGHT$",
	0xEA: "MID$",
}

// ConvertBasicToText converts AppleSoft BASIC to text
func ConvertBasicToText(basic []byte) string {
	var builder strings.Builder

	i := 0

	for {
		lo := basic[i]
		i++
		hi := basic[i]
		i++

		if lo == 0 && hi == 0 {
			return builder.String()
		}

		line := int(basic[i]) + int(basic[i+1])*256
		i += 2

		fmt.Fprintf(&builder, "%d ", line)

		for {
			t := basic[i]
			if t == 0 {
				builder.WriteString("\n")
				i++
				break
			}
			if t > 127 {
				builder.WriteString(" ")
				builder.WriteString(tokens[t])
				builder.WriteString(" ")
			} else {
				builder.WriteString(string(t))
			}
			i++
		}
	}
}

// ConvertTextToBasic converts text to AppleSoft BASIC
func ConvertTextToBasic(text string) ([]byte, error) {
	// convert line endings
	text = strings.Replace(text, "\r\n", "\n", -1)
	text = strings.Replace(text, "\r", "\n", -1)

	starting := true
	parsingLineNumber := false
	parsingData := false
	parsingString := false
	parsingRem := false
	foundToken := false

	currentByte := 0x0801
	var lineNumberString string

	basicFile := new(bytes.Buffer)
	basicLine := new(bytes.Buffer)

	skipChars := 0

	// parse character by character
	for index, c := range text {

		// skip initial whitespace and look for the start of a line number
		if starting {
			if c == '\n' { // skip blank lines
				continue
			}
			if c == ' ' {
				continue
			}
			if c >= '0' && c <= '9' {
				starting = false
				parsingLineNumber = true
			} else {
				return nil, errors.New("unexpected character before line number")
			}
		}

		if skipChars > 0 && c != '\n' {
			skipChars--
			continue
		}

		// parse line number
		if parsingLineNumber {
			if c >= '0' && c <= '9' {
				lineNumberString += string(c)
			} else {
				lineNumber, err := strconv.ParseUint(lineNumberString, 10, 16)
				if err != nil {
					return nil, err
				}
				basicLine.WriteByte(byte(lineNumber % 256)) // low byte
				basicLine.WriteByte(byte(lineNumber / 256)) // high byte
				parsingLineNumber = false
				lineNumberString = ""
			}
		}

		if !parsingLineNumber {
			if c == '\n' {
				starting = true
				parsingLineNumber = false
				parsingData = false
				parsingRem = false
				parsingString = false
				foundToken = false
				skipChars = 0
				currentByte += basicLine.Len()
				currentByte += 3
				// write address of next line
				basicFile.WriteByte(byte(currentByte % 256))
				basicFile.WriteByte(byte(currentByte / 256))
				// write the line
				basicFile.Write(basicLine.Bytes())
				basicFile.WriteByte(0x00)
				basicLine.Reset()
			} else if parsingData {
				basicLine.WriteByte(byte(c))
			} else if parsingRem {
				basicLine.WriteByte(byte(c))
			} else if parsingString {
				basicLine.WriteByte(byte(c))
				if c == '"' {
					parsingString = false
				}
			} else if c == '"' {
				parsingString = true
				basicLine.WriteByte(byte(c))
			} else {
				if c == ' ' {
					continue
				}

				for key, token := range tokens {
					if index < len(text)-len(token) {
						if text[index:index+len(token)] == token {
							basicLine.WriteByte(byte(key))
							skipChars = len(token)
							foundToken = true
							if key == 0x83 {
								parsingData = true
							}
							if key == 0xB2 {
								parsingRem = true
							}
						}
					}
				}

				if foundToken {
					foundToken = false
				} else {
					basicLine.WriteByte(byte(c))
				}
			}
		}
	}
	basicFile.WriteByte(0x00)
	basicFile.WriteByte(0x00)

	return basicFile.Bytes(), nil
}

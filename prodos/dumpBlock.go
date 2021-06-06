package prodos

import "fmt"

func DumpBlock(buffer []byte) {
	for i := 0; i < len(buffer); i += 16 {
		for j := i; j < i+16; j++ {
			fmt.Printf("%02X ", buffer[j])
		}
		for j := i; j < i+16; j++ {
			c := buffer[j] & 127
			if c >= 32 {
				fmt.Printf("%c", c)
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Printf("\n")
	}
}

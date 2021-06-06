package main

import (
	"fmt"
	"os"

	"github.com/tjboldt/ProDOS-Utilities/prodos"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Printf("Usage:\n")
		fmt.Printf("  ProDOS-Utilities DRIVE_IMAGE\n")
		fmt.Printf("  ProDOS-Utilities DRIVE_IMAGE /FULL_PATH\n")
		os.Exit(1)
	}

	fileName := os.Args[1]
	pathName := ""

	if len(os.Args) == 3 {
		pathName = os.Args[2]
	}

	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		os.Exit(1)
	}

	// empty path or volume name means read root directory
	volumeHeader, fileEntries := prodos.ReadDirectory(file, pathName)

	fmt.Printf("VOLUME: %s\n\n", volumeHeader.VolumeName)
	fmt.Printf("NAME           TYPE  BLOCKS  MODIFIED          CREATED            ENDFILE  SUBTYPE\n\n")

	for i := 0; i < len(fileEntries); i++ {
		fmt.Printf("%-15s %s %7d  %s %s %8d %8d\n",
			fileEntries[i].FileName,
			prodos.FileTypeToString(fileEntries[i].FileType),
			fileEntries[i].BlocksUsed,
			prodos.TimeToString(fileEntries[i].ModifiedTime),
			prodos.TimeToString(fileEntries[i].CreationTime),
			fileEntries[i].EndOfFile,
			fileEntries[i].AuxType,
		)
	}
	fmt.Printf("\n")
}

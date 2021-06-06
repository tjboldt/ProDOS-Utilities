package main

import (
	"fmt"
	"os"

	"github.com/tjboldt/ProDOS-Utilities/prodos"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Printf("Usage:")
		fmt.Printf("  ProDOS-Utilities DRIVE_IMAGE")
		fmt.Printf("  ProDOS-Utilities DRIVE_IMAGE /FULL_PATH")
	}

	fileName := os.Args[1]
	pathName := ""

	if len(os.Args) == 3 {
		pathName = os.Args[2]
	}

	// empty path or volume name means read root directory
	volumeHeader, fileEntries := prodos.ReadDirectory(fileName, pathName)

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

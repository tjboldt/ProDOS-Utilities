# ProDOS-Utilities
This project is just starting but is intended to be both a command line tool and library to provide access to ProDOS based hard drive images. It is written in Go to be cross platform (Linux, Windows, macOS etc.). Functionality, naming and parameters are subject to change without notice. This project was started so I would be able to automate writing the firmware file into the drive image for one of my other projects [Apple2-IO-RPi](https://github.com/tjboldt/Apple2-IO-RPi).

## Current command line functionality
1. Export files up to 128KB (only seedling and sapling files are supported)
2. List any directory
3. Display volume bitmap

## Current library functionality
1. Read block
2. Write block
3. Read file (up to 128KB)
4. Read volume bitmap
5. Write volume bitmap
6. Get list of file entries from any path
7. Get volume header

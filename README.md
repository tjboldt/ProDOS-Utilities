# ProDOS-Utilities
This project is just starting but is intended to be both a command line tool and library to provide access to ProDOS based hard drive images. It is written in Go to be cross platform (Linux, Windows, macOS etc.). Functionality, naming and parameters are subject to change without notice. This project was started so I would be able to automate writing the firmware file into the drive image for one of my other projects [Apple2-IO-RPi](https://github.com/tjboldt/Apple2-IO-RPi).

## DISCLAIMER
Being a work in progress, be warned that this code is likely to corrupt drive images so be sure to have backups. Also, command line parameters are likely to change significantly in the future.

## How to get it
There are binaries [here](https://github.com/tjboldt/ProDOS-Utilities/releases/latest)

## Current TODO list
1. Delete directories
2. Add file/directory tests
3. Add rename
4. Add in-place file/directory moves

## Example commands and output

### Specify drive image with -d DRIVEIMAGE (default command lists directory, default path is root of volume)
```
ProDOS-Utilities -d ~/Downloads/Total\ Replay\ v4.01.hdv

/TOTAL.REPLAY

 NAME           TYPE  BLOCKS  MODIFIED          CREATED            ENDFILE  SUBTYPE

 LAUNCHER.SYSTEM SYS      24  2021-FEB-18 01:25 2021-FEB-18 01:25    11416     8192
 TITLE           BIN      17  2019-JUN-29 13:04 2019-JUN-29 13:04     8184     8192
 COVER           BIN      17  2018-MAY-20 15:03 2018-MAY-20 15:03     8184     8192
 HELP            BIN      17  2019-MAY-20 10:24 2019-MAY-20 10:24     8184     8192
 GAMES.CONF      TXT      16  2021-FEB-18 01:12 2021-FEB-18 01:12     7344    32768
 ATTRACT.CONF    TXT      13  2021-FEB-18 01:22 2021-FEB-18 01:22     5748    32768
 FX.CONF         TXT       6  2021-JAN-03 23:00 2021-JAN-03 23:00     2443    32768
 DFX.CONF        TXT       5  2020-MAY-10 15:29 2020-MAY-10 15:29     1981    32768
 PREFS.CONF      TXT       1  2021-FEB-18 01:25 2021-FEB-18 01:25      512    32768
 CREDITS         TXT       1  2021-FEB-18 01:15 2021-FEB-18 01:15      449    32768
 HELPTEXT        TXT       1  2020-JUN-15 11:56 2020-JUN-15 11:56      206    32768
 DECRUNCH        BIN       1  2020-MAR-14 20:35 2020-MAR-14 20:35      303      512
 JOYSTICK        BIN       6  2020-APR-24 23:07 2020-APR-24 23:07     2370     2048
 FINDER.DATA     $C9       3  2020-JUN-27 17:40 2020-JUN-27 17:40      541        0
 FINDER.ROOT     $C9       1  2020-MAY-24 10:56 2020-MAY-24 10:56        9        0
 TITLE.HGR       DIR      23  2021-FEB-18 01:25 2021-FEB-18 01:25    11776        0
 TITLE.DHGR      DIR       4  2021-FEB-18 01:25 2021-FEB-18 01:25     2048        0
 ACTION.HGR      DIR      58  2021-FEB-18 01:25 2021-FEB-18 01:25    29696        0
 ACTION.DHGR     DIR       9  2021-FEB-18 01:25 2021-FEB-18 01:25     4608        0
 ACTION.GR       DIR       1  2021-FEB-18 01:25 2021-FEB-18 01:25      512        0
 ARTWORK.SHR     DIR      17  2021-FEB-18 01:25 2021-FEB-18 01:25     8704        0
 ATTRACT         DIR      26  2021-FEB-18 01:25 2021-FEB-18 01:25    13312        0
 SS              DIR      16  2021-FEB-18 01:25 2021-FEB-18 01:25     8192        0
 DEMO            DIR      15  2021-FEB-18 01:25 2021-FEB-18 01:25     7680        0
 TITLE.ANIMATED  DIR       1  2021-FEB-18 01:25 2021-FEB-18 01:25      512        0
 ICONS           DIR       1  2021-FEB-18 01:25 2021-FEB-18 01:25      512        0
 FX              DIR      16  2021-FEB-18 01:25 2021-FEB-18 01:25     8192        0
 X               DIR      26  2021-FEB-18 01:25 2021-FEB-18 01:25    13312        0
 PRELAUNCH       DIR      27  2021-FEB-18 01:25 2021-FEB-18 01:25    13824        0
 GAMEHELP        DIR      26  2021-FEB-18 01:25 2021-FEB-18 01:25    13312        0

BLOCKS FREE:    15    BLOCKS USED: 65520      TOTAL BLOCKS: 65535
```

### Specifying -p PATHNAME and -c COMMAND

```
ProDOS-Utilities -d ~/Downloads/Total\ Replay\ v4.01.hdv -c ls -p /total.replay/icons

/TOTAL.REPLAY/ICONS

 NAME           TYPE  BLOCKS  MODIFIED          CREATED            ENDFILE  SUBTYPE

 TR.ICONS        $CA       8  2020-MAY-24 10:56 2020-MAY-24 10:56     3524        0

BLOCKS FREE:    15    BLOCKS USED: 65520      TOTAL BLOCKS: 65535
```

### Create a new ProDOS hard drive image
```
ProDOS-Utilities -d new.hdv -c create -b 65535
```

### Add all files from a host directory
```
ProDOS-Utilities -d new.hdv -c putall -i .
```

### Hex dump a block with command readblock and block number (both decimal and hexadecimal input work)
```
ProDOS-Utilities -d new.hdv -c readblock -b 0
Block 0x0000 (0):

0000: 01 38 B0 03 4C 1C 09 78 86 43 C9 03 08 8A 29 70 .80.L..x.CI...)p
0010: 4A 4A 4A 4A 09 C0 85 49 A0 FF 84 48 28 C8 B1 48 JJJJ.@.I ..H(H1H
0020: D0 3A B0 0E A9 03 8D 00 08 E6 3D A5 49 48 A9 5B P:0.)....f=%IH)[
0030: 48 60 85 40 85 48 A0 5E B1 48 99 94 09 C8 C0 EB H`.@.H ^1H...H@k
0040: D0 F6 A2 06 BC 32 09 BD 39 09 99 F2 09 BD 40 09 Pv".<2.=9..r.=@.
0050: 9D 7F 0A CA 10 EE A9 09 85 49 A9 86 A0 00 C9 F9 ...J.n)..I). .Iy
0060: B0 2F 85 48 84 60 84 4A 84 4C 84 4E 84 47 C8 84 0/.H.`.J.L.N.GH.
0070: 42 C8 84 46 A9 0C 85 61 85 4B 20 27 09 B0 66 E6 BH.F)..a.K '.0ff
0080: 61 E6 61 E6 46 A5 46 C9 06 90 EF AD 00 0C 0D 01 afafF%FI..o-....
0090: 0C D0 52 A9 04 D0 02 A5 4A 18 6D 23 0C A8 90 0D .PR).P.%J.m#.(..
00A0: E6 4B A5 4B 4A B0 06 C9 0A F0 71 A0 04 84 4A AD fK%KJ0.I.pq ..J-
00B0: 20 09 29 0F A8 B1 4A D9 20 09 D0 DB 88 10 F6 A0  .).(1JY .P[..v 
00C0: 16 B1 4A 4A 6D 1F 09 8D 1F 09 A0 11 B1 4A 85 46 .1JJm..... .1J.F
00D0: C8 B1 4A 85 47 A9 00 85 4A A0 1E 84 4B 84 61 C8 H1J.G)..J ..K.aH
00E0: 84 4D 20 27 09 B0 35 E6 61 E6 61 A4 4E E6 4E B1 .M '.05fafa$NfN1
00F0: 4A 85 46 B1 4C 85 47 11 4A D0 18 A2 01 A9 00 A8 J.F1L.G.JP.".).(
0100: 91 60 C8 D0 FB E6 61 EA EA CA 10 F4 CE 1F 09 F0 .`HP{fajjJ.tN..p
0110: 07 D0 D8 CE 1F 09 D0 CA 58 4C 00 20 4C 47 09 02 .PXN..PJXL. LG..
0120: 26 50 52 4F 44 4F 53 A5 60 85 44 A5 61 85 45 6C &PRODOS%`.D%a.El
0130: 48 00 08 1E 24 3F 45 47 76 F4 D7 D1 B6 4B B4 AC H...$?EGvtWQ6K4,
0140: A6 2B 18 60 4C BC 09 20 58 FC A0 14 B9 58 09 99 &+.`L<. X| .9X..
0150: B1 05 88 10 F7 4C 55 09 D5 CE C1 C2 CC C5 A0 D4 1...wLU.UNABLE T
0160: CF A0 CC CF C1 C4 A0 D0 D2 CF C4 CF D3 A5 53 29 O LOAD PRODOS%S)
0170: 03 2A 05 2B AA BD 80 C0 A9 2C A2 11 CA D0 FD E9 .*.+*=.@),".JP}i
0180: 01 D0 F7 A6 2B 60 A5 46 29 07 C9 04 29 03 08 0A .Pw&+`%F).I.)...
0190: 28 2A 85 3D A5 47 4A A5 46 6A 4A 4A 85 41 0A 85 (*.=%GJ%FjJJ.A..
01A0: 51 A5 45 85 27 A6 2B BD 89 C0 20 BC 09 E6 27 E6 Q%E.'&+=.@ <.f'f
01B0: 3D E6 3D B0 03 20 BC 09 BC 88 C0 60 A5 40 0A 85 =f=0. <.<.@`%@..
01C0: 53 A9 00 85 54 A5 53 85 50 38 E5 51 F0 14 B0 04 S)..T%S.P8eQp.0.
01D0: E6 53 90 02 C6 53 38 20 6D 09 A5 50 18 20 6F 09 fS..FS8 m.%P. o.
01E0: D0 E3 A0 7F 84 52 08 28 38 C6 52 F0 CE 18 08 88 Pc ..R.(8FRpN...
01F0: F0 F5 00 00 00 00 00 00 00 00 00 00 00 00 00 00 pu..............
```

### Export files (using .bas file extension coverts Applesoft to text file)
```
ProDOS-Utilities -d example.hdv -c get -o Startup.bas -p /EXAMPLE/STARTUP; cat Startup.bas
10  PRINT "HELLO WORLD" 
```

# Gotracker

![Go](https://github.com/gotracker/gotracker/workflows/Go/badge.svg)
 
## What is it?

It's a tracked music player written in Go.

## Why does this exist?

[Heucuva](https://github.com/heucuva/) needed to learn Go forever ago and figured this was a good way to do it.

## What does it play?

Files from/of the following formats/trackers:
* S3M - ScreamTracker 3
* MOD - Protracker/Fasttracker/Startrekker (_internally up-converted to S3M_)
* XM - Fasttracker II
* IT - Impulse Tracker
* Maybe more! (check the support list from the [gotracker/player](https://github.com/gotracker/player) library)

## What systems does it work on?

* Windows (Windows 2000 or newer)
  * Sound Card
    * WinMM (`WAVE_MAPPER` device)
    * DirectSound (via optional build flag: `directsound`)
    * PulseAudio (via optional build flag: `pulseaudio`) - NOTE: Not recommended except for WSL (Linux) builds!
  * File
    * Wave/RIFF file (built-in)
    * Flac (via optional build flag: `flac`)
* Linux
  * Sound Card
    * PulseAudio
  * File
    * Wave/RIFF file (built-in)
    * Flac (via optional build flag: `flac`)

## How do I build this thing?

### What you need

For a Windows build, we recommend the following:
* Windows 2000 (or newer) - we used Windows 11 Pro (Windows 11 Version 23H2 - 22631.3007)
* Visual Studio Code
  * Go extension for VSCode v0.19.0 (or newer) 
  * Go v1.21.5 (or newer)

For a non-Windows (e.g.: Linux) build, we recommend the following:
* Ubuntu 20.04 (or newer) - we used Ubuntu 22.04.2 LTS running in WSL2
* Go v1.21.5 (or newer)

### How to build (on Windows)

1. First, load the project folder in VSCode.  If this is the first time you've ever opened a Go project, VSCode will splash up a thousand alerts asking to install various things for Go. Allow it to install them before continuing on.
2. Next, open a Terminal for `powershell`.
3. Enter the following commands
   ```powershell
   go mod download
   go build
   ```
   When the command completes, you should now have the gotracker.exe file. Drag an .S3M file on top of it!

### How to build (on Linux)

1. Build the player with the following commands
   ```bash
   go mod download
   go build
   ```

NOTE: In order to use PulseAudio, you must have your `PULSE_SERVER` connection string environment variable configured:
* e.g.:
  ```bash
  PULSE_SERVER=tcp:127.0.0.1:4713
  ```
  (*Take note that there are bugs associated with TCP connection strings; see bugs section below*)
  For more information about the `PULSE_SERVER` environment variable, please see the [PulseAudio documentation](https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/User/ServerStrings/).

## How does it work?

Not well, but it's good enough to play some moderately complex stuff.

## Bugs

### Known bugs

| Tags | Notes |
|------|-------|
| `windows` `winmm` | Setting the number of channels to more than 2 may cause WinMM and/or Gotracker to do unusual things. You might be able to get a hardware 4-channel capable card (such as the Aureal Vortex 2 AU8830) to work, but driver inconsistencies and weirdnesses in WinMM will undoubtedly cause needless strife. |
| `pulseaudio` | PulseAudio support is offered through a Pure Go interface originally created by Johann Freymuth, called [jfreymuth/pulse](https://github.com/jfreymuth/pulse). While it seems to work pretty well, it does have some inconsistencies when compared to the FreeDesktop supported C interface. If you see an error about there being a "`missing port in address`" specifically when using a TCP connection string, make sure to append the default port specifier of `:4713` to the end of the `PULSE_SERVER` environment variable. |
| `windows` `directsound` | DirectSound integration is not great code. It works well enough after recent code changes fixing event support, but it's still pretty ugly. |
| `flac` | Flac encoding is still very beta. |

NOTE: for more known bugs, please check the list from the [gotracker/playback](https://github.com/gotracker/playback) library.

### Unknown bugs

* There are many, we're sure.

# Gotracker

![build](https://github.com/gotracker/gotracker/workflows/Go/badge.svg)
 
## What is it?

It's a tracked music player written in Go.

## Why does this exist?

[Heucuva](https://github.com/heucuva/) needed to learn Go forever ago and figured this was a good way to do it.

## What does it play?

At the moment, just S3M (Screamtracker 3) files and very terribly simulated MOD (Protracker/Fasttracker) files.

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
* Windows 2000 (or newer) - we used Windows 10 Pro (Windows 10 Version 20H2)
* Visual Studio Code
  * Go extension for VSCode v0.19.0 (or newer) 
  * Go v1.15.2 (though it will probably compile with Go v1.05 or newer)

For a non-Windows (e.g.: Linux) build, we recommend the following:
* Ubuntu 20.04 (or newer) - we used Ubuntu 20.04.1 LTS running in WSL2
* Go v1.15.2 (or newer)

### How to build (on Windows)

1. First, load the project folder in VSCode.  If this is the first time you've ever opened a Go project, VSCode will splash up a thousand alerts asking to install various things for Go. Allow it to install them before continuing on.
2. Next, open a Terminal for `powershell`.
3. Enter the following command
   ```powershell
   go build
   ```
   When the command completes, you should now have the gotracker.exe file. Drag an .S3M file on top of it!

### How to build (on Linux)

1. Build the player with the following command
   ```bash
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
| `s3m` | Unknown/unhandled commands (effects) will cause a panic. There aren't many left, but there are still some laying around. |
| `player` | The rendering system is fairly bad - it originally was designed only to work with S3M, but we decided to rework some of it to be more flexible. We managed to pull most of the mixing functionality out into somewhat generic structures/algorithms, but it still needs a lot of work. |
| `s3m` | Attempting to load a corrupted S3M file may cause the deserializer to panic or go running off into the weeds indefinitely. |
| `mod` | MOD file support is buggy, at best. |
| `s3m` `opl2` | Attempting to play an S3M file with Adlib/OPL2 instruments does not produce the expected output. The OPL2 code has something wrong with it - it sounds pretty bad, though steps have been taken to remedy its strange output. |
| `windows` `winmm` | Setting the number of channels to more than 2 may cause WinMM and/or Gotracker to do unusual things. You might be able to get a hardware 4-channel capable card (such as the Aureal Vortex 2 AU8830) to work, but driver inconsistencies and weirdnesses in WinMM will undoubtedly cause needless strife. |
| `player` | Channel readouts are associated to the buffer being fed into the output device, so the log line showing the row/channels being played might appear unattached to what's coming from the sound system. |
| `s3m` | Setting the default `C2SPD` value for the `s3m` package to something other than 8363 will cause some unusual behavior - Lower values will reduce the fidelity of the audio, but it will generally sound the same. However, the LFOs (vibrato, tremelo) will become significantly more pronounced the lower the `C2SPD` becomes. The inverse of the observed phenomenon occurs when the `C2SPD` value gets raised. At a certain point much higher than 8363, the LFOs become effectively useless. |
| `player` `mixing` | The mixer still uses some simple saturation mixing techniques, but it's a lot better than it used to be. |
| `pulseaudio` | PulseAudio support is offered through a Pure Go interface originally created by Johann Freymuth, called [jfreymuth/pulse](https://github.com/jfreymuth/pulse). While it seems to work pretty well, it does have some inconsistencies when compared to the FreeDesktop supported C interface. If you see an error about there being a "`missing port in address`" specifically when using a TCP connection string, make sure to append the default port specifier of `:4713` to the end of the `PULSE_SERVER` environment variable. |
| `windows` `directsound` | DirectSound integration is not great code. It works well enough after recent code changes fixing event support, but it's still pretty ugly. |
| `flac` | Flac encoding is still very beta. |


### Unknown bugs

* There are many, we're sure.

## Further reading

Take a look at the fmoddoc2 documentation that the folks at FireLight studios released forever ago - it has great info how how to make a mod player, upgrade it to an s3m player, and then dork around with the internals a bit.

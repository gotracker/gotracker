# Gotracker

## What is it?

It's a tracked music player written in Go.

## Why does this exist?

I needed to learn Go forever ago and figured this was a good way to do it.

## What does it play?

At the moment, just S3M (Screamtracker 3) files and very terribly simulated MOD (Protracker/Fasttracker) files.

## What systems does it work on?

* Windows (Windows 2000 or newer)
  * WinMM (`WAVE_MAPPER` device)
  * File (Wave/RIFF file)
* Linux
  * File (Wave/RIFF file)
  * PulseAudio (via optional build flag)

## How do I build this thing?

### What you need

For a Windows build, I recommend the following:
* Windows 2000 (or newer) - I used Windows 10 Pro (Windows 10 Version 20H2)
* MinGW-w64 with GCC/G++ - I used v8.0.0, but newer is probably ok [download here](https://sourceforge.net/projects/mingw-w64/)
  * You may need to add the `bin` folder in the MinGW-w64 install directory to your `PATH` environment variable.
* Visual Studio Code
  * Go extension for VSCode v0.19.0 (or newer) 
  * Go v1.15.2 (though it will probably compile with Go v1.05 or newer)

For a non-Windows (e.g.: Linux) build, I recommend the following:
* Ubuntu 20.04 (or newer) - I used Ubuntu 20.04.1 LTS running in WSL2
* GCC/G++ 8.0.0 or newer - I used GCC 9.3.0
* Go v1.15.2 (or newer)
* If you want PulseAudio support, there are a few other things to include (install via apt/yum/dnf):
  * libpulse-dev
  * pulseaudio

### How to build (on Windows)

1. First, load the project folder in VSCode.  If this is the first time you've ever opened a Go project, VSCode will splash up a thousand alerts asking to install various things for Go. Allow it to install them before continuing on.
2. Next, open a Terminal for `powershell`.
3. Enter the following command
   ```powershell
   go build
   ```
   When the command completes, you should now have the gotracker.exe file. Drag an .S3M file on top of it!

### How to build (on Linux, without PulseAudio support)

1. Build the player with the following command
   ```bash
   go build
   ```

### How to build (on Linux, with PulseAudio support)

1. Build the player with the following command
   ```bash
   go build -tags=pulseaudio
   ```

## How does it work?

Not well, but it's good enough to play some moderately complex stuff.

## Bugs

### Known bugs

| Tags | Notes |
|--------|---------|
| `s3m` | Unknown/unhandled commands (effects) will cause a panic. There aren't many left, but there are still some laying around. |
| `windows` `winmm` | WinMM support might cause pops and clicks when another prepared buffer chains in. |
| `player` | The rendering system is atrocious - it originally was designed only to work with S3M, but I decided to rework some of it to be more flexible. I didn't get very far, but it was enough to be miserable to look at. |
| `s3m` | Attempting to load a corrupted S3M file may cause the deserializer to panic or go running off into the weeds indefinitely. |
| `mod` | MOD file support is generally terrible. |
| `s3m` | Attempting to play an S3M file with Adlib/OPL2 instruments has unexpected behavior. |
| `windows` `winmm` | Setting the number of channels to more than 2 may cause WinMM and/or Gotracker to do unusual things. You might be able to get a hardware 4-channel capable card (such as the Aureal Vortex 2 AU8830) to work, but driver inconsistencies and weirdnesses in WinMM will undoubtedly cause needless strife. |
| `player` | Channel readouts are associated to the buffer being fed into the output device, so the log line showing the row/channels being played might appear unattached to what's coming from the sound system. |
| `s3m` | Setting the default `C2SPD` value for the `s3m` package to something other than 8363 will cause some unusual behavior - Lower values will reduce the fidelity of the audio, but it will generally sound the same. However, the LFOs (vibrato, tremelo) will become significantly more pronounced the lower the `C2SPD` becomes. The inverse of the observed phenomenon occurs when the `C2SPD` value gets raised. At a certain point much higher than 8363, the LFOs become effectively useless. |
| `player` `mixing` | The mixer uses simple saturation mixing techniques, which relies on knowing exactly how many channels there are in order to properly calculate the mean value for the output. Without this correct number of channels, we can observe the output clipping outside its bounds, as witnessed by pops, clicks, and distortion on the output. This is caused by the player leveraging track channels, which could play multichannel samples, panned across multichannel output devices. It all ends up being somewhat confusing if you look at it from a bird's eye view. The long and short of it is that I need to spend the time building out the linear algebra for properly handling these concerns. |


### Unknown bugs

* There are many, I'm sure.

## Further reading

Take a look at the fmoddoc2 documentation that the folks at FireLight studios released forever ago - it has great info how how to make a mod player, upgrade it to an s3m player, and then dork around with the internals a bit.

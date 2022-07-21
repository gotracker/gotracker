//go:build linux || pulseaudio
// +build linux pulseaudio

package device

import (
	"context"

	"github.com/gotracker/gomixing/mixing"
	"github.com/gotracker/playback/output"
	"github.com/pkg/errors"

	deviceCommon "github.com/gotracker/gotracker/internal/output/device/common"
	"github.com/gotracker/gotracker/internal/output/device/pulseaudio"
)

const pulseaudioName = "pulseaudio"

type pulseaudioDevice struct {
	device
	mix mixing.Mixer
	pa  *pulseaudio.Client
}

func (pulseaudioDevice) GetKind() deviceCommon.Kind {
	return deviceCommon.KindSoundCard
}

// Name returns the device name
func (pulseaudioDevice) Name() string {
	return pulseaudioName
}

func newPulseAudioDevice(settings deviceCommon.Settings) (Device, error) {
	d := pulseaudioDevice{
		device: device{
			onRowOutput: settings.OnRowOutput,
		},
		mix: mixing.Mixer{
			Channels:      settings.Channels,
			BitsPerSample: settings.BitsPerSample,
		},
	}

	play, err := pulseaudio.New("Music", settings.SamplesPerSecond, settings.Channels, settings.BitsPerSample)
	if err != nil {
		return nil, err
	}

	d.pa = play
	return &d, nil
}

// Play starts the wave output device playing
func (d *pulseaudioDevice) Play(in <-chan *output.PremixData) error {
	return d.PlayWithCtx(context.Background(), in)
}

// PlayWithCtx starts the wave output device playing
func (d *pulseaudioDevice) PlayWithCtx(ctx context.Context, in <-chan *output.PremixData) error {
	panmixer := mixing.GetPanMixer(d.mix.Channels)
	if panmixer == nil {
		return errors.New("invalid pan mixer - check channel count")
	}

	myCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-myCtx.Done():
			return myCtx.Err()
		case row, ok := <-in:
			if !ok {
				return nil
			}
			// TODO: in 8-bit mode, pulseaudio wants uint8 format, instead of int8
			mixedData := d.mix.Flatten(panmixer, row.SamplesLen, row.Data, row.MixerVolume)
			d.pa.Output(mixedData)
			if d.onRowOutput != nil {
				d.onRowOutput(deviceCommon.KindSoundCard, row)
			}
		}
	}
}

// Close closes the wave output device
func (d *pulseaudioDevice) Close() error {
	if d.pa != nil {
		return d.pa.Close()
	}
	return nil
}

func init() {
	Map[pulseaudioName] = deviceDetails{
		create: newPulseAudioDevice,
		Kind:   deviceCommon.KindSoundCard,
	}
}

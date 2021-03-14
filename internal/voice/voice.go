package voice

import (
	"github.com/gotracker/voice"

	"gotracker/internal/instrument"
	"gotracker/internal/player/intf"
	"gotracker/internal/song"
)

// New returns a new Voice from the instrument and output channel provided
func New(inst song.Instrument, output *intf.OutputChannel) voice.Voice {
	switch data := inst.GetData().(type) {
	case *instrument.PCM:
		var (
			voiceFilter  song.Filter
			pluginFilter song.Filter
		)
		if factory := inst.GetFilterFactory(); factory != nil {
			voiceFilter = factory(output.Playback.GetSampleRate())
		}
		if factory := inst.GetPluginFilterFactory(); factory != nil {
			pluginFilter = factory(output.Playback.GetSampleRate())
		}
		return NewPCM(PCMConfiguration{
			C2SPD:         inst.GetC2Spd(),
			InitialVolume: inst.GetDefaultVolume(),
			AutoVibrato:   inst.GetAutoVibrato(),
			DataIntf:      data,
			OutputFilter:  output,
			VoiceFilter:   voiceFilter,
			PluginFilter:  pluginFilter,
		})
	case *instrument.OPL2:
		return NewOPL2(OPLConfiguration{
			Chip:          output.Playback.GetOPL2Chip(),
			Channel:       output.ChannelNum,
			C2SPD:         inst.GetC2Spd(),
			InitialVolume: inst.GetDefaultVolume(),
			AutoVibrato:   inst.GetAutoVibrato(),
			DataIntf:      data,
		})
	}
	return nil
}

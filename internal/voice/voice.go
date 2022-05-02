package voice

import (
	"github.com/gotracker/voice"

	"gotracker/internal/filter"
	"gotracker/internal/player/intf"
	"gotracker/internal/song/instrument"
)

// New returns a new Voice from the instrument and output channel provided
func New[TChannelData any](inst *instrument.Instrument, output *intf.OutputChannel[TChannelData]) voice.Voice {
	switch data := inst.GetData().(type) {
	case *instrument.PCM:
		var (
			voiceFilter  filter.Filter
			pluginFilter filter.Filter
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

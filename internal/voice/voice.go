package voice

import (
	"github.com/gotracker/voice"

	"github.com/gotracker/gotracker/internal/filter"
	"github.com/gotracker/gotracker/internal/player/output"
	"github.com/gotracker/gotracker/internal/song/instrument"
)

// New returns a new Voice from the instrument and output channel provided
func New(inst *instrument.Instrument, output *output.Channel) voice.Voice {
	switch data := inst.GetData().(type) {
	case *instrument.PCM:
		var (
			voiceFilter  filter.Filter
			pluginFilter filter.Filter
		)
		if factory := inst.GetFilterFactory(); factory != nil {
			voiceFilter = factory(inst.C2Spd.ToFrequency(), output.GetSampleRate())
		}
		if factory := inst.GetPluginFilterFactory(); factory != nil {
			pluginFilter = factory(inst.C2Spd.ToFrequency(), output.GetSampleRate())
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
			Chip:          output.Config.GetOPL2Chip(),
			Channel:       output.ChannelNum,
			C2SPD:         inst.GetC2Spd(),
			InitialVolume: inst.GetDefaultVolume(),
			AutoVibrato:   inst.GetAutoVibrato(),
			DataIntf:      data,
		})
	}
	return nil
}

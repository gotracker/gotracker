package voice

import (
	"gotracker/internal/instrument"
	"gotracker/internal/player/intf"
	voiceIntf "gotracker/internal/player/intf/voice"
)

// New returns a new Voice from the instrument and output channel provided
func New(inst intf.Instrument, output *intf.OutputChannel) voiceIntf.Voice {
	switch data := inst.GetData().(type) {
	case *instrument.PCM:
		return NewPCM(PCMConfiguration{
			C2SPD:         inst.GetC2Spd(),
			InitialVolume: inst.GetDefaultVolume(),
			AutoVibrato:   inst.GetAutoVibrato(),
			DataIntf:      data,
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

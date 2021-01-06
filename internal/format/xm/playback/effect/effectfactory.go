package effect

import (
	"gotracker/internal/format/xm/layout/channel"
	"gotracker/internal/player/intf"
)

// VolEff is a combined effect that includes a volume effect and a standard effect
type VolEff struct {
	intf.CombinedEffect
	eff intf.Effect
}

func (e VolEff) String() string {
	if e.eff == nil {
		return "..."
	}
	return e.eff.String()
}

// Factory produces an effect for the provided channel pattern data
func Factory(mi intf.Memory, data intf.ChannelData) intf.Effect {
	cd, ok := data.(*channel.Data)
	if !ok {
		return nil
	}

	if !cd.HasEffect() {
		return nil
	}

	eff := VolEff{}
	if cd.What.HasVolume() {
		ve := volumeEffectFactory(mi, cd.Volume)
		if ve != nil {
			eff.Effects = append(eff.Effects, ve)
		}
	}

	if e := standardEffectFactory(mi, cd); e != nil {
		eff.Effects = append(eff.Effects, e)
		eff.eff = e
	}

	switch len(eff.Effects) {
	case 0:
		return nil
	case 1:
		return eff.Effects[0]
	default:
		return &eff
	}
}

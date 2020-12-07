package notes

type Command uint8

const (
	CommandNone            = Command(0 + iota)
	CommandSetSpeed        // Axx
	CommandOrderJump       // Bxx
	CommandRowJump         // Cxy
	CommandVolSlide        // Dxy (%)
	CommandPortaDown       // Exx (%)
	CommandPortaUp         // Fxx (%)
	CommandPortaToNote     // Gxx (*)
	CommandVibrato         // Hxy (*)
	CommandTremor          // Ixy (%)
	CommandArpeggio        // Jxx (%)
	CommandVibratoVolSlide // Kxy (%) = H00 + Dxy
	CommandPortaVolSlide   // Lxy (%) = G00 + Dxy
	CommandReserved01      // Mxx
	CommandReserved02      // Nxx
	CommandSampleOffset    // Oxx (*)
	CommandRetrigVolSlide  // Qxy (%)
	CommandTremolo         // Rxy (%)
	CommandSpecial         // Sxy (%)
	CommandSetTempo        // Txx
	CommandFineVibrato     // Uxy (*)
	CommandGlobalVolume    // Vxx
)

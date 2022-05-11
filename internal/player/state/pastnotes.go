package state

import (
	"sync"

	"github.com/gotracker/gotracker/internal/song/note"
)

type pastNote[TChannelData any] struct {
	activeState *Active[TChannelData]
}

func (pn pastNote[TChannelData]) IsValid() bool {
	if pn.activeState.Voice == nil {
		return false
	}

	return !pn.activeState.Voice.IsDone()
}

type pastNotesForChannel[TChannelData any] struct {
	pn []*pastNote[TChannelData]
}

func (p *pastNotesForChannel[TChannelData]) Remove(pn *pastNote[TChannelData]) []*pastNote[TChannelData] {
	var kept, removed []*pastNote[TChannelData]
	for _, a := range p.pn {
		if a != pn {
			kept = append(kept, a)
		} else {
			removed = append(removed, a)
		}
	}
	p.pn = kept
	return removed
}

type pastNoteOnChannel[TChannelData any] struct {
	ch int
	pn *pastNote[TChannelData]
}

type PastNotesProcessor[TChannelData any] struct {
	order []pastNoteOnChannel[TChannelData]
	ch    map[int]*pastNotesForChannel[TChannelData]
	mu    sync.Mutex
	max   int
}

func (p *PastNotesProcessor[TChannelData]) Add(ch int, data *Active[TChannelData]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch == nil {
		p.ch = make(map[int]*pastNotesForChannel[TChannelData])
	}

	if c := len(p.order) - p.max; c > 0 {
		o := p.order[0:c]
		p.order = p.order[c:]

		for _, pn := range o {
			pn.pn.activeState.Reset()
			pnoc := p.ch[pn.ch]
			if pnoc == nil {
				p.ch[pn.ch] = &pastNotesForChannel[TChannelData]{}
				continue
			}
			for _, v := range p.ch[pn.ch].Remove(pn.pn) {
				v.activeState.Reset()
			}
		}
	}

	pn := &pastNote[TChannelData]{
		activeState: data,
	}

	cl := pastNoteOnChannel[TChannelData]{
		ch: ch,
		pn: pn,
	}

	pnoc := p.ch[ch]
	if pnoc == nil {
		pnoc = &pastNotesForChannel[TChannelData]{}
		p.ch[ch] = pnoc
	}

	pnoc.pn = append(pnoc.pn, pn)
	p.order = append(p.order, cl)
}

func (p *PastNotesProcessor[TChannelData]) Do(ch int, action note.Action) {
	if action == note.ActionContinue {
		return
	}

	pnoc := p.ch[ch]
	if pnoc == nil {
		return
	}

	for _, pn := range pnoc.pn {
		if pn.activeState.Voice == nil {
			continue
		}

		switch action {
		case note.ActionRelease:
			pn.activeState.Voice.Release()
		case note.ActionFadeout:
			pn.activeState.Voice.Release()
			pn.activeState.Voice.Fadeout()
		}
	}

	if action == note.ActionCut {
		pnoc.pn = nil
	}
}

func (p *PastNotesProcessor[TChannelData]) GetNotesForChannel(ch int) []*Active[TChannelData] {
	var pastNotes []*Active[TChannelData]
	if pnoc := p.ch[ch]; pnoc != nil {
		var npns []*pastNote[TChannelData]
		for _, pn := range pnoc.pn {
			if !pn.IsValid() {
				continue
			}

			pastNotes = append(pastNotes, pn.activeState)
		}
		pnoc.pn = npns
	}
	return pastNotes
}

func (p *PastNotesProcessor[TChannelData]) SetMax(max int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.max = max
}

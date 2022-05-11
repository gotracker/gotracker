package state

import (
	"sync"

	"github.com/gotracker/gotracker/internal/song/note"
)

type pastNote struct {
	activeState *Active
}

func (pn pastNote) IsValid() bool {
	if pn.activeState.Voice == nil {
		return false
	}

	return !pn.activeState.Voice.IsDone()
}

type pastNotesForChannel struct {
	pn []*pastNote
}

func (p *pastNotesForChannel) Remove(pn *pastNote) []*pastNote {
	var kept, removed []*pastNote
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

type pastNoteOnChannel struct {
	ch int
	pn *pastNote
}

type PastNotesProcessor struct {
	order []pastNoteOnChannel
	ch    map[int]*pastNotesForChannel
	mu    sync.Mutex
	max   int
}

func (p *PastNotesProcessor) Add(ch int, data *Active) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch == nil {
		p.ch = make(map[int]*pastNotesForChannel)
	}

	if c := len(p.order) - p.max; c > 0 {
		o := p.order[0:c]
		p.order = p.order[c:]

		for _, pn := range o {
			pn.pn.activeState.Reset()
			pnoc := p.ch[pn.ch]
			if pnoc == nil {
				p.ch[pn.ch] = &pastNotesForChannel{}
				continue
			}
			for _, v := range p.ch[pn.ch].Remove(pn.pn) {
				v.activeState.Reset()
			}
		}
	}

	pn := &pastNote{
		activeState: data,
	}

	cl := pastNoteOnChannel{
		ch: ch,
		pn: pn,
	}

	pnoc := p.ch[ch]
	if pnoc == nil {
		pnoc = &pastNotesForChannel{}
		p.ch[ch] = pnoc
	}

	pnoc.pn = append(pnoc.pn, pn)
	p.order = append(p.order, cl)
}

func (p *PastNotesProcessor) Do(ch int, action note.Action) {
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

func (p *PastNotesProcessor) GetNotesForChannel(ch int) []*Active {
	var pastNotes []*Active
	if pnoc := p.ch[ch]; pnoc != nil {
		var npns []*pastNote
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

func (p *PastNotesProcessor) SetMax(max int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.max = max
}

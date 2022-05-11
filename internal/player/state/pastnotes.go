package state

import (
	"github.com/gotracker/gotracker/internal/song/note"
)

type pastNote struct {
	ch          int
	activeState *Active
}

func (pn *pastNote) IsValid() bool {
	if pn.activeState.Voice == nil {
		return false
	}

	return !pn.activeState.Voice.IsDone()
}

type PastNotesProcessor struct {
	order []*pastNote
	max   int
}

func (p *PastNotesProcessor) Add(ch int, data *Active) {
	if c := len(p.order) - p.max; c > 0 {
		o := p.order[0:c]
		p.order = p.order[c:]

		for _, pn := range o {
			pn.activeState.Reset()
		}
	}

	cl := &pastNote{
		ch:          ch,
		activeState: data,
	}

	p.order = append(p.order, cl)
}

func (p *PastNotesProcessor) Do(ch int, action note.Action) {
	if action == note.ActionContinue {
		return
	}

	for _, pn := range p.order {
		if pn.ch != ch {
			continue
		}

		if !pn.IsValid() {
			continue
		}

		switch action {
		case note.ActionCut:
			pn.activeState.Reset()
		case note.ActionRelease:
			pn.activeState.Voice.Release()
		case note.ActionFadeout:
			pn.activeState.Voice.Release()
			pn.activeState.Voice.Fadeout()
		}
	}
}

func (p *PastNotesProcessor) GetNotesForChannel(ch int) []*Active {
	var pastNotes []*Active
	for _, pn := range p.order {
		if pn.ch != ch {
			continue
		}

		if !pn.IsValid() {
			continue
		}

		pastNotes = append(pastNotes, pn.activeState)
	}
	return pastNotes
}

func (p *PastNotesProcessor) SetMax(max int) {
	p.max = max
}

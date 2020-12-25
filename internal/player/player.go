package player

import (
	"context"
	"errors"
	"log"
	"time"

	device "github.com/gotracker/gosound"

	"gotracker/internal/player/render"
	"gotracker/internal/player/state"
)

type playerState int

const (
	playerStateIdle = playerState(iota)
	playerStatePlaying
	playerStatePaused
	playerStateStopped
)

// Player is a player of fine tracked musics
type Player struct {
	output         chan<- *device.PremixData
	ctx            context.Context
	cancel         context.CancelFunc
	state          playerState
	playCh         chan struct{}
	pauseCh        chan struct{}
	resumeCh       chan struct{}
	stopCh         chan struct{}
	lastUpdateTime time.Time
	ss             *state.Song
	sampler        *render.Sampler
}

// NewPlayer returns a new Player instance
func NewPlayer(ctx context.Context, output chan<- *device.PremixData, sampler *render.Sampler) (*Player, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if output == nil {
		return nil, errors.New("a valid output channel must be provided")
	}

	if sampler == nil {
		return nil, errors.New("a valid sampler must be provided")
	}

	myCtx, cancel := context.WithCancel(ctx)

	p := Player{
		output:   output,
		ctx:      myCtx,
		cancel:   cancel,
		state:    playerStateIdle,
		playCh:   make(chan struct{}, 1),
		pauseCh:  make(chan struct{}, 1),
		resumeCh: make(chan struct{}, 1),
		stopCh:   make(chan struct{}, 1),
		sampler:  sampler,
	}

	go func() {
		defer p.cancel()
		if err := p.runStateMachine(); err != nil {
			if err != state.ErrStopSong {
				log.Fatalln(err)
			}
		}
		p.state = playerStateStopped
	}()

	return &p, nil
}

// Play starts a player playing
func (p *Player) Play(ss *state.Song) error {
	if err := p.ctx.Err(); err != nil {
		return err
	}

	p.ss = ss

	p.playCh <- struct{}{}
	return nil
}

// WaitUntilDone waits until the player is done
func (p *Player) WaitUntilDone() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

func (p *Player) runStateMachine() error {
	for {
		var stateFunc func() error
		switch p.state {
		case playerStateIdle:
			stateFunc = p.runStateIdle
		case playerStatePlaying:
			stateFunc = p.runStatePlaying
		case playerStatePaused:
			stateFunc = p.runStatePaused
		default:
			return state.ErrStopSong
		}
		if stateFunc == nil {
			return state.ErrStopSong
		}
		if err := stateFunc(); err != nil {
			return err
		}
		//if errors.Is(err, state.ErrStopSong) {
		//	return err
		//}
		time.Sleep(time.Duration(8) * time.Millisecond)
	}
}

func (p *Player) runStateIdle() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case <-p.playCh:
		p.lastUpdateTime = time.Now()
		p.state = playerStatePlaying
	case <-p.pauseCh:
		// eat it if we're idle.
	case <-p.resumeCh:
		// eat it if we're idle.
	case <-p.stopCh:
		return state.ErrStopSong
	}
	return nil
}

func (p *Player) runStatePaused() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case <-p.playCh:
		return errors.New("already playing")
	case <-p.pauseCh:
		// eat it if we're already paused.
	case <-p.resumeCh:
		p.lastUpdateTime = time.Now()
		p.state = playerStatePlaying
	case <-p.stopCh:
		return state.ErrStopSong
	}
	return nil
}

func (p *Player) runStatePlaying() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case <-p.playCh:
		return errors.New("already playing")
	case <-p.resumeCh:
		// eat it if we're already playing.
	case <-p.stopCh:
		return state.ErrStopSong
	default:
	}

	// run our update
	now := time.Now()
	delta := now.Sub(p.lastUpdateTime)
	if err := p.update(delta); err != nil {
		return err
	}
	p.lastUpdateTime = now
	return nil
}

func (p *Player) update(delta time.Duration) error {
	premix, err := p.ss.RenderOneRow(p.sampler)
	if err != nil {
		return err
	}
	if premix != nil && premix.Data != nil && len(premix.Data) != 0 {
		p.output <- premix
	}
	return nil
}

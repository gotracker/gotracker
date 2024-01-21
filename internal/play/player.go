package play

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gotracker/playback/player/machine"
	"github.com/gotracker/playback/player/sampler"
	"github.com/gotracker/playback/song"
	"github.com/gotracker/playback/tracing"
)

type playerState int

const (
	playerStateIdle = playerState(iota)
	playerStatePlaying
	playerStatePaused
	playerStateStopped
)

type playerOperation int

const (
	playerOperationPlay = playerOperation(iota)
	playerOperationResume
	playerOperationPause
	playerOperationStop
)

type playerOp struct {
	op       playerOperation
	response func(err error)
}

// Player is a player of fine tracked musics
type Player struct {
	ctx            context.Context
	cancel         context.CancelCauseFunc
	state          playerState
	opCh           chan playerOp
	lastUpdateTime time.Time
	m              machine.MachineTicker
	s              *sampler.Sampler
	tracer         tracing.Tracer
	ticker         *time.Ticker
	tickerCh       <-chan time.Time
	myTickerCh     chan time.Time
}

// NewPlayer returns a new Player instance
func NewPlayer(ctx context.Context, tickInterval time.Duration) (*Player, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	myCtx, cancel := context.WithCancelCause(ctx)

	p := Player{
		ctx:    myCtx,
		cancel: cancel,
		state:  playerStateIdle,
		opCh:   make(chan playerOp, 1),
	}

	if tickInterval != time.Duration(0) {
		p.ticker = time.NewTicker(tickInterval)
		p.tickerCh = p.ticker.C
	} else {
		p.myTickerCh = make(chan time.Time, 1)
		p.tickerCh = p.myTickerCh
		p.myTickerCh <- time.Now()
	}

	go func() {
		defer func() {
			close(p.opCh)

			if p.ticker != nil {
				p.ticker.Stop()
			} else {
				close(p.myTickerCh)
			}
		}()
		err := p.runStateMachine()
		if err == nil {
			err = song.ErrStopSong
		}
		p.state = playerStateStopped
		p.cancel(err)
	}()

	return &p, nil
}

// Play starts a player playing
func (p *Player) Play(m machine.MachineTicker, out *sampler.Sampler, tracer tracing.Tracer) error {
	if err := p.ctx.Err(); err != nil {
		return err
	}

	p.m = m
	p.s = out
	p.tracer = tracer
	return p.enqueueAndAwaitResponse(playerOperationPlay)
}

func (p *Player) enqueueAndAwaitResponse(op playerOperation) error {
	var (
		wg     sync.WaitGroup
		result error
	)

	wg.Add(1)
	p.opCh <- playerOp{
		op: op,
		response: func(err error) {
			defer wg.Done()
			result = err
		},
	}
	wg.Wait()
	return result
}

// WaitUntilDone waits until the player is done
func (p *Player) WaitUntilDone() error {
	<-p.ctx.Done()
	if err := p.ctx.Err(); err != nil {
		switch {
		case errors.Is(err, song.ErrStopSong):
			return nil
		case errors.Is(err, context.Canceled):
			err := context.Cause(p.ctx)
			if errors.Is(err, song.ErrStopSong) {
				return nil
			}
			return err
		default:
			return err
		}
	}
	return nil
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
			return song.ErrStopSong
		}
		if err := stateFunc(); err != nil {
			return err
		}
	}
}

func (p *Player) runStateIdle() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case op := <-p.opCh:
		switch op.op {
		case playerOperationPlay:
			p.lastUpdateTime = time.Now()
			p.state = playerStatePlaying
			op.response(nil)
		case playerOperationPause:
			// eat it if we're idle.
			op.response(nil)
		case playerOperationResume:
			op.response(nil)
		case playerOperationStop:
			op.response(nil)
			return song.ErrStopSong
		default:
			op.response(fmt.Errorf("unhandled player operation while idle: %d", op.op))
			return song.ErrStopSong
		}
	}
	return nil
}

func (p *Player) runStatePaused() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case op := <-p.opCh:
		switch op.op {
		case playerOperationPlay:
			op.response(errors.New("already playing"))
		case playerOperationPause:
			// eat it if we're already paused.
			op.response(nil)
		case playerOperationResume:
			op.response(nil)
			p.lastUpdateTime = time.Now()
			p.state = playerStatePlaying
		case playerOperationStop:
			op.response(nil)
			return song.ErrStopSong
		default:
			op.response(fmt.Errorf("unhandled player operation while paused: %d", op.op))
			return song.ErrStopSong
		}
	}
	return nil
}

func (p *Player) runStatePlaying() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case op := <-p.opCh:
		switch op.op {
		case playerOperationPlay:
			op.response(errors.New("already playing"))
		case playerOperationPause:
			op.response(nil)
			p.state = playerStatePaused
			return nil
		case playerOperationResume:
			// eat it if we're already playing.
			op.response(nil)
		case playerOperationStop:
			op.response(nil)
			return song.ErrStopSong
		default:
			op.response(fmt.Errorf("unhandled player operation while playing: %d", op.op))
			return song.ErrStopSong
		}
	case <-p.tickerCh:
		if p.ticker == nil {
			// give ourselves something to hit the next time through
			p.myTickerCh <- time.Now()
		}
	}

	// run our update
	now := time.Now()
	delta := now.Sub(p.lastUpdateTime)
	err := p.update(delta)
	p.lastUpdateTime = now
	return err
}

func (p *Player) update(delta time.Duration) error {
	remaining := delta

	var first time.Duration
	firstSet := false

	for !firstSet || remaining < first {
		if err := func() error {
			defer func() {
				if p.tracer != nil {
					p.tracer.OutputTraces()
				}
			}()

			start := time.Now()
			if err := p.m.Tick(p.s); err != nil {
				return err
			}
			dur := time.Since(start)

			if !firstSet {
				firstSet = true
				first = dur
			}

			remaining -= dur
			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

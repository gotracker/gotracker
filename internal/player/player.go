package player

import (
	"context"
	"errors"
	"log"
	"time"

	device "github.com/gotracker/gosound"

	"gotracker/internal/player/intf"
	"gotracker/internal/song"
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
	playRespCh     chan error
	pauseCh        chan struct{}
	pauseRespCh    chan error
	resumeCh       chan struct{}
	resumeRespCh   chan error
	stopCh         chan struct{}
	stopRespCh     chan error
	lastUpdateTime time.Time
	playback       intf.Playback
	ticker         *time.Ticker
	tickerCh       <-chan time.Time
	myTickerCh     chan time.Time
}

// NewPlayer returns a new Player instance
func NewPlayer(ctx context.Context, output chan<- *device.PremixData, tickInterval time.Duration) (*Player, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if output == nil {
		return nil, errors.New("a valid output channel must be provided")
	}

	myCtx, cancel := context.WithCancel(ctx)

	p := Player{
		output:       output,
		ctx:          myCtx,
		cancel:       cancel,
		state:        playerStateIdle,
		playCh:       make(chan struct{}, 1),
		playRespCh:   make(chan error, 1),
		pauseCh:      make(chan struct{}, 1),
		pauseRespCh:  make(chan error, 1),
		resumeCh:     make(chan struct{}, 1),
		resumeRespCh: make(chan error, 1),
		stopCh:       make(chan struct{}, 1),
		stopRespCh:   make(chan error, 1),
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
		defer p.cancel()
		if err := p.runStateMachine(); err != nil {
			if err != song.ErrStopSong {
				log.Fatalln(err)
			}
		}
		p.state = playerStateStopped
	}()

	return &p, nil
}

// Play starts a player playing
func (p *Player) Play(playback intf.Playback) error {
	if err := p.ctx.Err(); err != nil {
		return err
	}

	p.playback = playback

	p.playCh <- struct{}{}
	return <-p.playRespCh
}

// WaitUntilDone waits until the player is done
func (p *Player) WaitUntilDone() error {
	<-p.ctx.Done()
	return p.ctx.Err()
}

func (p *Player) runStateMachine() error {
	defer func() {
		err := errors.New("end")
		p.playRespCh <- err
		p.pauseRespCh <- err
		p.resumeRespCh <- err
		p.stopRespCh <- err

		close(p.playCh)
		close(p.playRespCh)
		close(p.pauseCh)
		close(p.pauseRespCh)
		close(p.resumeCh)
		close(p.resumeRespCh)
		close(p.stopCh)
		close(p.stopRespCh)

		if p.ticker != nil {
			p.ticker.Stop()
		} else {
			close(p.myTickerCh)
		}
	}()
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
		if stateFunc == nil {
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
	case <-p.playCh:
		p.lastUpdateTime = time.Now()
		p.state = playerStatePlaying
		p.playRespCh <- nil
	case <-p.pauseCh:
		// eat it if we're idle.
		p.pauseRespCh <- nil
	case <-p.resumeCh:
		// eat it if we're idle.
		p.resumeRespCh <- nil
	case <-p.stopCh:
		p.stopRespCh <- nil
		return song.ErrStopSong
	}
	return nil
}

func (p *Player) runStatePaused() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case <-p.playCh:
		p.playRespCh <- errors.New("already playing")
	case <-p.pauseCh:
		// eat it if we're already paused.
		p.pauseRespCh <- nil
	case <-p.resumeCh:
		p.resumeRespCh <- nil
		p.lastUpdateTime = time.Now()
		p.state = playerStatePlaying
	case <-p.stopCh:
		p.stopRespCh <- nil
		return song.ErrStopSong
	}
	return nil
}

func (p *Player) runStatePlaying() error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case <-p.playCh:
		p.playRespCh <- errors.New("already playing")
		return nil
	case <-p.pauseCh:
		p.pauseRespCh <- nil
		p.state = playerStatePaused
		return nil
	case <-p.resumeCh:
		// eat it if we're already playing.
		p.resumeRespCh <- nil
	case <-p.stopCh:
		p.stopRespCh <- nil
		return song.ErrStopSong
	case <-p.tickerCh:
		if p.ticker == nil {
			// give ourselves something to hit the next time through
			p.myTickerCh <- time.Now()
		}
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
	return p.playback.Update(delta, p.output)
}

package playlist

import (
	"math"
	"math/rand"
	"time"
)

type Position struct {
	Order int
	Row   int
}

type Song struct {
	Filepath string
	Start    Position
	End      Position
	Loop     bool
}

type Playlist struct {
	songs             []Song
	currentPlayOrder  []int
	lastPlayed        []int
	lastPlayedMaxSize int
}

func New() *Playlist {
	p := Playlist{}
	return &p
}

func (p *Playlist) Add(s Song) {
	p.songs = append(p.songs, s)
	p.currentPlayOrder = append(p.currentPlayOrder, len(p.songs)-1)
	p.lastPlayedMaxSize = int(math.Floor(float64(len(p.songs)) / math.Sqrt2))
}

func (p *Playlist) MarkPlayed(s *Song) {
	for i := range p.songs {
		if &p.songs[i] == s {
			p.lastPlayed = append(p.lastPlayed, i)
			n := len(p.lastPlayed) - p.lastPlayedMaxSize
			if n > 0 {
				p.lastPlayed = p.lastPlayed[n:]
			}
			return
		}
	}
}

func (p *Playlist) GetPlaylist(randomized bool) []int {
	if randomized {
		rand.Seed(time.Now().Unix())
	randomize:
		rand.Shuffle(len(p.currentPlayOrder), func(i, j int) {
			p.currentPlayOrder[j], p.currentPlayOrder[i] = p.currentPlayOrder[i], p.currentPlayOrder[j]
		})
		if len(p.currentPlayOrder) > p.lastPlayedMaxSize && p.lastPlayedMaxSize >= 1 {
			for _, lastIdx := range p.lastPlayed {
				for _, curIdx := range p.currentPlayOrder[:p.lastPlayedMaxSize] {
					if curIdx == lastIdx {
						goto randomize
					}
				}
			}
		}
	}
	return p.currentPlayOrder
}

func (p *Playlist) GetSong(idx int) *Song {
	if idx < 0 || idx >= len(p.songs) {
		return nil
	}

	return &p.songs[idx]
}

package player

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

type tracingMsgFunc func() string

type tracingState struct {
	chMap  map[int]*tracingChannelState
	traces []tracingMsgFunc
	c      chan func(w io.Writer)
	wg     sync.WaitGroup
}

type tracingChannelState struct {
	traces []tracingMsgFunc
}

func (t *Tracker) TraceChannel(ch int, msgFunc tracingMsgFunc) {
	if t.tracingFile == nil {
		return
	}

	tc := t.tracingState.chMap[ch]
	if tc == nil {
		tc = &tracingChannelState{}
		t.tracingState.chMap[ch] = tc
	}

	tc.traces = append(tc.traces, msgFunc)
}

func (t *Tracker) TraceTick(msgFunc tracingMsgFunc) {
	if t.tracingFile == nil {
		return
	}

	t.tracingState.traces = append(t.tracingState.traces, msgFunc)
}

type tracingColumn struct {
	heading string
	rows    []string
}

type TracingTable struct {
	cols    []*tracingColumn
	name    string
	maxRows int
}

func NewTracingTable(name string, headers ...string) TracingTable {
	tt := TracingTable{
		name: name,
	}
	for _, h := range headers {
		tt.cols = append(tt.cols, &tracingColumn{
			heading: h,
		})
	}
	return tt
}

func (tt *TracingTable) AddRow(cols ...any) {
	for i, col := range cols {
		c := tt.cols[i]
		c.rows = append(c.rows, fmt.Sprint(col))
	}
	tt.maxRows++
}

func (tt TracingTable) Fprintln(w io.Writer, colSep string, withRowNums bool) error {
	head := []string{tt.name}
	for _, c := range tt.cols {
		head = append(head, c.heading)
	}
	if _, err := fmt.Fprintln(w, strings.Join(head, colSep)); err != nil {
		return err
	}

	for r := 0; r < tt.maxRows; r++ {
		numCols := len(tt.cols)
		colStart := 0
		if withRowNums {
			numCols++
			colStart++
		}
		cols := []string{""}
		if withRowNums {
			cols[0] = fmt.Sprintf("[%d]", r+1)
		}
		for _, col := range tt.cols {
			if r >= len(col.rows) {
				return errors.New("not enough rows to satisfy TracingTable writer")
			}
			cols = append(cols, col.rows[r])
		}
		if _, err := fmt.Fprintln(w, strings.Join(cols, colSep)); err != nil {
			return err
		}
	}

	return nil
}

type TraceableIntf interface {
	OutputTraces(out chan<- func(w io.Writer))
}

func (t *Tracker) OutputTraces() {
	if t.tracingFile != nil && t.Traceable != nil {
		if t.tracingState.c == nil {
			t.tracingState.c = make(chan func(w io.Writer), 1000*1000)
			go func() {
				defer close(t.tracingState.c)
				defer t.tracingFile.Close()

				t.tracingState.wg.Add(1)
				defer t.tracingState.wg.Done()

				for {
					select {
					case tr, ok := <-t.tracingState.c:
						if !ok {
							return
						}
						tr(t.tracingFile)
					}
				}
			}()
		}
		t.Traceable.OutputTraces(t.tracingState.c)
	}
}

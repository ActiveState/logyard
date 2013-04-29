package state

import (
	"fmt"
	"github.com/ActiveState/log"
	"logyard/util/retry"
	"sync"
)

type StateMachine struct {
	ActionCh chan int
	stopCh   chan bool
	process  Process
	retryer  retry.Retryer
	state    State
	rev      int64
	mux      sync.Mutex
}

func NewStateMachine(process Process, retryer retry.Retryer) *StateMachine {
	m := &StateMachine{}
	m.process = process
	m.retryer = retryer
	m.state = Stopped{m}
	m.rev = 1
	m.ActionCh = make(chan int)
	m.stopCh = make(chan bool)
	go m.Run()
	return m
}

func (m *StateMachine) Run() {
	for {
		select {
		case action := <-m.ActionCh:
			log.Infof("Incoming state change request: %v", action)
			func() {
				m.mux.Lock()
				defer m.mux.Unlock()
				oldState := m.state
				log.Infof("About to state change [%s]: %s -[%v]-> ?? (%d)\n",
					m.process.String(), oldState, action, m.rev)
				m.state = m.state.Transition(action, m.rev)
				m.rev += 1
				log.Infof("State change [%s]: %s => %s (%d)\n",
					m.process.String(), oldState, m.state, m.rev)
			}()
		case <-m.stopCh:
			// XXX: not sure if this will be run even if m.ActionCh
			// has backlog.
			log.Infof("Exiting state machine loop")
			return
		}
	}
}

func (m *StateMachine) GetState() (State, int64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.IsStopped() {
		panic("stopped")
	}
	return m.state, m.rev
}

func (m *StateMachine) Stop() {
	log.Info("Stopping STM...")
	m.stopCh <- true
	m.mux.Lock()
	defer m.mux.Unlock()

	// reset fields to prevent (buggy) future use
	close(m.ActionCh)
	m.process = nil
	m.state = nil
	m.rev = -10

	// sentinal to indicate the stopped state.
	m.ActionCh = nil
	log.Info("Stopped STM.")
}

func (m *StateMachine) IsStopped() bool {
	// XXX: ideally we should use locking here, but don't want to
	// introduce a deadlock when called from `SetStateCustom` which
	// also uses locking.
	return m.ActionCh == nil
}

func (m *StateMachine) SetStateCustom(rev int64, fn func() State) int64 {
	m.mux.Lock()
	defer m.mux.Unlock()
	if !m.IsStopped() && rev == m.rev {
		oldState := m.state
		m.state = fn()
		if m.state == nil {
			panic("nil state")
		}
		m.rev += 1
		fmt.Printf("Custom state change [%s]: %s => %s (%d)\n",
			m.process.String(), oldState, m.state, m.rev)
		return m.rev
	}
	fmt.Printf("Can't set state; rev changed (expected %d, has %d) or stopped (%v)\n",
		rev, m.rev, m.IsStopped())
	return -1
}

func (m *StateMachine) SetState(rev int64, state State) int64 {
	return m.SetStateCustom(rev, func() State {
		return state
	})
}

func (s *StateMachine) stop(rev int64) State {
	err := s.process.Stop()
	if err != nil {
		return Fatal{s}
	}
	return Stopped{s}
}

func (s *StateMachine) start(rev int64) State {
	// start it
	log.Infof("[drain:%s] STM starting process", s.process.String())
	err := s.process.Start()
	if err != nil {
		return Fatal{s}
	} else {
		rev = rev + 1 // account for settig of RunningState
		go s.monitor(rev)
		return Running{s}
	}
	panic("unreachable")
}

func (s *StateMachine) monitor(rev int64) {
	err := s.process.Wait()
	log.Infof("[drain:%s] Process exited with %v",
		s.process.String(), err)
	if err == nil {
		// If a process exited cleanly (no errors), then just mark it
		// as STOPPED without retrying.
		s.SetState(rev, Stopped{s}) // rev confict here is normal.
	} else {
		s.SetStateCustom(rev, func() State {
			rev = rev + 1 // account for setting of RetryingState
			go s.doretry(rev, err)
			return Retrying{s}
		})
	}
}

func (s *StateMachine) doretry(rev int64, err error) {
	// This could block.
	if s.retryer.Wait(
		fmt.Sprintf("[drain:%s] Drain exited abruptly -- %v",
			s.process.String(), err)) {
		log.Infof("[drain:%s] Retrying now.", s.process.String())
		// TODO: move 'drain' specific message (above) out of the
		// state package.
		s.SetStateCustom(rev, func() State {
			return s.start(rev)
		})
	} else {
		log.Infof("[drain:%s] retried too long; marking as FATAL",
			s.process.String())
		s.SetState(rev, Fatal{s})
	}
}

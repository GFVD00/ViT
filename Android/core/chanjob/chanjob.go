package chanjob

import "sync/atomic"

type Flags struct {
	Send   bool
	Receive bool
    Voice bool
}

type Status struct{
    st atomic.Bool
    notify chan struct{} 
}

func CreateStatus()*Status{
     NewStatus:= &Status{
        notify: make(chan struct{},1),
     }
     NewStatus.Set(false)
     return NewStatus
}

func (s *Status)Set(val bool){
    s.st.Store(val)

    select{
      case s.notify <-struct{}{}:
      default:
    }
}

func (s *Status)Get()bool{
     return s.st.Load()
}

func (s *Status)WaitStatus(){
     <-s.notify
}

func ChangeChan(s,r,v bool,ch chan<-*Flags){
    flags := &Flags{
		Send:   s,
		Receive: r,
        Voice: v,
	}
    ch <-flags
}

func CreateChan() chan *Flags {
    flags := &Flags{
		Send:   false,
		Receive: false,
        Voice: true,
	}
    ch:=make(chan *Flags,1)
	ch <-flags

    return ch
}

func GetChan(ch <-chan Flags)Flags{
    flags:= <-ch
    return flags
}


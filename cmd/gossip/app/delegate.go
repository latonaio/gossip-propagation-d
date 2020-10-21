package app

type Delegate struct {
	cacheCh chan []byte
}

func NewDelegate() *Delegate {
	return &Delegate{
		cacheCh: make(chan []byte),
	}
}

// see Delegate
func (d *Delegate) NotifyMsg(cache []byte) {
	d.cacheCh <- cache
}

// see Delegate
func (d *Delegate) NodeMeta(limit int) []byte {
	return []byte("")
}

// see Delegate
func (d *Delegate) LocalState(join bool) []byte {
	// not use, noop
	return []byte("")
}

// see Delegate
func (d *Delegate) GetBroadcasts(overhead, limit int) [][]byte {
	// not use, noop
	return nil
}

// see Delegate
func (d *Delegate) MergeRemoteState(buf []byte, join bool) {
	// not use, noop
}

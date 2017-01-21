package lion

type multiPusher struct {
	pushers []Pusher
}

func newMultiPusher(pushers []Pusher) *multiPusher {
	w := make([]Pusher, len(pushers))
	copy(w, pushers)
	return &multiPusher{pushers}
}

func (m *multiPusher) Push(entry *Entry) error {
	var retErr error
	for _, pusher := range m.pushers {
		if err := pusher.Push(entry); err != nil {
			retErr = err
		}
	}
	return retErr
}

func (m *multiPusher) Flush() error {
	var retErr error
	for _, pusher := range m.pushers {
		if err := pusher.Flush(); err != nil {
			retErr = err
		}
	}
	return retErr
}

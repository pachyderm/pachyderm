package datum

type SetOption func(*Set)

func WithMetaOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.metaOutputClient = ptc
	}
}

func WithPFSOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.pfsOutputClient = ptc
	}
}

type DatumOption func(*Datum)

func WithRetry(numRetries int) DatumOption {
	return func(d *Datum) {
		d.numRetries = numRetries
	}
}

func WithRecoveryCallback(cb func() error) DatumOption {
	return func(d *Datum) {
		d.recoveryCallback = cb
	}
}

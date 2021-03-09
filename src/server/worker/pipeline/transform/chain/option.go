package chain

import "github.com/pachyderm/pachyderm/v2/src/server/worker/datum"

type JobChainOption func(*JobChain)

func WithBase(di datum.Iterator) JobChainOption {
	return func(jc *JobChain) {
		jc.base = di
	}
}

func WithNoSkip() JobChainOption {
	return func(jc *JobChain) {
		jc.noSkip = true
	}
}

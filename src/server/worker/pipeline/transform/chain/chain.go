
// Interface - put job into black box
// only for jobs in the running state
// black box returns datum.Iterator of datums to be processed as they are safe to be processed
// Notify black box when a job succeeds or fails so it can propagate datums to downstream jobs

type JobData interface {
	Iterator() (datum.Iterator, error)
}

type JobDatumIterator interface {
	func Next() bool
	func Datum() []*common.Input
	func NumAvailable() int
}

type RunningJobs interface {
	func Start(JobData) (JobDatumIterator, error)
	func Succeed(JobData, recoveredDatums DatumSet) error
	func Fail(JobData) error
}

type DatumSet map[string]struct{}{}

type jobDatumIterator struct {
	data JobData
	rj *runningJobs

	unyielded DatumSet // Datums that are waiting on an ancestor job
	yielding DatumSet // Datums that may be yielded as the iterator progresses
	yielded DatumSet // Datums that have been yielded

	ancestors []*jobDatumIterator
	dit datum.Iterator

	parent *jobDatumIterator // If set, this job is additive-only and must wait on its parent to finish before iteration can complete

	done chan struct{}
}

type runningJobs struct {
	mutex sync.Mutex
	jobs []*jobDatumIterator
	datumsBase DatumSet
}

func NewRunningJobs(driver driver.Driver) (RunningJobs, error) {
	datumsBase := make(
	// TODO: init datums base

	return &runningJobs{
		jobs: []*jobDatumIterator{},	
		datumsBase: datumsBase,
	}, nil
}


func (rj *runningJobs) Start(jd JobData) (JobDatumIterator, error) {
	dit, err := jd.Iterator()
	if err != nil {
		return nil, err
	}

	jdi := &jobDatumIterator {
		data: jd,
		rj: rj,
		unyielded: make(DatumSet),
		yielding: make(DatumSet),
		yielded: make(DatumSet),
		dit: dit,
	}

	rj.mutex.Lock()
	defer rj.mutex.Unlock()

	dit.Reset()
	for i := 0; i < dit.Len(); i++ {
		inputs := jdi.dit.DatumN(i)
		datumHash := rj.hash(inputs)
	}

	return jdi, nil
}

func (rj *runningJobs) hash(inputs []*common.Input) string {
	return common.HashDatum(rj.driver.PipelineInfo().Pipeline.Name, rj.driver.PipelineInfo().Salt, inputs)
}

func (rj *runningJobs) Fail(jd JobData) error {
	rj.mutex.Lock()
	defer rj.mutex.Unlock()

	index, err := rj.indexOf(jd)
	if err != nil {
		return err
	}


}

func (rj *runningJobs) Succeed(jd JobData, recoveredDatums DatumSet) (DatumSet, error) {
	rj.mutex.Lock()
	defer rj.mutex.Unlock()

	index, err := rj.indexOf(jd)
	if err != nil {
		return err
	}

	jdi := rj.jobs[index]

	if len(jdi.yielding) != 0 || len(jdi.unyielded) != 0 {
		return fmt.Errorf(
			"cannot advance a job with remaining datums: %d + %d of %d",
			len(unyielded), len(yielding), len(unyielded) + len(yielding) + len(yielded),
		)
	}

	newBaseDatums := make(DatumSet)
	// new base DatumSet will be yielded - recovered (if not additive-only), or base + yielded - recovered (if additive-only)
	for hash := range jdi.yielded {
		if _, ok := recoveredDatums[hash]; !ok {
			newBaseDatums[hash] = struct{}{}
		}
	}

	if jdi.additiveOnly {
		// Need to merge with ancestor's base datum set
	}

	return newBaseDatums, nil
}

func (jdi *jobDatumIterator) Next() bool {
	for {
		for len(jdi.yielding) == 0 {
			if len(jdi.unyielded) == 0 || len(jdi.ancestors) == 0 {
				return false
			}

			// Wait on an ancestor job
			cases := make([]reflect.SelectCase, len(jdi.ancestors))
			for _, x := range jdi.ancestors {
				cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(x.done)})
			}

			index, _, _ := reflect.Select(cases)
			finishedAncestor := jdi.ancestors[index]
			jdi.ancestors = append(jdi.ancestors[:index], jdi.ancestors[index+1:]...)

			// TODO: update 'yielding' from 'unyielded'

			jdi.dit.Reset()
		}

		for jdi.dit.Next() {
			inputs := jdi.dit.Datum()
			datumHash := jdi.rj.hash(inputs)
			if _, ok := jdi.yielding[datumHash]; ok {
				delete(jdi.yielding[datumHash])
				jdi.yielded[datumHash] = struct{}{}
				return true
			}
		}
		// TODO: assert that len(jdi.yielding) == 0
	}
}

func (jdi *jobDatumIterator) NumAvailable() int {
	return len(jdi.yielding)
}

func (jdi *jobDatumIterator) Datum() []*common.Input {
	return jdi.dit.Datum()
}
package karlsenstratum

import (
	"math/big"
	"sync"
	"time"

	"github.com/karlsen-network/karlsen-stratum-bridge/v2/src/gostratum"
	"github.com/karlsen-network/karlsend/v2/app/appmessage"
)

const maxjobs = 256

type MiningState struct {
	Jobs        map[int]*appmessage.RPCBlock
	JobLock     sync.Mutex
	jobCounter  int
	bigDiff     big.Int
	initialized bool
	useBigJob   bool
	connectTime time.Time
	stratumDiff *karlsenDiff
}

func MiningStateGenerator() any {
	return &MiningState{
		Jobs:        map[int]*appmessage.RPCBlock{},
		JobLock:     sync.Mutex{},
		connectTime: time.Now(),
	}
}

func GetMiningState(ctx *gostratum.StratumContext) *MiningState {
	return ctx.State.(*MiningState)
}

func (ms *MiningState) AddJob(job *appmessage.RPCBlock) int {
	ms.JobLock.Lock()
	ms.jobCounter++
	idx := ms.jobCounter
	ms.Jobs[idx%maxjobs] = job
	ms.JobLock.Unlock()
	return idx
}

func (ms *MiningState) GetJob(id int) (*appmessage.RPCBlock, bool) {
	ms.JobLock.Lock()
	job, exists := ms.Jobs[id%maxjobs]
	ms.JobLock.Unlock()
	return job, exists
}

package master

import (
	"context"
	"github.com/tiglabs/baud/proto/metapb"
	"github.com/tiglabs/baud/util/log"
	"runtime/debug"
	"sync"
	"time"
)

type WorkerManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	workers     map[string]Worker
	workersLock sync.RWMutex
	wg          sync.WaitGroup

	cluster *Cluster
}

func NewWorkerManager(cluster *Cluster) *WorkerManager {
	wm := &WorkerManager{
		workers: make(map[string]Worker),
		cluster: cluster,
	}
	wm.ctx, wm.cancel = context.WithCancel(context.Background())

	return wm
}

func (wm *WorkerManager) Start() error {
	wm.addWorker(NewSpaceStateTransitionWorker(wm.cluster))

	wm.workersLock.RLock()
	defer wm.workersLock.RUnlock()

	for _, worker := range wm.workers {
		wm.runWorker(worker)
	}

	return nil
}

func (wm *WorkerManager) Shutdown() {
	wm.cancel()
	wm.wg.Wait()
}

func (wm *WorkerManager) addWorker(worker Worker) {
	if worker == nil {
		return
	}

	wm.workersLock.Lock()
	defer wm.workersLock.Unlock()

	if _, ok := wm.workers[worker.getName()]; ok {
		log.Error("worker[%v] have already existed in worker manager.", worker.getName())
		return
	}

	wm.workers[worker.getName()] = worker
}

func (wm *WorkerManager) runWorker(worker Worker) {
	wm.wg.Add(1)
	go func() {
		defer wm.wg.Done()

		timer := time.NewTimer(worker.getInterval())
		defer timer.Stop()

		for {
			select {
			case <-wm.ctx.Done():
				return
			case <-timer.C:

				func() {
					log.Info("worker[%v] is running.", worker.getName())

					defer func() {
						if e := recover(); e != nil {
							log.Error("recover worker panic. e[%s] \nstack:[%s]", e, debug.Stack())
						}
					}()

					worker.run()
				}()

				timer.Reset(worker.getInterval())
			}
		}
	}()
}

type Worker interface {
	getName() string
	getInterval() time.Duration
	run()
}

type SpaceStateTransitionWorker struct {
	cluster *Cluster
}

func NewSpaceStateTransitionWorker(cluster *Cluster) *SpaceStateTransitionWorker {
	return &SpaceStateTransitionWorker{
		cluster: cluster,
	}
}

func (w *SpaceStateTransitionWorker) getName() string {
	return "Space State Transition Worker"
}

func (w *SpaceStateTransitionWorker) getInterval() time.Duration {
	return time.Second * 60
}

func (w *SpaceStateTransitionWorker) run() {
	dbs := w.cluster.dbCache.getAllDBs()
	for _, db := range dbs {
		spaces := db.spaceCache.getAllSpaces()
		for _, space := range spaces {

			func() {
				space.propertyLock.Lock()
				defer space.propertyLock.Unlock()

				if space.Status == metapb.SS_Init {

					var noReplica = false
					var searchPivot *Partition
					var searchNum = 100
					for {
						if searchPivot == nil {
							searchPivot = &Partition{
								Partition: &metapb.Partition{
									StartSlot: 0,
								},
							}
						}
						items := space.searchTree.ascendScan(searchPivot, searchNum)
						if items == nil || len(items) == 0 {
							break
						}

						for i := 0; i < len(items); i++ {
							itemPartition := items[i].partition
							if len(itemPartition.Replicas) == 0 {
								noReplica = true

								if err := GetPMSingle(nil).PushEvent(NewPartitionCreateEvent(itemPartition)); err != nil {
									log.Error("fail to push event for creating partition[%v].", itemPartition)
								}
							}
						}

						if len(items) < searchNum {
							break
						}
						searchPivot = items[len(items)-1].partition
					}

					if !noReplica {
						space.Status = metapb.SS_Running
					}
				}
			}()
		}
	}
}

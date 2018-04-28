package master

import (
	"github.com/tiglabs/baudengine/util/log"
	"sync"
)

type Master struct {
	config *Config

	cluster   *Cluster
	apiServer *ApiServer
	rpcServer *RpcServer

	processorManager *ProcessorManager
	workerManager    *WorkerManager
	wg               sync.WaitGroup
}

func NewServer() *Master {
	return new(Master)
}

func (ms *Master) Start(config *Config) error {
	ms.config = config

	ms.cluster = NewCluster(config)
	if err := ms.cluster.Start(); err != nil {
		log.Error("fail to start cluster. err:[%v]", err)
		return err
	}

	ms.rpcServer = NewRpcServer(config, ms.cluster)
	if err := ms.rpcServer.Start(); err != nil {
		log.Error("fail to start rpc server. err:[%v]", err)
		ms.cluster.Close()
		return err
	}

	ms.apiServer = NewApiServer(config, ms.cluster)
	if err := ms.apiServer.Start(); err != nil {
		log.Error("fail to start api server. err:[%v]", err)
		ms.rpcServer.Close()
		ms.cluster.Close()
		return err
	}

	GetPMSingle(ms.cluster).Start()
	GetPSRpcClientSingle(config)

	ms.workerManager = NewWorkerManager(ms.cluster)
	if err := ms.workerManager.Start(); err != nil {
		log.Error("fail to start worker manager. err:[%v]", err)
		ms.apiServer.Close()
		ms.rpcServer.Close()
		ms.cluster.Close()
		return err
	}

	return nil
}

func (ms *Master) Shutdown() {
	if ms.apiServer != nil {
		ms.apiServer.Close()
	}
	if ms.rpcServer != nil {
		ms.rpcServer.Close()
	}
	if ms.workerManager != nil {
		ms.workerManager.Shutdown()
	}
	GetPMSingle(nil).Stop()
	GetPSRpcClientSingle(nil).Close()

	if ms.cluster != nil {
		ms.cluster.Close()
	}
}

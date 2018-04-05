# Architecture

Baud is a document-graph distributed database. 

## Data Model

Entity, Concept/Category, Edge/Association, Space, DB/Graph

Each entity has an internal 'Unique ID' (UID) generated by the system, which is unique aross the entire graph. 

Entities as documents: UID -> JSON. 

Entity -instanceOf-> Class -subclassOf-> Class ...

Edges as documents: (UIDi, UIDj) -> JSON.

any attribute of entities or edges can be indexed and morever full-text search is a first-class citizen. 

## Components

master, raft replication for high availability

partitionserver (ps)

router

### cluster management

conainer-native - baud runs on Kubernetes clusters

## Partitioning

db -> entity or edge space -> partition -> slot

partition = slot id range

3 or more partitionservers form a replication group by means of raft. 

partition re-sharding is implemented through async filtered replication. 

multiple partitions of different spaces could be co-located on the same replicated partitionserver group. 


## Master

three/five/.. BM instances form a replicated BM service, or leverage a distributed coordination service like etcd/consul to store the metadata of Baud itself. 

we currently choose the former approach. 

* e.g. Start a master via cmd shell,

host2:$ baud -cm -http-addr host2:5001 -raft-addr host2:5002 -topo http://host1:5001 -data ~/node


### data structures

* database metadata

db (name -> id)

space (name -> id): entity or edge

partition (slot id range of (source) entity uid) : entity or edge

* cluster topo metadata

master nodes

ps nodes

router nodes

### persistence

marshalled and written to boltdb

### key operations

* Create a Space

0, foreach partition among the space

1, call JDOS to start several ps nodes;

2, ask the baudserver nodes to form a raft group as well as optional async replicas

3, call the raft leader to create a partition


* Split a Partition

0, call JDOS to start PS nodes

1, call the nodes to form two new raft groups

2, call the two raft leaders to setup async filtered replication with the original to-be-splitted partition leader

3, replicate

4, cutover

* Merge Partitions

0, call JDOS to start PS nodes

1, call the nodes to form a new raft groups

2, call the raft leader to setup async replication with the original to-be-merged partition leaders

3, replicate

4, cutover

* PS metrics reporting


* Router metrics reporting


## PS

Several PS nodes form a raft group, partitionserver group (PSG). And one PSG usualy serves a partition - a part of entity or edge space. 

### Inside a partition

for entity partition, UID -> Document; 
for edge partition, (UID1, UID2) -> Document;

* store

* indexing

* search


### Key Operations


## Router



## Manageability

Ops Center

Dashboard

### Monitoring

cluster-level statistics

space-level info

individual nodes

GC

SlowLog

### Deployment and Configration


### Upgrade


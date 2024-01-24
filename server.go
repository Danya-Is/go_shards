package main

import (
	"log"
	"math"
	"time"
)

type FileLogRecord struct {
	fileId  int
	shardId int
}

type Epoch struct {
	startTime   time.Time
	shardAmount int
}

type ServerState struct {
	epochs             []Epoch
	currentShardAmount int
}

type Server struct {
	state                 ServerState
	maxVirtualShardAmount int
	lastFileId            int
	fileLog               []FileLogRecord
	fileMap               map[int]time.Time
}

func (server *Server) hashFunc(id int) int {
	return id % server.maxVirtualShardAmount
}

func (server *Server) shardId(virtualShardId int, shardAmount int) int {
	return int(math.Floor(float64(virtualShardId) /
		(float64(server.maxVirtualShardAmount) / float64(shardAmount))))
}

func (server *Server) addFile(fileId int) {
	server.fileMap[fileId] = time.Now()
	server.fileLog = append(server.fileLog,
		FileLogRecord{
			fileId:  fileId,
			shardId: server.shardId(server.hashFunc(fileId), server.state.currentShardAmount),
		})
	log.Printf("File %d_%s added", fileId, server.fileMap[fileId].String())
}

func (server *Server) Init(maxVirtualShardAmount int) {
	server.maxVirtualShardAmount = maxVirtualShardAmount
}

func (server *Server) AddFiles(fileIds []int) {
	for _, fileId := range fileIds {
		server.addFile(fileId)
	}
	server.lastFileId += len(fileIds)
}

// AddShard todo add n shards
func (server *Server) AddShard() bool {
	newEpochs := make([]Epoch, len(server.state.epochs))
	copy(newEpochs, server.state.epochs)
	server.state.epochs = append(newEpochs, Epoch{time.Now(), server.state.currentShardAmount + 1})
	server.state.currentShardAmount++
	// todo err handle
	return true
}

func (server *Server) GetFile(fileId int) (shardId int) {
	if addTime, ok := server.fileMap[fileId]; ok {
		virtualShardId := server.hashFunc(fileId)

		for i := len(server.state.epochs) - 1; i >= 0; i-- {
			if server.state.epochs[i].startTime.Before(addTime) {
				return server.shardId(virtualShardId, server.state.epochs[i].shardAmount)
			}
		}

		log.Printf("Epoch for file %d not found", fileId)
		return -1
	} else {
		log.Printf("File %d not found", fileId)
		return -1
	}
}

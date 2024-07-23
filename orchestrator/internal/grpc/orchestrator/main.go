package orchestrator

import (
	pb "github.com/vingp/DistributedCalculator/proto"
)

type Server struct {
	pb.OrchestratorService // сервис из сгенерированного пакета
}

func NewServer() *Server {
	return &Server{}
}

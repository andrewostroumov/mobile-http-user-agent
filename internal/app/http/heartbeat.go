package http

import (
	"encoding/json"
	"net/http"
)

type beat struct {
	Status   string   `json:"status"`
	Database database `json:"db"`
}

type database struct {
	MaxConns   int32 `json:"max_connections"`
	TotalConns int32 `json:"total_connections"`
	IdleConns  int32 `json:"idle_connections"`
}

func (s *Server) heartbeat(w http.ResponseWriter, r *http.Request) {
	b := beat{
		Status: "ok",
		Database: database{
			TotalConns: s.repo.Stat().TotalConns(),
			MaxConns:   s.repo.Stat().MaxConns(),
			IdleConns:  s.repo.Stat().IdleConns(),
		},
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(&b)
}

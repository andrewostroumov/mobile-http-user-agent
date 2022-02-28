package http

import (
	"encoding/json"
	"github.com/andrewostroumov/mobile-http-user-agent/pkg/devices"
	"net/http"
)

type Agent struct {
	Mobile string `json:"mobile_chrome_user_agent"`
}

func (s *Server) userAgent(w http.ResponseWriter, r *http.Request) {
	device, err := s.repo.RandDevice()
	if err != nil {
		s.contextLogger.Errorf("rand device query error: %v", err)
		w.WriteHeader(500)
		return
	}

	a := Agent {
		Mobile: devices.BuildMobileAgent(device),
	}

	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(&a)
}

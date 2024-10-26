package handler

import (
	"sync/atomic"

	"github.com/finchrelia/chirpy-server/internal/database"
)

type APIConfig struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	JWT            string
	PolkaKey       string
}

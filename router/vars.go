package router

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbpool *pgxpool.Pool
var configFileTimeout = 5 * time.Second

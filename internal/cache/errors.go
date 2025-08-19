package cache

import "errors"

var (
	ErrExpiredCredentials = errors.New("expired credentials; do not serve from cache")
	ErrNearExpiration     = errors.New("too close to expiring; do not serve from cache")
)

package proxy

import "errors"

var (
	ErrWrongPongResponse   = errors.New("wrong pong response")
	ErrWrongSubject        = errors.New("subscribe on unexpected subject")
	ErrNoAvailableInstance = errors.New("no available instance")
	ErrChooseWrongBucket   = errors.New("error choose wrong bucket")
)

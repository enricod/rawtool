package main

import "time"

type WorkRequest struct {
	SourceFileName string
	Delay          time.Duration
}

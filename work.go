package main

import "time"

type WorkRequest struct {
	SourceImage MyImage
	Delay       time.Duration
}

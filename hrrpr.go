package main

import "io"

type Hrrpr interface {
	Get(string) error
	GetReader() io.Reader
}

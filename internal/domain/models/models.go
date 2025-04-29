package models

type Client struct {
	ID         int64
	IP         string
	Capacity   int
	RatePerSec int
}

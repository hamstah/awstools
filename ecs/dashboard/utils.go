package main

import "time"

func doEvery(d time.Duration, f func()) {
	f()
	for _ = range time.Tick(d) {
		f()
	}
}

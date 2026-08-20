package window

import "time"

// samplePoint holds one reading inside the ring buffer.
type samplePoint struct {
	tempC float64
	ts    time.Time
}

// ringBuffer is a time-ordered slice-backed buffer for sliding window samples.
type ringBuffer struct {
	points      []samplePoint
	tempScratch []float64
}

func newRingBuffer(capacity int) *ringBuffer {
	if capacity < 16 {
		capacity = 16
	}
	return &ringBuffer{points: make([]samplePoint, 0, capacity)}
}

// append adds a sample and returns points removed by expiry.
func (r *ringBuffer) append(p samplePoint, cutoff time.Time) []samplePoint {
	r.points = append(r.points, p)
	expired := r.expireBefore(cutoff)
	return expired
}

// expireBefore drops samples strictly older than cutoff.
func (r *ringBuffer) expireBefore(cutoff time.Time) []samplePoint {
	i := 0
	for i < len(r.points) && r.points[i].ts.Before(cutoff) {
		i++
	}
	if i == 0 {
		return nil
	}
	expired := make([]samplePoint, i)
	copy(expired, r.points[:i])
	r.points = r.points[i:]
	return expired
}

// each iterates current points oldest-first.
func (r *ringBuffer) each(fn func(samplePoint)) {
	for _, p := range r.points {
		fn(p)
	}
}

// len returns the number of buffered samples.
func (r *ringBuffer) len() int {
	return len(r.points)
}

// oldest returns the oldest timestamp or zero.
func (r *ringBuffer) oldest() time.Time {
	if len(r.points) == 0 {
		return time.Time{}
	}
	return r.points[0].ts
}

// newest returns the newest timestamp or zero.
func (r *ringBuffer) newest() time.Time {
	if len(r.points) == 0 {
		return time.Time{}
	}
	return r.points[len(r.points)-1].ts
}

// copyTemps returns a defensive copy of buffered temperatures.
func (r *ringBuffer) copyTemps() []float64 {
	if len(r.points) == 0 {
		return nil
	}
	out := make([]float64, len(r.points))
	for i, p := range r.points {
		out[i] = p.tempC
	}
	return out
}

// tempsAlias returns the internal temperature scratch used only by buggy builds.
// Healthy callers must use copyTemps.
func (r *ringBuffer) ensureTempScratch() []float64 {
	if cap(r.tempScratch) < len(r.points) {
		r.tempScratch = make([]float64, len(r.points))
	} else {
		r.tempScratch = r.tempScratch[:len(r.points)]
	}
	for i, p := range r.points {
		r.tempScratch[i] = p.tempC
	}
	return r.tempScratch
}

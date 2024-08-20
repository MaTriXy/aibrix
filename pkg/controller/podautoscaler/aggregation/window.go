/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aggregation

import (
	"math"
	"time"
)

/*
*
Referenced the knative implementation: pkg/autoscaler/aggregation/max/window.go,
but did not use the Ascending Minima Algorithm as we may need other aggregation functions beyond Max.
*/
type entry struct {
	value int32
	index int
}

// window is a sliding window that keeps track of recent {size} values.
type window struct {
	valueList     []entry
	first, length int
}

// newWindow creates a new window of specified size.
func newWindow(size int) *window {
	return &window{
		valueList: make([]entry, size),
	}
}

func (w *window) Size() int {
	return len(w.valueList)
}

func (w *window) index(i int) int {
	return i % w.Size()
}

// Record updates the window with a new value and index.
// It also removes all entries that are too old (index too small compared to the new index).
func (w *window) Record(value int32, index int) {
	// Remove elements that are outside the sliding window range.
	for w.length > 0 && w.valueList[w.first].index <= index-w.Size() {
		w.first = w.index(w.first + 1)
		w.length--
	}

	// Add the new value to the valueList.
	if w.length < w.Size() { // Ensure we do not exceed the buffer
		w.valueList[w.index(w.first+w.length)] = entry{value: value, index: index}
		w.length++
	}
}

// Max returns the maximum value in the current window.
func (w *window) Max() int32 {
	if w.length > 0 {
		maxValue := w.valueList[w.first].value
		for i := 1; i < w.length; i++ {
			valueIndex := w.index(w.first + i)
			if w.valueList[valueIndex].value > maxValue {
				maxValue = w.valueList[valueIndex].value
			}
		}
		return maxValue
	}
	return -1 // return a default value if no entries exist
}

// Min returns the minimum value in the current window.
func (w *window) Min() int32 {
	if w.length > 0 {
		minValue := w.valueList[w.first].value
		for i := 1; i < w.length; i++ {
			valueIndex := w.index(w.first + i)
			if w.valueList[valueIndex].value < minValue {
				minValue = w.valueList[valueIndex].value
			}
		}
		return minValue
	}
	return -1 // return a default value if no entries exist
}

type TimeWindow struct {
	window      *window
	granularity time.Duration
}

func NewTimeWindow(duration, granularity time.Duration) *TimeWindow {
	buckets := int(math.Ceil(float64(duration) / float64(granularity)))
	return &TimeWindow{window: newWindow(buckets), granularity: granularity}
}

func (t *TimeWindow) Record(now time.Time, value int32) {
	index := int(now.Unix()) / int(t.granularity.Seconds())
	t.window.Record(value, index)
}

func (t *TimeWindow) Max() int32 {
	return t.window.Max()
}

func (t *TimeWindow) Min() int32 {
	return t.window.Min()
}

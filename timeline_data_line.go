package prago

import (
	"fmt"
)

type timelineDataLine struct {
	Name     string
	StyleCSS string
	IsZero   bool
}

func getTimelineDataLines(min, max float64) (ret []*timelineDataLine) {
	maxTicks := 5

	if max-min < 3 {
		maxTicks = 3
	}

	ticks := CalculateGraphTicks(min, max, maxTicks)
	for _, v := range ticks {
		ret = append(ret, getTimelineDataLine(min, max, v))
	}

	return ret
}

func getTimelineDataLine(min, max float64, value float64) (ret *timelineDataLine) {

	maxDistance := max - min
	distanceFromMin := value - min
	var bottom float64
	if maxDistance != 0 {
		bottom = distanceFromMin / maxDistance
	}
	var isZero bool
	if value == 0 {
		isZero = true
	}
	return &timelineDataLine{
		Name:     humanizeFloat(value, "cs"),
		StyleCSS: fmt.Sprintf("bottom: %v%%;", bottom*100),
		IsZero:   isZero,
	}
}

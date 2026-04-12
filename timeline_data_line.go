package prago

import (
	"fmt"
)

type timelineDataLine struct {
	Name     string
	StyleCSS string
}

func getTimelineDataLines(min, max float64) (ret []*timelineDataLine) {
	maxTicks := 10

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
	bottom := distanceFromMin / maxDistance
	return &timelineDataLine{
		Name:     humanizeFloat(value, "cs"),
		StyleCSS: fmt.Sprintf("bottom: %v%%;", bottom*100),
	}
}

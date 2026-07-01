package prago

import (
	"fmt"
	"time"
)

func stringerString(in any) string {
	return in.(string)
}

func stringerInt64(in any) string {
	return fmt.Sprintf("%d", in.(int64))
}

func stringerFloat64(in any) string {
	return fmt.Sprintf("%f", in.(float64))
}

func stringerBool(in any) string {
	if in.(bool) {
		return "on"
	}
	return ""
}

func stringerDate(in any) string {
	tm := in.(time.Time)
	if tm.IsZero() {
		return ""
	}
	return tm.Format("2006-01-02")
}

func stringerDateTime(in any) string {
	tm := in.(time.Time)
	if tm.IsZero() {
		return ""
	}
	return tm.Format("2006-01-02 15:04")
}

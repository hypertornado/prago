package admin

import (
	"fmt"
	"github.com/hypertornado/prago"
	"os"
	"runtime"
	"strings"
	"time"
)

func stats(request prago.Request) {
	if !AuthenticateSysadmin(GetUser(request)) {
		Render403(request)
		return
	}

	stats := [][2]string{}

	stats = append(stats, [2]string{"App name", request.App().Data()["appName"].(string)})
	stats = append(stats, [2]string{"App version", request.App().Data()["version"].(string)})

	port := request.App().Data()["port"].(int)
	stats = append(stats, [2]string{"Port", fmt.Sprintf("%d", port)})

	developmentModeStr := "false"
	if request.App().DevelopmentMode {
		developmentModeStr = "true"
	}
	stats = append(stats, [2]string{"Development mode", developmentModeStr})
	stats = append(stats, [2]string{"Started at", request.App().StartedAt.Format(time.RFC3339)})

	stats = append(stats, [2]string{"Go version", runtime.Version()})
	stats = append(stats, [2]string{"Compiler", runtime.Compiler})
	stats = append(stats, [2]string{"GOARCH", runtime.GOARCH})
	stats = append(stats, [2]string{"GOOS", runtime.GOOS})
	stats = append(stats, [2]string{"GOMAXPROCS", fmt.Sprintf("%d", runtime.GOMAXPROCS(-1))})

	configStats := request.App().Config().Export()

	osStats := [][2]string{}
	osStats = append(osStats, [2]string{"EGID", fmt.Sprintf("%d", os.Getegid())})
	osStats = append(osStats, [2]string{"EUID", fmt.Sprintf("%d", os.Geteuid())})
	osStats = append(osStats, [2]string{"GID", fmt.Sprintf("%d", os.Getgid())})
	osStats = append(osStats, [2]string{"Page size", fmt.Sprintf("%d", os.Getpagesize())})
	osStats = append(osStats, [2]string{"PID", fmt.Sprintf("%d", os.Getpid())})
	osStats = append(osStats, [2]string{"PPID", fmt.Sprintf("%d", os.Getppid())})
	wd, _ := os.Getwd()
	osStats = append(osStats, [2]string{"Working directory", wd})
	hostname, _ := os.Hostname()
	osStats = append(osStats, [2]string{"Hostname", hostname})

	var mStats runtime.MemStats
	runtime.ReadMemStats(&mStats)
	memStats := [][2]string{}
	memStats = append(memStats, [2]string{"Alloc", fmt.Sprintf("%d", mStats.Alloc)})
	memStats = append(memStats, [2]string{"TotalAlloc", fmt.Sprintf("%d", mStats.TotalAlloc)})
	memStats = append(memStats, [2]string{"Sys", fmt.Sprintf("%d", mStats.Sys)})
	memStats = append(memStats, [2]string{"Lookups", fmt.Sprintf("%d", mStats.Lookups)})
	memStats = append(memStats, [2]string{"Mallocs", fmt.Sprintf("%d", mStats.Mallocs)})
	memStats = append(memStats, [2]string{"Frees", fmt.Sprintf("%d", mStats.Frees)})
	memStats = append(memStats, [2]string{"HeapAlloc", fmt.Sprintf("%d", mStats.HeapAlloc)})
	memStats = append(memStats, [2]string{"HeapSys", fmt.Sprintf("%d", mStats.HeapSys)})
	memStats = append(memStats, [2]string{"HeapIdle", fmt.Sprintf("%d", mStats.HeapIdle)})
	memStats = append(memStats, [2]string{"HeapInuse", fmt.Sprintf("%d", mStats.HeapInuse)})
	memStats = append(memStats, [2]string{"HeapReleased", fmt.Sprintf("%d", mStats.HeapReleased)})
	memStats = append(memStats, [2]string{"HeapObjects", fmt.Sprintf("%d", mStats.HeapObjects)})
	memStats = append(memStats, [2]string{"StackInuse", fmt.Sprintf("%d", mStats.StackInuse)})
	memStats = append(memStats, [2]string{"StackSys", fmt.Sprintf("%d", mStats.StackSys)})
	memStats = append(memStats, [2]string{"MSpanInuse", fmt.Sprintf("%d", mStats.MSpanInuse)})
	memStats = append(memStats, [2]string{"MSpanSys", fmt.Sprintf("%d", mStats.MSpanSys)})
	memStats = append(memStats, [2]string{"MCacheInuse", fmt.Sprintf("%d", mStats.MCacheInuse)})
	memStats = append(memStats, [2]string{"MCacheSys", fmt.Sprintf("%d", mStats.MCacheSys)})
	memStats = append(memStats, [2]string{"BuckHashSys", fmt.Sprintf("%d", mStats.BuckHashSys)})
	memStats = append(memStats, [2]string{"GCSys", fmt.Sprintf("%d", mStats.GCSys)})
	memStats = append(memStats, [2]string{"OtherSys", fmt.Sprintf("%d", mStats.OtherSys)})
	memStats = append(memStats, [2]string{"NextGC", fmt.Sprintf("%d", mStats.NextGC)})
	memStats = append(memStats, [2]string{"LastGC", fmt.Sprintf("%d", mStats.LastGC)})
	memStats = append(memStats, [2]string{"PauseTotalNs", fmt.Sprintf("%d", mStats.PauseTotalNs)})
	memStats = append(memStats, [2]string{"NumGC", fmt.Sprintf("%d", mStats.NumGC)})

	environmentStats := [][2]string{}
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		environmentStats = append(environmentStats, [2]string{pair[0], pair[1]})
	}

	request.SetData("stats", stats)
	request.SetData("configStats", configStats)
	request.SetData("osStats", osStats)
	request.SetData("memStats", memStats)
	request.SetData("environmentStats", environmentStats)
	request.SetData("admin_yield", "admin_stats")
	prago.Render(request, 200, "admin_layout")
}
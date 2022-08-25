package prago

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

func (app *App) initSystemStats() {
	startedAt := time.Now()

	app.Action("_stats").Name(unlocalized("Prago Stats")).Permission(sysadminPermission).Template("admin_systemstats").DataSource(
		func(request *Request) interface{} {

			stats := [][2]string{}
			stats = append(stats, [2]string{"App name", app.codeName})
			stats = append(stats, [2]string{"App version", app.version})

			developmentModeStr := "false"
			if app.developmentMode {
				developmentModeStr = "true"
			}
			stats = append(stats, [2]string{"Development mode", developmentModeStr})
			stats = append(stats, [2]string{"Started at", startedAt.Format(time.RFC3339)})

			stats = append(stats, [2]string{"Go version", runtime.Version()})
			stats = append(stats, [2]string{"Compiler", runtime.Compiler})
			stats = append(stats, [2]string{"GOARCH", runtime.GOARCH})
			stats = append(stats, [2]string{"GOOS", runtime.GOOS})
			stats = append(stats, [2]string{"GOMAXPROCS", fmt.Sprintf("%d", runtime.GOMAXPROCS(-1))})

			databaseStats := [][2]string{}
			dbStats := app.db.Stats()
			databaseStats = append(databaseStats, [2]string{"MaxOpenConnections", fmt.Sprintf("%d", dbStats.MaxOpenConnections)})
			databaseStats = append(databaseStats, [2]string{"OpenConnections", fmt.Sprintf("%d", dbStats.OpenConnections)})
			databaseStats = append(databaseStats, [2]string{"InUse", fmt.Sprintf("%d", dbStats.InUse)})
			databaseStats = append(databaseStats, [2]string{"Idle", fmt.Sprintf("%d", dbStats.Idle)})
			databaseStats = append(databaseStats, [2]string{"WaitCount", fmt.Sprintf("%d", dbStats.WaitCount)})
			databaseStats = append(databaseStats, [2]string{"WaitDuration", fmt.Sprintf("%v", dbStats.WaitDuration)})
			databaseStats = append(databaseStats, [2]string{"MaxIdleClosed", fmt.Sprintf("%d", dbStats.MaxIdleClosed)})
			databaseStats = append(databaseStats, [2]string{"MaxIdleTimeClosed", fmt.Sprintf("%d", dbStats.MaxIdleTimeClosed)})
			databaseStats = append(databaseStats, [2]string{"MaxLifetimeClosed", fmt.Sprintf("%d", dbStats.MaxLifetimeClosed)})

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

			/*if app.ElasticClient != nil {
				stats := app.ElasticClient.GetStats()
				if stats != nil {
					for _, indice := range stats.Indices {
						indice
					}
				}
			}*/

			ret := map[string]interface{}{}

			ret["roles"] = app.accessManager.roles
			ret["stats"] = stats
			ret["databaseStats"] = databaseStats
			ret["osStats"] = osStats
			ret["memStats"] = memStats
			ret["environmentStats"] = environmentStats
			ret["accessView"] = getResourceAccessView(app)
			ret["routes"] = app.mainController.router.export()
			return ret
		},
	)
}

type elasticStatsIndex struct {
	Name string
}

type accessView struct {
	Roles     []string
	Resources []accessViewResource
}

type accessViewResource struct {
	Name  string
	Roles []accessViewRole
}

type accessViewRole struct {
	Value string
}

func getResourceAccessView(app *App) accessView {
	ret := accessView{}
	for k := range app.accessManager.roles {
		ret.Roles = append(ret.Roles, k)
	}

	sort.Strings(ret.Roles)

	for _, resourceData := range app.resources {
		viewResource := accessViewResource{
			Name: resourceData.getID(),
		}
		for _, v := range ret.Roles {
			yeah := "+"
			no := "-"
			s := ""
			user := &user{Role: v}
			if app.authorize(user, resourceData.canView) {
				s += yeah
			} else {
				yeah = no
				s += no
			}
			if app.authorize(user, resourceData.canUpdate) {
				s += yeah
			} else {
				s += no
			}
			if app.authorize(user, resourceData.canCreate) {
				s += yeah
			} else {
				s += no
			}
			if app.authorize(user, resourceData.canDelete) {
				s += yeah
			} else {
				s += yeah
			}
			if app.authorize(user, resourceData.canExport) {
				s += yeah
			} else {
				s += no
			}
			viewResource.Roles = append(viewResource.Roles, accessViewRole{s})
		}
		ret.Resources = append(ret.Resources, viewResource)
	}

	return ret
}

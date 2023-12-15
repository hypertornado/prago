package prago

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func (app *App) initSystemStats() {
	startedAt := time.Now()

	sysadminBoard.Dashboard(unlocalized("Access view")).Table(func(r *Request) *Table {
		ret := app.Table()
		accessView := getResourceAccessView(app)

		header := []string{""}
		header = append(header, accessView.Roles...)

		ret.Header(header...)

		for _, resource := range accessView.Resources {
			var cells []*TableCell = []*TableCell{Cell(resource.Name)}
			for _, role := range resource.Roles {
				cells = append(cells, Cell(role.Value))
			}
			ret.Row(cells...)
		}
		return ret
	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("Auth roles")).Table(func(r *Request) *Table {
		ret := app.Table()

		roles := app.accessManager.roles
		for role, permission := range roles {
			var permStr string
			for k, v := range permission {
				if v {
					permStr += string(k) + " "
				}
			}
			ret.Row(Cell(role), Cell(permStr))
		}

		return ret

	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("Base app info")).Table(func(r *Request) *Table {

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

		return statsTable(app, stats)

	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("Database info")).Table(func(r *Request) *Table {
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

		return statsTable(app, databaseStats)

	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("OS info")).Table(func(r *Request) *Table {
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

		return statsTable(app, osStats)

	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("Memory info")).Table(func(r *Request) *Table {
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

		return statsTable(app, memStats)

	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("Enviroment variables")).Table(func(r *Request) *Table {

		environmentStats := [][2]string{}
		for _, e := range os.Environ() {
			pair := strings.Split(e, "=")
			environmentStats = append(environmentStats, [2]string{pair[0], pair[1]})
		}

		return statsTable(app, environmentStats)
	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("Elasticsearch")).Table(func(r *Request) *Table {
		ret := app.Table()
		esStats := elasticsearchStats(app)
		for _, v := range esStats {
			ret.Row(Cell(v[0]), Cell(v[1]), Cell(v[2]))
		}

		return ret
	}, sysadminPermission)

	sysadminBoard.Dashboard(unlocalized("Routes")).Table(func(r *Request) *Table {
		ret := app.Table()
		routes := app.mainController.router.export()
		for _, v := range routes {
			ret.Row(Cell(v[0]), Cell(v[1]))
		}

		return ret
	}, sysadminPermission)

}

func statsTable(app *App, data [][2]string) *Table {
	ret := app.Table()

	collator := collate.New(language.Czech)

	sort.Slice(data, func(i, j int) bool {
		if collator.CompareString(data[i][0], data[j][0]) <= 0 {
			return true
		} else {
			return false
		}
	})

	for _, v := range data {
		ret.Row(Cell(v[0]).Header(), Cell(v[1]))
	}
	return ret
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
		for _, role := range ret.Roles {
			yeah := "+"
			no := "-"
			s := ""
			if app.authorize(true, role, resourceData.canView) {
				s += yeah
			} else {
				yeah = no
				s += no
			}
			if app.authorize(true, role, resourceData.canUpdate) {
				s += yeah
			} else {
				s += no
			}
			if app.authorize(true, role, resourceData.canCreate) {
				s += yeah
			} else {
				s += no
			}
			if app.authorize(true, role, resourceData.canDelete) {
				s += yeah
			} else {
				s += yeah
			}
			if app.authorize(true, role, resourceData.canExport) {
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

func elasticsearchStats(app *App) [][3]string {
	client := app.ElasticSearchClient()
	if client == nil {
		return nil
	}

	stats, err := client.GetStats()
	if err != nil {
		panic(err)
	}

	var ret [][3]string

	var indiceNames []string
	for k := range stats.Indices {
		indiceNames = append(indiceNames, k)
	}

	sort.Strings(indiceNames)

	for _, v := range indiceNames {
		ret = append(ret, [3]string{
			v,
			fmt.Sprintf("%d docs", stats.Indices[v].Total.Docs.Count),
			fmt.Sprintf("%s size", byteCountSI(stats.Indices[v].Total.Store.SizeInBytes)),
		})
	}

	return ret

}

func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

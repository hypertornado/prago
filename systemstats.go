package prago

import (
	"fmt"
	"html/template"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

func (app *App) initSystemStats() {
	startedAt := time.Now()

	currentStatsDashboard := sysadminBoard.Dashboard(unlocalized("Requests"))

	currentStatsDashboard.Figure(unlocalized("Current requests"), "sysadmin").Value(func(request *Request) int64 {
		return currentRequestCounter.Load()
	}).RefreshTime(1)

	currentStatsDashboard.Figure(unlocalized("Total requests from server start"), "sysadmin").Value(func(request *Request) int64 {
		return totalRequestCounter.Load()
	}).RefreshTime(1)

	dbStatsDashboard := sysadminBoard.Dashboard(unlocalized("Database"))

	dbStatsDashboard.Figure(unlocalized("Idle"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().Idle)
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("OpenConnections"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().OpenConnections)
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("MaxOpenConnections"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxOpenConnections)
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("InUse"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().InUse)
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("WaitCount"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().WaitCount)
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("WaitDuration"), "sysadmin").ValueString(func(request *Request) string {
		return app.db.Stats().WaitDuration.String()
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("MaxIdleClosed"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxIdleClosed)
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("MaxIdleTimeClosed"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxIdleTimeClosed)
	}).RefreshTime(1)

	dbStatsDashboard.Figure(unlocalized("MaxLifetimeClosed"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxLifetimeClosed)
	}).RefreshTime(1)

	sysadminBoard.Dashboard(unlocalized("DB import")).Table(func(request *Request) *Table {
		ret := app.Table()

		dbConfig, err := getDBConfig(app.codeName)
		must(err)

		ret.Row(Cell(fmt.Sprintf("mysql -u %s -p%s -f -D %s < script.sql",
			dbConfig.User,
			dbConfig.Password,
			dbConfig.Name,
		)))
		return ret
	}, "sysadmin")

	sysadminBoard.Dashboard(unlocalized("pprof command")).Table(func(request *Request) *Table {
		ret := app.Table()

		ret.Row(Cell(app.getPprofProfilePath()))
		return ret
	}, "sysadmin")

	cacheInfoDashboard := sysadminBoard.Dashboard(unlocalized("Cache"))
	cacheInfoDashboard.Figure(unlocalized("Number of items"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.numberOfItems()
	}).RefreshTime(1)
	cacheInfoDashboard.Figure(unlocalized("Total requests"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.totalRequests.Load()
	}).RefreshTime(1)
	cacheInfoDashboard.Figure(unlocalized("Current requests"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.currentRequests.Load()
	}).RefreshTime(1)
	cacheInfoDashboard.Figure(unlocalized("Reload waiting"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.reloadWaiting.Load()
	}).RefreshTime(1)

	baseAppInfoDashboard := sysadminBoard.Dashboard(unlocalized("Base app info"))

	baseAppInfoDashboard.Figure(unlocalized("App name"), sysadminPermission).ValueString(func(r *Request) string {
		return app.codeName
	})
	baseAppInfoDashboard.Figure(unlocalized("App version"), sysadminPermission).ValueString(func(r *Request) string {
		return app.version
	})
	baseAppInfoDashboard.Figure(unlocalized("Development mode"), sysadminPermission).ValueString(func(r *Request) string {
		if app.developmentMode {
			return "true"
		}
		return "false"
	})
	baseAppInfoDashboard.Figure(unlocalized("Started at"), sysadminPermission).ValueString(func(r *Request) string {
		return startedAt.Format(time.RFC3339)
	})
	baseAppInfoDashboard.Figure(unlocalized("Go version"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.Version()
	})
	baseAppInfoDashboard.Figure(unlocalized("Compiler"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.Compiler
	})
	baseAppInfoDashboard.Figure(unlocalized("GOARCH"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.GOARCH
	})
	baseAppInfoDashboard.Figure(unlocalized("GOOS"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.GOOS
	})
	baseAppInfoDashboard.Figure(unlocalized("GOMAXPROCS"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", runtime.GOMAXPROCS(-1))
	})
	baseAppInfoDashboard.Figure(unlocalized("Localhost URL"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%s:%d", getLocalIP(), app.port)
	})

	osInfoDashboard := sysadminBoard.Dashboard(unlocalized("OS info"))

	osInfoDashboard.Figure(unlocalized("EGID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getegid())
	})
	osInfoDashboard.Figure(unlocalized("EUID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Geteuid())
	})
	osInfoDashboard.Figure(unlocalized("GID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getgid())
	})
	osInfoDashboard.Figure(unlocalized("Page size"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getpagesize())
	})
	osInfoDashboard.Figure(unlocalized("PID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getpid())
	})
	osInfoDashboard.Figure(unlocalized("PPID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getppid())
	})
	osInfoDashboard.Figure(unlocalized("Working directory"), sysadminPermission).ValueString(func(r *Request) string {
		wd, _ := os.Getwd()
		return wd
	})
	osInfoDashboard.Figure(unlocalized("Hostname"), sysadminPermission).ValueString(func(r *Request) string {
		hostname, _ := os.Hostname()
		return hostname
	})

	memoryInfoDashboard := sysadminBoard.Dashboard(unlocalized("Memory"))
	var mStatsExmaple runtime.MemStats
	fieldCount := reflect.TypeOf(mStatsExmaple).NumField()
	for i := 0; i < fieldCount; i++ {
		field := reflect.TypeOf(mStatsExmaple).Field(i)
		if field.Type.Kind() != reflect.Uint64 {
			continue
		}
		memoryInfoDashboard.Figure(unlocalized(field.Name), sysadminPermission).Value(func(r *Request) int64 {
			var mStats runtime.MemStats
			runtime.ReadMemStats(&mStats)
			uinvVal := reflect.ValueOf(mStats).Field(i).Uint()
			return int64(uinvVal)
		}).RefreshTime(1)

	}

	environmentDashboard := sysadminBoard.Dashboard(unlocalized("Enviroment variables"))
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		environmentDashboard.Figure(unlocalized(pair[0]), sysadminPermission).ValueString(func(r *Request) string {
			return pair[1]
		})

	}

	/*environmentDashboard.Table(func(r *Request) *Table {
		environmentStats := [][2]string{}
		for _, e := range os.Environ() {
			pair := strings.Split(e, "=")
			environmentStats = append(environmentStats, [2]string{pair[0], pair[1]})
		}

		return statsTable(app, environmentStats)
	}, sysadminPermission)*/

	ActionUI(app, "_routes", func(r *Request) template.HTML {
		ret := app.Table()
		routes := app.router.export()
		for _, v := range routes {
			ret.Row(Cell(v[0]), Cell(v[1]))
		}

		return ret.ExecuteHTML()
	}).Permission(sysadminPermission).Name(unlocalized("Routes")).Board(sysadminBoard)

	ActionUI(app, "_authorization", func(r *Request) template.HTML {
		ret := app.Table()
		accessView := getResourceAccessView(app)

		header := []string{""}
		header = append(header, accessView.Roles...)

		ret.Header(header...)

		for _, resource := range accessView.Resources {
			var cells []*TableCell = []*TableCell{Cell(resource.Name).Header()}
			for _, role := range resource.Roles {
				cells = append(cells, Cell(role.Value))
			}
			ret.Row(cells...)
		}

		ret.Table()

		roles := app.accessManager.roles
		for role, permission := range roles {
			var permStr string
			for k, v := range permission {
				if v {
					permStr += string(k) + " "
				}
			}
			ret.Row(Cell(role).Header(), Cell(permStr))
		}

		return ret.ExecuteHTML()
	}).Permission(sysadminPermission).Name(unlocalized("Authorization")).Board(sysadminBoard)

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

	for _, resource := range app.resources {
		viewResource := accessViewResource{
			Name: resource.getID(),
		}
		for _, role := range ret.Roles {
			yeah := "+"
			no := "-"
			s := ""
			if app.authorize(true, role, resource.canView) {
				s += yeah
			} else {
				yeah = no
				s += no
			}
			if app.authorize(true, role, resource.canUpdate) {
				s += yeah
			} else {
				s += no
			}
			if app.authorize(true, role, resource.canCreate) {
				s += yeah
			} else {
				s += no
			}
			if app.authorize(true, role, resource.canDelete) {
				s += yeah
			} else {
				s += yeah
			}
			if app.authorize(true, role, resource.canExport) {
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

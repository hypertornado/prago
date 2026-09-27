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

	currentStatsDashboard := sysadminBoard.AddDashboard(Unlocalized("Requests"))

	currentStatsDashboard.AddFigure(Unlocalized("Current requests"), "sysadmin").Value(func(request *Request) int64 {
		return currentRequestCounter.Load()
	}).RefreshTime(1)

	currentStatsDashboard.AddFigure(Unlocalized("Total requests from server start"), "sysadmin").Value(func(request *Request) int64 {
		return totalRequestCounter.Load()
	}).RefreshTime(1)

	currentStatsDashboard.AddFigure(Unlocalized("Form tasks"), "sysadmin").Value(func(request *Request) int64 {
		return int64(len(app.formTasksMap))
	}).RefreshTime(1)

	currentStatsDashboard.AddFigure(Unlocalized("Conflict checks"), "sysadmin").Value(func(request *Request) int64 {
		return int64(len(app.getCurrentConflicts()))
	}).RefreshTime(1).URL("/admin/_conflicts")

	dbStatsDashboard := sysadminBoard.AddDashboard(Unlocalized("Database"))

	dbStatsDashboard.AddFigure(Unlocalized("Idle"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().Idle)
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("OpenConnections"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().OpenConnections)
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("MaxOpenConnections"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxOpenConnections)
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("InUse"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().InUse)
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("WaitCount"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().WaitCount)
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("WaitDuration"), "sysadmin").ValueString(func(request *Request) string {
		return app.db.Stats().WaitDuration.String()
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("MaxIdleClosed"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxIdleClosed)
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("MaxIdleTimeClosed"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxIdleTimeClosed)
	}).RefreshTime(1)

	dbStatsDashboard.AddFigure(Unlocalized("MaxLifetimeClosed"), "sysadmin").Value(func(request *Request) int64 {
		return int64(app.db.Stats().MaxLifetimeClosed)
	}).RefreshTime(1)

	sysadminBoard.AddDashboard(Unlocalized("DB import")).AddTable(func(request *Request) *Table {
		ret := app.NewTable()

		dbConfig, err := GetDBConnectConfig(app.codeName)
		Must(err)

		ret.Row(Cell(fmt.Sprintf("mysql -u %s -p%s -f -D %s < script.sql",
			dbConfig.User,
			dbConfig.Password,
			dbConfig.Name,
		)))
		return ret
	}, "sysadmin")

	sysadminBoard.AddDashboard(Unlocalized("pprof command")).AddTable(func(request *Request) *Table {
		ret := app.NewTable()

		ret.Row(Cell(app.getPprofProfilePath()))
		return ret
	}, "sysadmin")

	cacheInfoDashboard := sysadminBoard.AddDashboard(Unlocalized("Cache"))
	cacheInfoDashboard.AddFigure(Unlocalized("Number of items"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.numberOfItems()
	}).RefreshTime(1)
	cacheInfoDashboard.AddFigure(Unlocalized("Total requests"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.totalRequests.Load()
	}).RefreshTime(1)
	cacheInfoDashboard.AddFigure(Unlocalized("Current requests"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.currentRequests.Load()
	}).RefreshTime(1)
	cacheInfoDashboard.AddFigure(Unlocalized("Reload waiting"), sysadminPermission).Value(func(r *Request) int64 {
		return app.cache.reloadWaiting.Load()
	}).RefreshTime(1)

	baseAppInfoDashboard := sysadminBoard.AddDashboard(Unlocalized("Base app info"))

	baseAppInfoDashboard.AddFigure(Unlocalized("App name"), sysadminPermission).ValueString(func(r *Request) string {
		return app.codeName
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("App version"), sysadminPermission).ValueString(func(r *Request) string {
		return app.version
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("Development mode"), sysadminPermission).ValueString(func(r *Request) string {
		if app.developmentMode {
			return "true"
		}
		return "false"
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("Started at"), sysadminPermission).ValueString(func(r *Request) string {
		return startedAt.Format(time.RFC3339)
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("Go version"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.Version()
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("Compiler"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.Compiler
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("GOARCH"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.GOARCH
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("GOOS"), sysadminPermission).ValueString(func(r *Request) string {
		return runtime.GOOS
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("GOMAXPROCS"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", runtime.GOMAXPROCS(-1))
	})
	baseAppInfoDashboard.AddFigure(Unlocalized("Localhost URL"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%s:%d", getLocalIP(), app.port)
	})

	osInfoDashboard := sysadminBoard.AddDashboard(Unlocalized("OS info"))

	osInfoDashboard.AddFigure(Unlocalized("EGID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getegid())
	})
	osInfoDashboard.AddFigure(Unlocalized("EUID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Geteuid())
	})
	osInfoDashboard.AddFigure(Unlocalized("GID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getgid())
	})
	osInfoDashboard.AddFigure(Unlocalized("Page size"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getpagesize())
	})
	osInfoDashboard.AddFigure(Unlocalized("PID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getpid())
	})
	osInfoDashboard.AddFigure(Unlocalized("PPID"), sysadminPermission).ValueString(func(r *Request) string {
		return fmt.Sprintf("%d", os.Getppid())
	})
	osInfoDashboard.AddFigure(Unlocalized("Working directory"), sysadminPermission).ValueString(func(r *Request) string {
		wd, _ := os.Getwd()
		return wd
	})
	osInfoDashboard.AddFigure(Unlocalized("Hostname"), sysadminPermission).ValueString(func(r *Request) string {
		hostname, _ := os.Hostname()
		return hostname
	})

	memoryInfoDashboard := sysadminBoard.AddDashboard(Unlocalized("Memory"))
	var mStatsExmaple runtime.MemStats
	fieldCount := reflect.TypeOf(mStatsExmaple).NumField()
	for i := 0; i < fieldCount; i++ {
		field := reflect.TypeOf(mStatsExmaple).Field(i)
		if field.Type.Kind() != reflect.Uint64 {
			continue
		}
		memoryInfoDashboard.AddFigure(Unlocalized(field.Name), sysadminPermission).Value(func(r *Request) int64 {
			var mStats runtime.MemStats
			runtime.ReadMemStats(&mStats)
			uinvVal := reflect.ValueOf(mStats).Field(i).Uint()
			return int64(uinvVal)
		}).RefreshTime(1)

	}

	environmentDashboard := sysadminBoard.AddDashboard(Unlocalized("Enviroment variables"))
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		environmentDashboard.AddFigure(Unlocalized(pair[0]), sysadminPermission).ValueString(func(r *Request) string {
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

	app.ActionUI("_routes", func(r *Request) template.HTML {
		ret := app.NewTable()
		routes := app.router.export()
		for _, v := range routes {
			ret.Row(Cell(v[0]), Cell(v[1]))
		}

		return ret.ExecuteHTML()
	}).Permission(sysadminPermission).Name(Unlocalized("Routes")).Board(sysadminBoard)

	app.ActionUI("_authorization", func(r *Request) template.HTML {
		ret := app.NewTable()
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

		ret.AddTable()

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
	}).Permission(sysadminPermission).Name(Unlocalized("Authorization")).Board(sysadminBoard)

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

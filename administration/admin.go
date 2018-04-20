package administration

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-commonmark/markdown"
	"github.com/gorilla/sessions"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/build"
	"github.com/hypertornado/prago/utils"
	"github.com/sendgrid/sendgrid-go"
	"html/template"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
)

//ErrItemNotFound is returned when no item is found
var ErrItemNotFound = errors.New("item not found")

//Admin is struct representing admin extension
type Administration struct {
	App              *prago.App
	Logo             string
	Background       string
	Prefix           string
	HumanName        string
	Resources        []*Resource
	resourceMap      map[reflect.Type]*Resource
	resourceNameMap  map[string]*Resource
	accessController *prago.Controller
	AdminController  *prago.Controller
	rootActions      []Action
	db               *sql.DB

	sendgridClient *sendgrid.SGClient
	noReplyEmail   string

	fieldTypes  map[string]FieldType
	javascripts []string
	css         []string
	roles       map[string]map[string]bool
}

//NewAdmin creates new administration on prefix url with name
func NewAdministration(app *prago.App, initFunction func(*Administration)) *Administration {
	admin := &Administration{
		App:              app,
		Prefix:           "/admin",
		HumanName:        app.AppName,
		Resources:        []*Resource{},
		resourceMap:      make(map[reflect.Type]*Resource),
		resourceNameMap:  make(map[string]*Resource),
		accessController: app.MainController().SubController(),

		db: connectMysql(app),

		sendgridClient: sendgrid.NewSendGridClientWithApiKey(app.Config.GetStringWithFallback("sendgridApi", "")),
		noReplyEmail:   app.Config.GetStringWithFallback("noReplyEmail", ""),

		fieldTypes:  make(map[string]FieldType),
		javascripts: []string{},
		css:         []string{},
		roles:       make(map[string]map[string]bool),
	}

	admin.AdminController = admin.accessController.SubController()

	admin.CreateResource(User{}, initUserResource)
	admin.CreateResource(File{}, initFilesResource)
	admin.CreateResource(activityLog{}, initActivityLog)

	admin.addDefaultFieldTypes()

	initFunction(admin)

	admin.accessController.AddBeforeAction(func(request prago.Request) {
		request.SetData("admin_header_prefix", admin.Prefix)
		request.SetData("background", admin.Background)
		request.SetData("javascripts", admin.javascripts)
		request.SetData("css", admin.css)
	})

	admin.accessController.AddAroundAction(
		createSessionAroundAction(
			admin.App.AppName,
			admin.App.Config.GetString("random"),
		),
	)

	googleApiKey := admin.App.Config.GetStringWithFallback("google", "")
	admin.AdminController.AddBeforeAction(func(request prago.Request) {
		request.SetData("google", googleApiKey)
	})

	bindDBBackupCron(admin.App)
	bindAPI(admin)

	admin.bindAdminCommand(admin.App)
	must(admin.initTemplates(admin.App))
	must(admin.App.LoadTemplateFromString(adminTemplates))
	bindStats(admin)

	admin.initRootActions()

	admin.AdminController.AddAroundAction(func(request prago.Request, next func()) {
		session := request.GetData("session").(*sessions.Session)
		userID, ok := session.Values["user_id"].(int64)

		if !ok {
			request.Redirect(admin.GetURL("user/login"))
			return
		}

		var user User
		err := admin.Query().WhereIs("id", userID).Get(&user)
		if err != nil {
			request.Redirect(admin.GetURL("user/login"))
			return

		}

		randomness := admin.App.Config.GetString("random")
		request.SetData("_csrfToken", user.CSRFToken(randomness))
		request.SetData("currentuser", &user)
		request.SetData("locale", GetLocale(request))

		headerData := admin.getHeaderData(request)
		request.SetData("admin_header", headerData)

		next()
	})

	admin.AdminController.Get(admin.GetURL(""), func(request prago.Request) {
		request.SetData("admin_header_home_selected", true)
		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getAdminNavigation(GetUser(request), ""),
			PageTemplate: "admin_home_navigation",
			PageData:     admin.getHomeData(request),
		})
	})

	admin.AdminController.Get(admin.GetURL("_help/markdown"), func(request prago.Request) {
		request.SetData("admin_yield", "admin_help_markdown")
		request.RenderView("admin_layout")
	})

	admin.AdminController.Get(admin.GetURL("_static/admin.js"), func(request prago.Request) {
		request.Response().Header().Set("Content-type", "text/javascript")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(adminJS))
	})
	admin.App.MainController().Get(admin.GetURL("_static/admin.css"), func(request prago.Request) {
		request.Response().Header().Set("Content-type", "text/css; charset=utf-8")
		request.Response().WriteHeader(200)
		request.Response().Write([]byte(adminCSS))
	})

	for _, resource := range admin.Resources {
		admin.initResource(resource)
	}

	admin.AdminController.Get(admin.GetURL("*"), func(request prago.Request) {
		render404(request)
	})

	admin.AddRole("sysadmin", admin.getSysadminPermissions())

	return admin
}

func (a Administration) GetURL(suffix string) string {
	ret := a.Prefix
	if len(suffix) > 0 {
		ret += "/" + suffix
	}
	return ret
}

func (a *Administration) AddAction(action Action) {
	a.rootActions = append(a.rootActions, action)
}

func (a *Administration) unsafeDropTables() error {
	for _, resource := range a.Resources {
		err := resource.unsafeDropTable()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Administration) AddFieldType(name string, fieldType FieldType) {
	a.fieldTypes[name] = fieldType
}

func (a *Administration) AddJavascript(url string) {
	a.javascripts = append(a.javascripts, url)
}

func (a *Administration) AddCSS(url string) {
	a.css = append(a.css, url)
}

//Migrate migrates all resource's tables
func (a *Administration) Migrate(verbose bool) error {
	tables, err := listTables(a.db)
	if err != nil {
		return err
	}
	for _, resource := range a.Resources {
		tables[resource.TableName] = false
		err := resource.migrate(verbose)
		if err != nil {
			return err
		}
	}

	if verbose {
		unusedTables := []string{}
		for k, v := range tables {
			if v == true {
				unusedTables = append(unusedTables, k)
			}
		}
		if len(unusedTables) > 0 {
			fmt.Printf("Unused tables: %s\n", strings.Join(unusedTables, ", "))
		}
	}

	return nil
}

//AddFlashMessage adds flash message to request
func AddFlashMessage(request prago.Request, message string) {
	session := request.GetData("session").(*sessions.Session)
	session.AddFlash(message)
	must(session.Save(request.Request(), request.Response()))
}

func (a *Administration) getResourceByName(name string) *Resource {
	return a.resourceNameMap[columnName(name)]
}

func (a *Administration) getDB() *sql.DB {
	return a.db
}

func (a *Administration) GetDB() *sql.DB {
	return a.getDB()
}

func (a *Administration) initRootActions() {
	for _, v := range a.rootActions {
		bindAction(a, nil, v, false)
	}
}

func (a *Administration) bindAdminCommand(app *prago.App) {
	app.AddCommand("admin", "migrate").Description("migrate database").
		Callback(func() {
			app.Log().Println("Migrating database")
			err := a.Migrate(true)
			if err == nil {
				app.Log().Println("Migrate done")
			} else {
				app.Log().Fatal(err)
			}
		})
}

func (a *Administration) initTemplates(app *prago.App) error {
	app.AddTemplateFunction("markdown", func(text string) template.HTML {
		return template.HTML(markdown.New(markdown.Breaks(true)).RenderToString([]byte(text)))
	})

	app.AddTemplateFunction("message", func(language, id string) template.HTML {
		return template.HTML(messages.Messages.Get(language, id))
	})

	app.AddTemplateFunction("thumb", func(ids string) string {
		for _, v := range strings.Split(ids, ",") {
			var image File
			err := a.Query().WhereIs("uid", v).Get(&image)
			if err == nil && image.IsImage() {
				return image.GetSmall()
			}
		}
		return ""
	})

	app.AddTemplateFunction("img", func(ids string) string {
		for _, v := range strings.Split(ids, ",") {
			var image File
			err := a.Query().WhereIs("uid", v).Get(&image)
			if err == nil && image.IsImage() {
				return image.GetLarge()
			}
		}
		return ""
	})

	return nil
}

func (resource Resource) GetItemURL(item interface{}, suffix string) string {
	ret := resource.GetURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func render403(request prago.Request) {
	request.SetData("message", messages.Messages.Get(GetLocale(request), "admin_403"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 403)
}

func render404(request prago.Request) {
	request.SetData("message", messages.Messages.Get(GetLocale(request), "admin_404"))
	request.SetData("admin_yield", "admin_message")
	request.RenderViewWithCode("admin_layout", 404)
}

func bindDBBackupCron(app *prago.App) {
	app.AddCronTask("backup db", func() {
		err := build.BackupApp(app)
		if err != nil {
			app.Log().Error("Error while creating backup:", err)
		}
	}, func(t time.Time) time.Time {
		return t.AddDate(0, 0, 1)
	})

	app.AddCronTask("remove old backups", func() {
		app.Log().Println("Removing old backups")
		deadline := time.Now().AddDate(0, 0, -7)
		backupPath := app.DotPath() + "/backups"
		files, err := ioutil.ReadDir(backupPath)
		if err != nil {
			app.Log().Println("error while removing old backups:", err)
			return
		}

		for _, file := range files {
			if file.ModTime().Before(deadline) {
				removePath := backupPath + "/" + file.Name()
				err := os.RemoveAll(removePath)
				if err != nil {
					app.Log().Println("Error while removing old backup file:", err)
				}
			}
		}
		app.Log().Println("Old backups removed")
	}, func(t time.Time) time.Time {
		return t.Add(1 * time.Hour)
	})
}

func columnName(fieldName string) string {
	return utils.PrettyURL(fieldName)
}

func Unlocalized(name string) func(string) string {
	return func(string) string {
		return name
	}
}

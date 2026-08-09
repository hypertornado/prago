package prago

import (
	"fmt"
	"sync"
	"time"
)

type settingsSingleton struct {
	settingsMap   map[string]*Setting
	settingsArray []*Setting
	resource      *Resource
	mutex         *sync.RWMutex
	cache         map[string]string
}

type pragoSettings struct {
	ID        int64
	Name      string
	Value     string `prago-type:"text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (app *App) initSettings() {
	app.settings = &settingsSingleton{
		settingsMap:   make(map[string]*Setting),
		settingsArray: []*Setting{},
		resource:      app.NewResource[pragoSettings](),
		mutex:         new(sync.RWMutex),
		cache:         make(map[string]string),
	}
	app.settings.resource.PermissionView("sysadmin")
	app.settings.resource.Board(sysadminBoard)

	must(app.settings.resource.migrate(false))
	initDefaultSettings(app)
}

func (app *App) NewSetting(id string, permission Permission) *Setting {

	s := app.settings.settingsMap[id]
	if s != nil {
		panic(fmt.Sprintf("setting %s already set", id))
	}

	setting := &Setting{
		app:        app,
		id:         id,
		name:       unlocalized(id),
		permission: permission,
	}

	app.settings.settingsMap[id] = setting
	app.settings.settingsArray = append(app.settings.settingsArray, setting)

	return setting
}

func (setting *Setting) Name(name func(string) string) *Setting {
	setting.name = name
	return setting
}

func (setting *Setting) DefaultValue(defaultValue string) *Setting {
	setting.defaultValue = defaultValue
	return setting
}

func (setting *Setting) ValueChangeCallback(fn func()) *Setting {
	setting.changeCallback = fn
	return setting
}

func (setting *Setting) GetValue() string {
	ret, err := setting.app.getSetting(setting.id)
	must(err)
	return ret
}

type Setting struct {
	app            *App
	id             string
	name           func(string) string
	permission     Permission
	defaultValue   string
	changeCallback func()
}

func (app *App) getSetting(id string) (string, error) {
	app.settings.mutex.RLock()
	defer app.settings.mutex.RUnlock()
	setting := app.settings.settingsMap[id]
	if setting == nil {
		return "", fmt.Errorf("can't find setting with id: %s", id)
	}

	cachedValue, ok := app.settings.cache[id]
	if ok {
		return cachedValue, nil
	}

	s := app.Query[pragoSettings]().Is("name", id).First()
	if s == nil {
		return setting.defaultValue, nil
	}

	app.settings.cache[id] = s.Value

	return s.Value, nil
}

func (app *App) mustGetSetting(id string) string {
	val, err := app.getSetting(id)
	must(err)
	return val
}

func (app *App) saveSetting(id, value string, request *Request) error {
	app.settings.mutex.Lock()
	defer app.settings.mutex.Unlock()

	setting := app.settings.settingsMap[id]
	if setting == nil {
		return fmt.Errorf("can't find setting with id: %s", id)
	}
	app.settings.cache = make(map[string]string)

	s := app.Query[pragoSettings]().Is("name", id).First()
	if s == nil {
		s = &pragoSettings{
			Name:  id,
			Value: value,
		}
		return request.CreateWithLog(s)
	} else {
		s.Value = value
		return request.UpdateWithLog(s)
	}
}

func initDefaultSettings(app *App) {
	app.NewSetting("random", "sysadmin").DefaultValue(randomString(20))
	app.NewSetting("google_key", "sysadmin")
	app.NewSetting("sendgrid_key", "sysadmin")
	app.NewSetting("no_reply_email", "sysadmin")
	app.NewSetting("port", "sysadmin").DefaultValue(fmt.Sprintf("%d", defaultPort))
	app.NewSetting("base_url", "sysadmin").DefaultValue("http://localhost:8585")
	app.NewSetting("ssh", "sysadmin")
	app.NewSetting("app_name", "sysadmin")
	app.NewSetting("background_image_url", "sysadmin")
	app.NewSetting("icon_image_url", "sysadmin")

	cdnCallback := func() {
		initCDN(app)
	}

	app.NewSetting("cdn_url", "sysadmin").
		DefaultValue("https://www.prago-cdn.com").
		ValueChangeCallback(cdnCallback)
	app.NewSetting("cdn_account", "sysadmin").
		DefaultValue(app.codeName).
		ValueChangeCallback(cdnCallback)
	app.NewSetting("cdn_password", "sysadmin").
		ValueChangeCallback(cdnCallback)

	app.name = func(string) string {
		ret := app.mustGetSetting("app_name")
		if ret != "" {
			return ret
		}
		return app.codeName
	}
}

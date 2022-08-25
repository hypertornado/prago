package prago

import (
	"fmt"
	"sync"
	"time"
)

type settingsSingleton struct {
	settingsMap   map[string]*setting
	settingsArray []*setting
	resource      *Resource[PragoSettings]
	mutex         *sync.RWMutex
	cache         map[string]string
}

type PragoSettings struct {
	ID        int64
	Name      string
	Value     string    `prago-type:"text" prago-preview:"true"`
	CreatedAt time.Time `prago-preview:"true"`
	UpdatedAt time.Time `prago-preview:"true"`
}

func (app *App) initSettings() {
	app.settings = &settingsSingleton{
		settingsMap:   make(map[string]*setting),
		settingsArray: []*setting{},
		resource:      NewResource[PragoSettings](app),
		mutex:         new(sync.RWMutex),
		cache:         make(map[string]string),
	}
	app.settings.resource.PermissionView("sysadmin")
	must(app.settings.resource.data.migrate(false))
	initDefaultSettings(app)
}

func (app *App) Setting(id string, permission Permission) *setting {

	s := app.settings.settingsMap[id]
	if s != nil {
		panic(fmt.Sprintf("setting %s already set", id))
	}

	setting := &setting{
		id:         id,
		name:       unlocalized(id),
		permission: permission,
	}

	app.settings.settingsMap[id] = setting
	app.settings.settingsArray = append(app.settings.settingsArray, setting)

	return setting
}

func (setting *setting) Name(name func(string) string) *setting {
	setting.name = name
	return setting
}

func (setting *setting) DefaultValue(defaultValue string) *setting {
	setting.defaultValue = defaultValue
	return setting
}

func (setting *setting) ValueChangeCallback(fn func()) *setting {
	setting.changeCallback = fn
	return setting
}

type setting struct {
	id             string
	name           func(string) string
	value          string
	permission     Permission
	defaultValue   string
	changeCallback func()
}

func (app *App) GetSetting(id string) (string, error) {
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

	s := app.settings.resource.Is("name", id).First()
	if s == nil {
		return setting.defaultValue, nil
	}

	app.settings.cache[id] = s.Value

	return s.Value, nil
}

func (app *App) MustGetSetting(id string) string {
	val, err := app.GetSetting(id)
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

	s := app.settings.resource.Is("name", id).First()
	if s == nil {
		s = &PragoSettings{
			Name:  id,
			Value: value,
		}
		return app.settings.resource.CreateWithLog(s, request)
	} else {
		s.Value = value
		return app.settings.resource.UpdateWithLog(s, request)
	}
}

func initDefaultSettings(app *App) {
	app.Setting("random", "sysadmin").DefaultValue(randomString(20))
	app.Setting("sendgrid_key", "sysadmin")
	app.Setting("no_reply_email", "sysadmin")
	app.Setting("port", "sysadmin").DefaultValue(fmt.Sprintf("%d", defaultPort))
	app.Setting("base_url", "sysadmin").DefaultValue("http://localhost:8585")
	app.Setting("ssh", "sysadmin")

	cdnCallback := func() {
		initCDN(app)
	}

	app.Setting("cdn_url", "sysadmin").
		DefaultValue("https://www.prago-cdn.com").
		ValueChangeCallback(cdnCallback)
	app.Setting("cdn_account", "sysadmin").
		DefaultValue(app.codeName).
		ValueChangeCallback(cdnCallback)
	app.Setting("cdn_password", "sysadmin").
		ValueChangeCallback(cdnCallback)

}

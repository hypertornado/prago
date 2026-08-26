package prago

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var otpLimiter = rate.NewLimiter(rate.Limit(100), 10)

const otpCoookieName = "prago_otp"

type otpKey struct {
	CookieID  string
	OTPKey    string
	User      *user
	CreatedAt time.Time
}

func (app *App) initOtp() {
	app.otpKeyMutex = &sync.Mutex{}
	app.otpKeyMap = map[string]*otpKey{}

	app.nologinFormAction("2f", func(form *Form, request *Request) {

		key := app.getOTPKey(request)
		if key == nil {
			request.Redirect("/admin/login")
			request.AddFlashMessage("Neplatné ověření")
			return
		}

		form.AddText("code", "Zadejte kód zaslaný na email "+key.User.Email).SetFocused().SetValue(request.Param("code"))

		if request.Param("code") != "" {
			form.AutosubmitFirstTime = true
		}

		form.AddSubmit("Ověřit kód")

	}, func(fv FormValidation, request *Request) {

		must(otpLimiter.Wait(request.r.Context()))

		key := app.getOTPKey(request)
		if key == nil {
			fv.AddError("Neplatné ověření")
			return
		}

		code := request.Param("code")
		code = strings.ReplaceAll(code, " ", "")
		code = strings.Trim(code, "\n\t\r")

		if code != key.OTPKey {
			fv.AddItemError("code", "Neplatný kód")
			return
		}

		fv.Redirect(app.doLogUser(request, key.User))
	})
}

func (app *App) createOTPKey(user *user, request *Request) *otpKey {
	app.otpKeyMutex.Lock()
	defer app.otpKeyMutex.Unlock()

	app.clearOTPs()

	if user.Email == "" {
		return nil
	}

	n := rand.Intn(1_000_000_000)
	if n < 100_000_000 {
		n += 100_000_000
	}

	key := &otpKey{
		CookieID:  randomString(64),
		OTPKey:    fmt.Sprintf("%d", n),
		User:      user,
		CreatedAt: time.Now(),
	}

	app.otpKeyMap[key.CookieID] = key

	err := app.Mailing("cs", func(md *MailingData) {
		md.AddRecipient(user.Name, user.Email)
		md.Name = "Ověřovací kód pro přihlášení - " + app.name(request.Locale())
		md.Description = template.HTML(fmt.Sprintf("Tento kód slouží pro přihlášení na web %s. S nikým ho nesdílejte.", app.BaseURL()))

		keyVal := key.OTPKey
		if len(keyVal) == 9 {
			keyVal = keyVal[:3] + " " + keyVal[3:6] + " " + keyVal[6:]
		}

		md.AddSection("", "").Code = keyVal

		md.Button = &Button{
			Name: "Potvrdit",
			URL:  fmt.Sprintf("%s/admin/user/2f?code=%s", app.BaseURL(), url.QueryEscape(key.OTPKey)),
		}

	})
	must(err)

	return key
}

func (app *App) getOTPKey(request *Request) *otpKey {
	app.otpKeyMutex.Lock()
	defer app.otpKeyMutex.Unlock()

	app.clearOTPs()

	cookie, err := request.r.Cookie(otpCoookieName)
	if err != nil || cookie == nil {
		return nil
	}
	return app.otpKeyMap[cookie.Value]
}

func (app *App) clearOTPs() {
	for k := range app.otpKeyMap {
		if app.otpKeyMap[k].CreatedAt.Before(time.Now().Add(-1 * time.Hour)) {
			delete(app.otpKeyMap, k)
		}
	}

}

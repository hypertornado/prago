package newsletter

import (
	"github.com/hypertornado/prago"
	administration "github.com/hypertornado/prago/extensions/admin"
	"time"
)

//https://github.com/chris-ramon/douceur
//https://github.com/aymerick/douceur

type NewsletterMiddleware struct {
	Admin       *administration.Admin
	SenderEmail string
}

func (nm NewsletterMiddleware) Init(app *prago.App) error {
	println("newsletter INIT")

	_, err := nm.Admin.CreateResource(Newsletter{})
	if err != nil {
		return err
	}

	_, err = nm.Admin.CreateResource(NewsletterPersons{})
	if err != nil {
		return err
	}

	return nil
}

type Newsletter struct {
	ID            int64
	Name          string `prago-preview:"true" prago-description:"Jméno newsletteru"`
	Subject       string
	Body          string `prago-type:"markdown"`
	PreviewSentAt time.Time
	SentAt        time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NewsletterPersons struct {
	ID           int64
	Name         string `prago-preview:"true" prago-description:"Jméno příjemce"`
	Email        string
	Confirmed    bool
	Unsubscribed bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

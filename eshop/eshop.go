package eshop

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/utils"
)

const maxNumberOfTickets = 10

var csrfToken = utils.RandomString(30)
var eshopInstance *Eshop

type Eshop struct {
	Configuration EshopConfiguration
}

type EshopConfiguration struct {
	BaseURL   string
	IDsPrefix int64
}

func InitEshop(admin *prago.App, config EshopConfiguration) (eshop *Eshop, err error) {
	if config.BaseURL == "" {
		panic("base url can't be empty")
	}
	if config.IDsPrefix <= 0 {
		panic("ids prefix wrongly set")
	}

	eshopInstance = &Eshop{
		Configuration: config,
	}
	admin.CreateResource(EshopProduct{}, initEshopProduct)
	admin.CreateResource(EshopOrder{}, initEshopOrder)
	admin.CreateResource(EshopTicket{}, initEshopTicket)
	return eshopInstance, nil
}

func initEshopProduct(resource *prago.Resource) {
	resource.HumanName = prago.Unlocalized("Eshop produkty")
	resource.CanView = "sysadmin"

	resource.AddItemAction(prago.CreateNavigationalItemAction(
		"generate",
		func(string) string { return "Generovat vstupenky" },
		"eshop_generate",
		func(resource prago.Resource, request prago.Request, user prago.User) interface{} {
			ret := map[string]interface{}{}

			options := []int{}
			for i := 1; i <= maxNumberOfTickets; i++ {
				options = append(options, i)
			}
			ret["CSRFToken"] = csrfToken
			ret["options"] = options
			return ret
		},
	))

	resource.AddItemAction(prago.Action{
		Permission: "sysadmin",
		URL:        "generate",
		Method:     "post",
		Handler: func(resource prago.Resource, request prago.Request, user prago.User) {

			if request.Params().Get("_csrfToken") != csrfToken {
				panic("wrong csrf token")
			}

			count, err := strconv.Atoi(request.Params().Get("count"))
			if err != nil {
				panic(err)
			}

			var product EshopProduct
			err = resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(&product)
			if err != nil {
				panic(err)
			}

			if count < 1 || count > maxNumberOfTickets {
				panic("wrong count")
			}

			var order EshopOrder
			order.User = user.ID

			err = resource.App.Create(&order)
			if err != nil {
				panic(err)
			}

			for i := 0; i < count; i++ {
				var ticket EshopTicket
				ticket.EshopOrder = order.ID
				ticket.EshopProduct = product.ID
				ticket.Price = product.Price
				ticket.Secret = getTicketSecret()
				err = resource.App.Create(&ticket)
				if err != nil {
					panic(err)
				}
			}

			prago.AddFlashMessage(request, "Vstupy vygenerovány")
			redirectURL := resource.App.GetAdminURL(fmt.Sprintf("eshoporder/%d", order.ID))
			request.Redirect(redirectURL)
		},
	})

	resource.AddItemAction(prago.CreateNavigationalItemAction(
		"control",
		prago.Unlocalized("Kontrolovat vstupenky"),
		"eshop_control",
		func(resource prago.Resource, request prago.Request, user prago.User) interface{} {
			ret := map[string]interface{}{}
			ret["CSRFToken"] = csrfToken
			//ret["options"] = options
			return ret
		},
	))

}

func initEshopOrder(resource *prago.Resource) {
	resource.HumanName = prago.Unlocalized("Eshop objednávky")
	resource.CanView = "sysadmin"
}

func initEshopTicket(resource *prago.Resource) {
	resource.HumanName = prago.Unlocalized("Eshop vstupenky")
	resource.CanView = "sysadmin"

	resource.AddItemAction(prago.Action{
		Name:       prago.Unlocalized("Stáhnout vstupenku"),
		Permission: "sysadmin",
		URL:        "vstupenka.pdf",
		Handler: func(resource prago.Resource, request prago.Request, user prago.User) {
			var ticket EshopTicket
			err := resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(&ticket)
			if err != nil {
				panic(err)
			}

			var product EshopProduct
			err = resource.App.Query().WhereIs("id", ticket.EshopProduct).Get(&product)

			qrCode := fmt.Sprintf("%s/vstupenka/%d/%s", eshopInstance.Configuration.BaseURL, ticket.PublicID(), ticket.Secret)

			request.Response().Header().Add("Content-type", "application/octet-stream")
			request.Response().WriteHeader(200)

			couponData := PDFCouponData{
				Name:        product.Name,
				Description: fmt.Sprintf("číslo vstupenky: %d\nkód: %s\nurl: %s", ticket.ID, ticket.Secret, qrCode),
				QRCode:      qrCode,
			}

			request.Response().Write(generatePDFCoupon([]PDFCouponData{couponData}))
		},
	})
}

type EshopProduct struct {
	ID          int64
	Banner      string `prago-type:"image" prago-preview:"true"`
	Logo        string `prago-type:"image" prago-preview:"true"`
	Name        string
	Hidden      bool
	Description string `prago-type:"text"`
	Text        string `prago-type:"markdown"`

	Quantity int64
	Price    int64

	OrderPosition int64 `prago-type:"order"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type EshopOrder struct {
	ID int64 `prago-preview:"true"`

	CustomerName  string `prago-preview:"true"`
	CustomerEmail string
	CustomerPhone string

	User int64 `prago-type:"relation" prago-preview:"true"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type EshopTicket struct {
	ID int64 `prago-preview:"true"`

	EshopProduct int64 `prago-type:"relation" prago-preview:"true"`
	EshopOrder   int64 `prago-type:"relation" prago-preview:"true"`

	Price int64

	Secret string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ticket EshopTicket) PublicID() int64 {
	return ticket.ID + eshopInstance.Configuration.IDsPrefix
}

func getTicketSecret() string {
	return utils.RandomString(8)
}

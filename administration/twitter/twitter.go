package twitter

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/hypertornado/prago"
)

type Account struct {
	consumer       string
	consumerSecret string
	access         string
	accessSecret   string
}

func (a Account) Tweet(text string) error {
	config := oauth1.NewConfig(a.consumer, a.consumerSecret)
	token := oauth1.NewToken(a.access, a.accessSecret)
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	_, _, err := client.Statuses.Update(text, nil)
	if err != nil {
		return err
	}

	return nil

}

func NewAccount(consumer, consumerSecret, access, accessSecret string) *Account {
	return &Account{
		consumer:       consumer,
		consumerSecret: consumerSecret,
		access:         access,
		accessSecret:   accessSecret,
	}
}

func newAccount(app *prago.App) (*Account, error) {
	keys := map[string]string{}
	for _, v := range []string{"consumer_key", "consumer_key_secret", "access_token", "access_token_secret"} {
		key := app.Config.GetStringWithFallback(v, "")
		if key == "" {
			return nil, fmt.Errorf("can't get config item '%s'", v)
		}
		keys[v] = key
	}
	return &Account{
		consumer:       keys["consumer_key"],
		consumerSecret: keys["consumer_key_secret"],
		access:         keys["access_token"],
		accessSecret:   keys["access_token_secret"],
	}, nil
}

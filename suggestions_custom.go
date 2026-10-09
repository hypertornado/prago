package prago

import (
	"strings"
	"sync"
)

type SuggestionsCustom struct {
	UUID          string
	handler       func(sr SuggestionsRequest) *SuggestionsResponse
	ContextFields string
	permission    Permission
}

func (app *App) initSuggestionsCustom() {
	app.suggestionsCustomMutex = &sync.RWMutex{}
	app.suggestionsCustom = map[string]*SuggestionsCustom{}
}

func SuggestionsResponseFromArray(in []string) *SuggestionsResponse {
	ret := &SuggestionsResponse{}
	for _, v := range in {

		ret.Suggestions = append(ret.Suggestions, &Suggestion{
			ID:   v,
			Name: v,
		})
	}
	return ret
}

func (app *App) Suggestions(permission Permission, contextFields []string, handler func(sr SuggestionsRequest) *SuggestionsResponse) *SuggestionsCustom {
	sc := &SuggestionsCustom{
		UUID:       randomString(20),
		handler:    handler,
		permission: permission,
	}

	var contextFieldItems []string
	for _, v := range contextFields {
		contextFieldItems = append(contextFieldItems, columnName(v))
	}
	sc.ContextFields = strings.Join(contextFieldItems, ",")

	app.suggestionsCustomMutex.Lock()
	defer app.suggestionsCustomMutex.Unlock()
	app.suggestionsCustom[sc.UUID] = sc
	return sc
}

func (app *App) getSuggestionCustom(uuid string) *SuggestionsCustom {
	app.suggestionsCustomMutex.RLock()
	defer app.suggestionsCustomMutex.RUnlock()
	return app.suggestionsCustom[uuid]
}

func (fi *FormItem) SetCustomSuggestion(suggestionCustom *SuggestionsCustom) *FormItem {
	fi.SuggestionsCustom = suggestionCustom
	if suggestionCustom != nil {
		fi.Autocomplete = "off"
	}

	return fi
}

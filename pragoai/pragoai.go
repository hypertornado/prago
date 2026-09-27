package pragoai

type Client struct {
	provider    string
	apiKey      string
	apiAdminKey string
}

func NewClaudeClient(apiKey string, apiAdminKey string) *Client {
	return &Client{
		provider:    "claude",
		apiKey:      apiKey,
		apiAdminKey: apiAdminKey,
	}
}

type Request struct {
	Text      string
	Model     string
	MaxTokens int64
	// ResponseDescription popisuje, co má odpověď obsahovat (např. formát,
	// jednotky, omezení). Pokud je vyplněné, přidá se do popisu nástroje,
	// kterým model odpověď vrací.
	ResponseDescription string
	// EnumValues omezuje odpověď typu string na jednu z uvedených hodnot.
	// Platí jen pro Get[string].
	EnumValues []string
	// Files jsou soubory přiložené k dotazu (obrázky nebo PDF).
	Files []File
	// Effort řídí míru interního uvažování modelu, jedna z hodnot
	// AvailableEfforts. Prázdná hodnota nechá výchozí chování modelu (high).
	// Model DefaultModelLow (Haiku) tento parametr nepodporuje a jeho
	// nastavení u něj skončí chybou z API.
	Effort string
}

// AvailableEfforts jsou povolené hodnoty pole Request.Effort.
var AvailableEfforts = []string{"low", "medium", "high", "xhigh", "max"}

// File je soubor přiložený k dotazu. MediaType musí být buď typ obrázku
// (např. "image/png", "image/jpeg", "image/gif", "image/webp"), nebo
// "application/pdf".
type File struct {
	Data      []byte
	MediaType string
}

type Model struct {
	ID   string
	Name string
}

const (
	DefaultModelLow    string = "claude-haiku-4-5-20251001"
	DefaultModelMiddle string = "claude-sonnet-5"
	DefaultModelHigh   string = "claude-opus-5"

	DefaultMaxTokens int64 = 16000
)

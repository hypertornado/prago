package development

var defaultPort = 8585

type DevelopmentSettings struct {
	LessDir    string
	LessTarget string
}

type MiddlewareDevelopment struct {
	Settings DevelopmentSettings
}

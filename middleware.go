package prago

//Middleware interface has Init method for middleware initialization
type Middleware interface {
	Init(*App) error
}

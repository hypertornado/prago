package selenium

type WindowTest struct {
	t      *Test
	Window *Window
}

func (w *WindowTest) GetSize() (width, height int) {
	return w.t.getInts(w.Window.GetSize())
}

func (w *WindowTest) SetSize(width, height int) {
	w.t.err(w.Window.SetSize(width, height))
}

func (w *WindowTest) GetPosition() (x, y int) {
	return w.t.getInts(w.Window.GetPosition())
}

func (w *WindowTest) SetPosition(x, y int) {
	w.t.err(w.Window.SetPosition(x, y))
}

func (w *WindowTest) Maximize() {
	w.t.err(w.Window.Maximize())
}

package selenium

import ()

type Window struct {
	Id      string
	session *Session
}

func (w *Window) GetSize() (width, height int, err error) {
	size := &size{}
	err = w.getValue("size", size)
	if err != nil {
		return -1, -1, err
	}
	return size.Width, size.Height, nil
}

func (w *Window) SetSize(width, height int) error {
	data := map[string]interface{}{
		"width":  width,
		"height": height,
	}
	return w.postRequest("size", data)
}

func (w *Window) GetPosition() (x, y int, err error) {
	position := &position{}
	err = w.getValue("position", position)
	if err != nil {
		return -1, -1, err
	}
	return position.X, position.Y, nil
}

func (w *Window) SetPosition(x, y int) error {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	return w.postRequest("position", data)
}

func (w *Window) Maximize() error {
	return w.postRequest("maximize", nil)
}

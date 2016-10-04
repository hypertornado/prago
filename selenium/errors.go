package selenium

import (
	"errors"
	"fmt"
)

var (
	UnknownErrorCode = errors.New("UnknownErrorCode")

	NoSuchDriver               = errors.New("NoSuchDriver")
	NoSuchElement              = errors.New("NoSuchElement")
	NoSuchFrame                = errors.New("NoSuchFrame")
	UnknownCommand             = errors.New("UnknownCommand")
	StaleElementReference      = errors.New("StaleElementReference")
	ElementNotVisible          = errors.New("ElementNotVisible")
	InvalidElementState        = errors.New("InvalidElementState")
	UnknownError               = errors.New("UnknownError")
	ElementIsNotSelectable     = errors.New("ElementIsNotSelectable")
	JavaScriptError            = errors.New("JavaScriptError")
	XPathLookupError           = errors.New("XPathLookupError")
	Timeout                    = errors.New("Timeout")
	NoSuchWindow               = errors.New("NoSuchWindow")
	InvalidCookieDomain        = errors.New("InvalidCookieDomain")
	UnableToSetCookie          = errors.New("UnableToSetCookie")
	UnexpectedAlertOpen        = errors.New("UnexpectedAlertOpen")
	NoAlertOpenError           = errors.New("NoAlertOpenError")
	ScriptTimeout              = errors.New("ScriptTimeout")
	InvalidElementCoordinates  = errors.New("InvalidElementCoordinates")
	IMENotAvailable            = errors.New("IMENotAvailable")
	IMEEngineActivationFailed  = errors.New("IMEEngineActivationFailed")
	InvalidSelector            = errors.New("InvalidSelector")
	SessionNotCreatedException = errors.New("SessionNotCreatedException")
	MoveTargetOutOfBounds      = errors.New("MoveTargetOutOfBounds")

	BadRequestError     = errors.New("BadRequestError")
	NotFoundError       = errors.New("NotFoundError")
	InternalServerError = errors.New("InternalServerError")
	NotImplementedError = errors.New("NotImplementedError")
)

func HTTPErrors(errorCode int) error {
	if errorCode >= 100 && errorCode < 400 {
		return nil
	}

	switch errorCode {
	case 400:
		return BadRequestError
	case 404:
		return NotFoundError
	case 500:
		return InternalServerError
	case 501:
		return NotImplementedError
	}
	return errors.New(fmt.Sprintf("HTTP error %d", errorCode))
}

func ErrorCodeToError(errorCode int) error {
	if errorCode == 0 {
		return nil
	}

	switch errorCode {
	case 6:
		return NoSuchDriver
	case 7:
		return NoSuchElement
	case 8:
		return NoSuchFrame
	case 9:
		return UnknownCommand
	case 10:
		return StaleElementReference
	case 11:
		return ElementNotVisible
	case 12:
		return InvalidElementState
	case 13:
		return UnknownError
	case 15:
		return ElementIsNotSelectable
	case 17:
		return JavaScriptError
	case 19:
		return XPathLookupError
	case 21:
		return Timeout
	case 23:
		return NoSuchWindow
	case 24:
		return InvalidCookieDomain
	case 25:
		return UnableToSetCookie
	case 26:
		return UnexpectedAlertOpen
	case 27:
		return NoAlertOpenError
	case 28:
		return ScriptTimeout
	case 29:
		return InvalidElementCoordinates
	case 30:
		return IMENotAvailable
	case 31:
		return IMEEngineActivationFailed
	case 32:
		return InvalidSelector
	case 33:
		return SessionNotCreatedException
	case 34:
		return MoveTargetOutOfBounds
	}

	return UnknownErrorCode
}

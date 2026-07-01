package prago

import (
	"fmt"
	"time"
)

var (
	messages         *messagesStruct = &messagesStruct{m}
	fallbackLanguage                 = "en"
)

type messagesStruct struct {
	m map[string]map[string]string
}

func (*messagesStruct) ItemsCount(i int64, locale string) (ret string) {
	if locale == "cs" {
		return itemsCountCS(i)
	}

	return itemsCountEN(i)
}

func itemsCountCS(i int64) (ret string) {
	if i == 1 {
		return fmt.Sprintf("%s položka", humanizeNumber(i))
	}
	if i >= 2 && i <= 4 {
		return fmt.Sprintf("%s položky", humanizeNumber(i))
	}
	return fmt.Sprintf("%s položek", humanizeNumber(i))
}

func itemsCountEN(i int64) (ret string) {
	if i == 1 {
		return fmt.Sprintf("%s item", humanizeNumber(i))
	}
	return fmt.Sprintf("%s items", humanizeNumber(i))
}

func (*messagesStruct) Timestamp(lang string, t time.Time, showTime bool) string {
	if t.IsZero() {
		return ""
	}
	var ret string
	if lang == "cs" {
		ret = t.Format("2. ") + czechMonth(t) + t.Format(" 2006")
	} else {
		ret = t.Format("2006-01-02")
	}
	if showTime {
		ret += " " + t.Format("15:04")
	}
	return ret
}

func czechMonth(date time.Time) string {
	months := []string{
		"ledna",
		"února",
		"března",
		"dubna",
		"května",
		"června",
		"července",
		"srpna",
		"září",
		"října",
		"listopadu",
		"prosince",
	}
	return months[date.Month()-1]
}

func (m *messagesStruct) Get(lang, id string, params ...any) string {
	ret := m.GetNullable(lang, id, params...)
	if ret == nil {
		ret = m.GetNullable(fallbackLanguage, id, params...)
	}
	if ret == nil {
		return id
	}
	return *ret
}

func (m *messagesStruct) GetNullable(lang, id string, params ...any) *string {
	loc, ok := m.m[id]
	if !ok {
		return nil
	}

	str, ok := loc[lang]
	if !ok {
		return nil
	}

	ret := fmt.Sprintf(str, params...)
	return &ret
}

func (m *messagesStruct) GetNameFunction(id string, params ...any) func(string) string {
	return func(lang string) string {
		return m.Get(lang, id, params...)
	}
}

var m = map[string]map[string]string{
	"yes": {
		"en": "✅ yes",
		"cs": "✅ ano",
	},
	"no": {
		"en": "no",
		"cs": "ne",
	},
	"yes_plain": {
		"en": "yes",
		"cs": "ano",
	},
	"no_plain": {
		"en": "no",
		"cs": "ne",
	},
	"empty": {
		"en": "empty",
		"cs": "prázdné",
	},
	"nonempty": {
		"en": "nonempty",
		"cs": "neprázdné",
	},

	"error": {
		"en": "Error happened",
		"cs": "Nastala chyba",
	},
	"admin": {
		"en": "Administration",
		"cs": "Administrace",
	},
	"signpost": {
		"en": "Signpost",
		"cs": "Rozcestník",
	},
	"signpost_long": {
		"en": "Signpost of %s",
		"cs": "Rozcestník administrace %s",
	},
	"log_out": {
		"en": "Log out",
		"cs": "Odhlásit se",
	},
	"new": {
		"en": "New item",
		"cs": "Nová položka",
	},
	"list": {
		"en": "List",
		"cs": "Seznam",
	},
	"edit": {
		"en": "Edit",
		"cs": "Upravit",
	},
	"view": {
		"en": "View",
		"cs": "Zobrazit",
	},
	"preview": {
		"en": "Preview",
		"cs": "Náhled",
	},
	"delete": {
		"en": "Delete",
		"cs": "Smazat",
	},
	"delete_confirmation": {
		"en": "Really want to delete this item?",
		"cs": "Opravdu chcete položku smazat?",
	},
	"delete_confirmation_name": {
		"en": "Really want to delete item „%s“?",
		"cs": "Opravdu chcete smazat položku „%s“?",
	},
	"back": {
		"en": "Back",
		"cs": "Zpět",
	},
	"create": {
		"en": "Create",
		"cs": "Vytvořit",
	},
	"export": {
		"en": "Export",
		"cs": "Export",
	},
	"stats": {
		"en": "Stats",
		"cs": "Statistiky",
	},
	"history": {
		"en": "Edits",
		"cs": "Úpravy",
	},
	"history_count": {
		"en": "Number of edits",
		"cs": "Počet úprav",
	},
	"history_last": {
		"en": "Last edited by",
		"cs": "Naposledy upraveno uživatelem",
	},
	"login_name": {
		"en": "Log into admin",
		"cs": "Přihlášení do administrace",
	},
	"email": {
		"en": "Email",
		"cs": "Email",
	},
	"email_or_username": {
		"en": "Username or email",
		"cs": "Uživatelské jméno nebo email",
	},
	"password": {
		"en": "Password",
		"cs": "Heslo",
	},
	"login_action": {
		"en": "Log in",
		"cs": "Přihlásit se",
	},
	"save": {
		"en": "Save",
		"cs": "Uložit",
	},
	"change": {
		"en": "Change",
		"cs": "Změnit",
	},
	"forgotten": {
		"en": "Forgotten password",
		"cs": "Zapomenuté heslo",
	},
	"register": {
		"en": "Create account",
		"cs": "Vytvořit nový účet",
	},
	"register_password": {
		"en": "Password",
		"cs": "Heslo",
	},
	"register_password_description": {
		"en": "At least 7 characters",
		"cs": "Alespoň 7 znaků",
	},
	"email_not_valid": {
		"en": "Invalid format of email",
		"cs": "Neplatný formát emailu.",
	},
	"email_already_registered": {
		"en": "User with this name already registered",
		"cs": "Uživatel s tímto emailem je již zaregistrován.",
	},
	"user_name_not_empty": {
		"en": "Username can't be empty",
		"cs": "Jméno uživatele nemůže být prázdné",
	},
	"validation_not_empty": {
		"en": "Item can't be empty",
		"cs": "Položka nemůže být prázdná",
	},
	"validation_value": {
		"en": "Wrong value of item",
		"cs": "Tato hodnota není povolená",
	},
	"validation_error": {
		"en": "Error while validating data",
		"cs": "Chyba při validaci dat",
	},
	"validation_date_format_error": {
		"en": "Wrong date format",
		"cs": "Špatný formát data",
	},
	"login_error": {
		"en": "Wrong user email or password.",
		"cs": "Špatný email, nebo heslo.",
	},
	"login_password_error": {
		"en": "Wrong password.",
		"cs": "Špatné heslo.",
	},
	"login_ok": {
		"en": "Log in was succesful",
		"cs": "Přihlášení proběhlo úspěšně",
	},
	"logout_ok": {
		"en": "User logged out",
		"cs": "Uživatel odhlášen",
	},
	"403": {
		"en": "Access denied",
		"cs": "Přístup zamítnut",
	},
	"404": {
		"en": "Page not found",
		"cs": "Stránka nenalezena",
	},
	"item_created": {
		"en": "Item created",
		"cs": "Položka byla vytvořena",
	},
	"item_edited": {
		"en": "Item edited",
		"cs": "Položka byla upravena",
	},
	"item_deleted": {
		"en": "Item deleted",
		"cs": "Položka byla smazána",
	},

	"Name": {
		"en": "Name",
		"cs": "Jméno",
	},
	"Description": {
		"en": "Description",
		"cs": "Popis",
	},
	"Image": {
		"en": "Image",
		"cs": "Obrázek",
	},
	"Hidden": {
		"en": "Hidden",
		"cs": "Skryté",
	},
	"CreatedAt": {
		"en": "Created At",
		"cs": "Vytvořeno",
	},
	"UpdatedAt": {
		"en": "Updated At",
		"cs": "Naposledy upraveno",
	},
	"OrderPosition": {
		"en": "Order position",
		"cs": "Pořadové číslo",
	},
	"File": {
		"en": "File",
		"cs": "Soubor",
	},
	"Place": {
		"en": "Place",
		"cs": "Místo",
	},

	"list_empty": {
		"en": "No items found",
		"cs": "Nic nebylo nalezeno.",
	},
	"tables": {
		"en": "Tables",
		"cs": "Tabulky",
	},
	"table": {
		"en": "Table",
		"cs": "Tabulka",
	},

	"settings": {
		"en": "Settings",
		"cs": "Nastavení",
	},
	"options": {
		"en": "Options",
		"cs": "Možnosti",
	},
	"settings_changed": {
		"en": "Settings changed",
		"cs": "Nastavení změněno",
	},
	"locale": {
		"en": "Locale",
		"cs": "Jazyk",
	},
	"user": {
		"en": "User",
		"cs": "Uživatel",
	},
	"users": {
		"en": "Users",
		"cs": "Uživatelé",
	},
	"password_change": {
		"en": "Change password",
		"cs": "Změnit heslo",
	},
	"password_old": {
		"en": "Old password",
		"cs": "Staré heslo",
	},
	"password_new": {
		"en": "New password",
		"cs": "Nové heslo",
	},
	"password_length": {
		"en": "Password must be at least 7 characters long",
		"cs": "Heslo musí mít alespoň 7 znaků",
	},
	"password_changed": {
		"en": "Password changed",
		"cs": "Heslo změněno",
	},

	"forgotten_name": {
		"en": "Renew forgotten password",
		"cs": "Obnovit zapomenuté heslo",
	},
	"forgotten_submit": {
		"en": "Send renew email",
		"cs": "Poslat email s instrukcemi",
	},
	"forgotten_email_subject": {
		"en": "Renew your password on %s",
		"cs": "Obnovit heslo pro %s",
	},
	"forgotten_email_body": {
		"en": "Forgot your password?<br><br>You can renew your password at <a href=\"%s\">%s</a><br><br>%s",
		"cs": "Zapomněli jste heslo?<br><br>Vaše heslo můžete obnovit na <a href=\"%s\">%s</a><br><br>%s",
	},
	"forgoten_sent": {
		"en": "Instructions for password renewal were send to %s",
		"cs": "Instrukce pro obnovení hesla poslána na email %s",
	},
	"forgoten_error": {
		"en": "Can't send password renewal instructions on %s",
		"cs": "Instrukce pro obnovu hesla nelze poslat na email %s",
	},
	"forgoten_set": {
		"en": "Set new password",
		"cs": "Nastavit nové heslo",
	},

	"confirm_email_subject": {
		"en": "Confirm your registration email",
		"cs": "Potvrďte svůj registrační email",
	},
	"confirm_button": {
		"en": "Confirm your email",
		"cs": "Potvrďte email",
	},
	"confirm_email_body": {
		"en": "Thanks for your registration.",
		"cs": "Děkujeme za registraci",
	},
	"confirm_email_ok": {
		"en": "Email confirmed",
		"cs": "Email potvrzen",
	},
	"confirm_email_fail": {
		"en": "Failed to confirm email",
		"cs": "Email se nepodařilo potvrdit",
	},
	"confirm_email_send": {
		"en": "Registration done. Please confirm your email %s",
		"cs": "Registrace hotova. Potvrďte prosím svůj email %s",
	},

	"flash_not_confirmed": {
		"en": "Your email is not confirmed. You can confirm it by clicking on link in your inbox.",
		"cs": "Váš email není potvrzen. Můžete ho potvrdit kliknutím na odkaz v emailu, který jsme vám poslali",
	},
	"flash_not_approved": {
		"en": "Your account waits to be approved by admin.",
		"cs": "Váš účet čeká na schválení administrátorem.",
	},

	"files": {
		"en": "Files",
		"cs": "Soubory",
	},
	"file": {
		"en": "File",
		"cs": "Soubor",
	},
	"width": {
		"en": "Width",
		"cs": "Šířka",
	},
	"height": {
		"en": "Height",
		"cs": "Výška",
	},
	"run": {
		"en": "Run",
		"cs": "Spustit",
	},
	"strong_connection_error": {
		"en": "Can't delete item with strong relation to '%s'",
		"cs": "Nelze smazat položku se silnou vazbou na '%s'",
	},
}

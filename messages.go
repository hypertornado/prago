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

func (m *messagesStruct) Get(lang, id string, params ...interface{}) string {
	ret := m.GetNullable(lang, id, params...)
	if ret == nil {
		ret = m.GetNullable(fallbackLanguage, id, params...)
	}
	if ret == nil {
		return id
	}
	return *ret
}

func (m *messagesStruct) GetNullable(lang, id string, params ...interface{}) *string {
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

func (m *messagesStruct) GetNameFunction(id string, params ...interface{}) func(string) string {
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

	"admin_error": {
		"en": "Error happened",
		"cs": "Nastala chyba",
	},
	"admin_admin": {
		"en": "Administration",
		"cs": "Administrace",
	},
	"admin_signpost": {
		"en": "Signpost",
		"cs": "Rozcestník",
	},
	"admin_signpost_long": {
		"en": "Signpost of %s",
		"cs": "Rozcestník administrace %s",
	},
	"boardpage": {
		"en": "Go to homepage",
		"cs": "Přejít na homepage",
	},

	"admin_log_out": {
		"en": "Log out",
		"cs": "Odhlásit se",
	},
	"admin_new": {
		"en": "New item",
		"cs": "Nová položka",
	},
	"admin_list": {
		"en": "List",
		"cs": "Seznam",
	},
	"admin_edit": {
		"en": "Edit",
		"cs": "Upravit",
	},
	"admin_view": {
		"en": "View",
		"cs": "Zobrazit",
	},
	"admin_preview": {
		"en": "Preview",
		"cs": "Náhled",
	},
	"admin_delete": {
		"en": "Delete",
		"cs": "Smazat",
	},
	"admin_delete_confirmation": {
		"en": "Really want to delete this item?",
		"cs": "Opravdu chcete položku smazat?",
	},
	"admin_delete_confirmation_name": {
		"en": "Really want to delete item %s?",
		"cs": "Opravdu chcete smazat položku %s?",
	},
	"admin_back": {
		"en": "Back",
		"cs": "Zpět",
	},
	"admin_create": {
		"en": "Create",
		"cs": "Vytvořit",
	},
	"admin_export": {
		"en": "Export",
		"cs": "Export",
	},
	"admin_stats": {
		"en": "Stats",
		"cs": "Statistiky",
	},
	"admin_history": {
		"en": "Edits",
		"cs": "Úpravy",
	},
	"admin_history_count": {
		"en": "Number of edits",
		"cs": "Počet úprav",
	},
	"admin_history_last": {
		"en": "Last edited by",
		"cs": "Naposledy upraveno uživatelem",
	},
	"admin_login_name": {
		"en": "Log into admin",
		"cs": "Přihlášení do administrace",
	},
	"admin_email": {
		"en": "Email",
		"cs": "Email",
	},
	"admin_email_or_username": {
		"en": "Username or email",
		"cs": "Uživatelské jméno nebo email",
	},
	"admin_password": {
		"en": "Password",
		"cs": "Heslo",
	},
	"admin_login_action": {
		"en": "Log in",
		"cs": "Přihlásit se",
	},
	"admin_save": {
		"en": "Save",
		"cs": "Uložit",
	},
	"admin_change": {
		"en": "Change",
		"cs": "Změnit",
	},
	"admin_forgotten": {
		"en": "Forgotten password",
		"cs": "Zapomenuté heslo",
	},
	"admin_register": {
		"en": "Create account",
		"cs": "Vytvořit nový účet",
	},
	"admin_register_password": {
		"en": "Password",
		"cs": "Heslo",
	},
	"admin_register_password_description": {
		"en": "At least 7 characters",
		"cs": "Alespoň 7 znaků",
	},
	"admin_email_not_valid": {
		"en": "Invalid format of email",
		"cs": "Neplatný formát emailu.",
	},
	"admin_email_already_registered": {
		"en": "User with this name already registered",
		"cs": "Uživatel s tímto emailem je již zaregistrován.",
	},
	"admin_user_name_not_empty": {
		"en": "Username can't be empty",
		"cs": "Jméno uživatele nemůže být prázdné",
	},
	"admin_validation_not_empty": {
		"en": "Item can't be empty",
		"cs": "Položka nemůže být prázdná",
	},
	"admin_validation_value": {
		"en": "Wrong value of item",
		"cs": "Tato hodnota není povolená",
	},
	"admin_validation_error": {
		"en": "Error while validating data",
		"cs": "Chyba při validaci dat",
	},
	"admin_validation_date_format_error": {
		"en": "Wrong date format",
		"cs": "Špatný formát data",
	},
	"admin_login_error": {
		"en": "Wrong user email or password.",
		"cs": "Špatný email, nebo heslo.",
	},
	"admin_login_password_error": {
		"en": "Wrong password.",
		"cs": "Špatné heslo.",
	},
	"admin_login_ok": {
		"en": "Log in was succesful",
		"cs": "Přihlášení proběhlo úspěšně",
	},
	"admin_logout_ok": {
		"en": "User logged out",
		"cs": "Uživatel odhlášen",
	},
	"admin_403": {
		"en": "Access denied",
		"cs": "Přístup zamítnut",
	},
	"admin_404": {
		"en": "Page not found",
		"cs": "Stránka nenalezena",
	},
	"admin_item_created": {
		"en": "Item created",
		"cs": "Položka byla vytvořena",
	},
	"admin_item_edited": {
		"en": "Item edited",
		"cs": "Položka byla upravena",
	},
	"admin_item_deleted": {
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
		"cs": "Datum vytvoření",
	},
	"UpdatedAt": {
		"en": "Updated At",
		"cs": "Datum poslední úpravy",
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

	"admin_list_empty": {
		"en": "No items found",
		"cs": "Nic nebylo nalezeno.",
	},
	"admin_tables": {
		"en": "Tables",
		"cs": "Tabulky",
	},
	"admin_table": {
		"en": "Table",
		"cs": "Tabulka",
	},

	"admin_settings": {
		"en": "Settings",
		"cs": "Nastavení",
	},
	"admin_options": {
		"en": "Options",
		"cs": "Možnosti",
	},
	"admin_options_visible": {
		"en": "Visible columns",
		"cs": "Viditelné sloupce",
	},
	"admin_settings_changed": {
		"en": "Settings changed",
		"cs": "Nastavení změněno",
	},
	"admin_locale": {
		"en": "Locale",
		"cs": "Jazyk",
	},
	"admin_user": {
		"en": "User",
		"cs": "Uživatel",
	},
	"admin_users": {
		"en": "Users",
		"cs": "Uživatelé",
	},
	"admin_password_change": {
		"en": "Change password",
		"cs": "Změnit heslo",
	},
	"admin_password_old": {
		"en": "Old password",
		"cs": "Staré heslo",
	},
	"admin_password_new": {
		"en": "New password",
		"cs": "Nové heslo",
	},
	"admin_password_length": {
		"en": "Password must be at least 7 characters long",
		"cs": "Heslo musí mít alespoň 7 znaků",
	},
	"admin_password_changed": {
		"en": "Password changed",
		"cs": "Heslo změněno",
	},

	"admin_forgotten_name": {
		"en": "Renew forgotten password",
		"cs": "Obnovit zapomenuté heslo",
	},
	"admin_forgotten_submit": {
		"en": "Send renew email",
		"cs": "Poslat email s instrukcemi",
	},
	"admin_forgotten_email_subject": {
		"en": "Renew your password on %s",
		"cs": "Obnovit heslo pro %s",
	},
	"admin_forgotten_email_body": {
		"en": "Forgot your password?<br><br>You can renew your password at <a href=\"%s\">%s</a><br><br>%s",
		"cs": "Zapomněli jste heslo?<br><br>Vaše heslo můžete obnovit na <a href=\"%s\">%s</a><br><br>%s",
	},
	"admin_forgoten_sent": {
		"en": "Instructions for password renewal were send to %s",
		"cs": "Instrukce pro obnovení hesla poslána na email %s",
	},
	"admin_forgoten_error": {
		"en": "Can't send password renewal instructions on %s",
		"cs": "Instrukce pro obnovu hesla nelze poslat na email %s",
	},
	"admin_forgoten_set": {
		"en": "Set new password",
		"cs": "Nastavit nové heslo",
	},

	"admin_confirm_email_subject": {
		"en": "Confirm your registration email on %s",
		"cs": "Potvrďte svůj registrační email na %s",
	},
	"admin_confirm_email_body": {
		"en": "Thanks for your registration,<br><br>you can confirm your email on <a href=\"%s\">%s</a>.<br><br>%s",
		"cs": "Děkujeme za registraci,<br><br>váš email můžete potvrdit na <a href=\"%s\">%s</a>.<br><br>%s",
	},
	"admin_confirm_email_ok": {
		"en": "Email confirmed",
		"cs": "Email potvrzen",
	},
	"admin_confirm_email_fail": {
		"en": "Failed to confirm email",
		"cs": "Email se nepodařilo potvrdit",
	},
	"admin_confirm_email_send": {
		"en": "Registration done. Please confirm your email %s",
		"cs": "Registrace hotova. Potvrďte prosím svůj email %s",
	},

	"admin_flash_not_confirmed": {
		"en": "Your email is not confirmed. You can confirm it by clicking on link in your inbox.",
		"cs": "Váš email není potvrzen. Můžete ho potvrdit kliknutím na odkaz v emailu, který jsme vám poslali",
	},
	"admin_flash_not_approved": {
		"en": "Your account waits to be approved by admin.",
		"cs": "Váš účet čeká na schválení administrátorem.",
	},

	"admin_files": {
		"en": "Files",
		"cs": "Soubory",
	},
	"admin_file": {
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
	"tasks": {
		"en": "Tasks",
		"cs": "Úlohy",
	},
	"tasks_run": {
		"en": "Run tasks",
		"cs": "Spustit úlohy",
	},
	"tasks_runned": {
		"en": "Runned tasks",
		"cs": "Spuštěné úlohy",
	},
	"run": {
		"en": "Run",
		"cs": "Spustit",
	},
}

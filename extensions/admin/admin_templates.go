package admin
const TEMPLATES = `
{{define "admin_edit"}}

<h2>{{message .locale "admin_edit"}} - {{.admin_item.Name}}</h2>

<a href="../{{.admin_resource.ID}}">{{message .locale "admin_back"}}</a>

{{tmpl "admin_form" .admin_form}}

{{end}}{{define "admin_flash"}}
{{if .flash_messages}}
<div class="flash">
{{range $message := .flash_messages}}
  <div class="flash_message">{{$message}}</div>
{{end}}
</div>
{{end}}
{{end}}{{define "admin_form"}}

<form method="{{.Method}}" action="{{.Action}}" class="form" enctype="multipart/form-data" novalidate>

{{if .Errors}}
  <div class="form_errors">
    {{range $error := .Errors}}
      <div class="form_errors_error">{{$error}}</div>
    {{end}}
  </div>
{{end}}

{{range $item := .Items}}

  {{if $item.Template}}
    {{tmpl $item.Template $item}}
  {{else}}
    <label class="form_label{{if .Errors}} form_label-errors{{end}}{{if .Required}} form_label-required{{end}}">
      {{if eq .HiddenName false}}
        <span class="form_label_text">{{.NameHuman}}</span>
      {{end}}
      {{if .Errors}}
        <div class="form_label_errors">
          {{range $error := .Errors}}
            <div class="form_label_errors_error">{{$error}}</div>
          {{end}}
        </div>
      {{end}}
      {{tmpl $item.SubTemplate $item}}
    </label>
  {{end}}
{{end}}
</form>

{{end}}{{define "admin_home"}}

<h2>{{.admin_header.appName}}</h2>

<table class="admin_table">
{{range $item := .admin_header.menu}}
  <tr>
    <td style="width: 100%;">
      <a href="{{$item.url}}">{{$item.name}}</a>
    </td>
  </tr>
{{end}}
</table>

{{end}}{{define "admin_item_input"}}
  <input name="{{.Name}}" value="{{.Value}}" class="input form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_email"}}
  <input name="{{.Name}}" value="{{.Value}}" type="email" class="input form_input" spellcheck="false"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_password"}}
  <input name="{{.Name}}" value="{{.Value}}" type="password" class="input form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_textarea"}}
  <textarea name="{{.Name}}" class="input form_input textarea"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>{{.Value}}</textarea>
{{end}}

{{define "admin_item_checkbox"}}
  <input type="checkbox" name="{{.Name}}" {{if .Value}}checked{{end}}{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
  <span class="form_label_text-inline">{{.NameHuman}}</span>
{{end}}

{{define "admin_item_date"}}
  <input type="date" name="{{.Name}}" value="{{.Value}}" class="input form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_timestamp"}}
  <input placeholder="Example: 2001-12-06 20:30" name="{{.Name}}" value="{{.Value}}" class="input form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_image"}}
  {{if .Value}}
  <img src="/img/200x0/{{.Value}}.jpg" style="max-width: 100px; max-height: 100px; display: block; margin: 5px;">
  {{end}}

  {{if .Readonly}}
  <input type="file" name="{{.Name}}" accept=".jpeg,.jpg" class="input form_input"{{if .Focused}} autofocus{{end}}>
  {{end}}
{{end}}

{{define "admin_item_submit"}}
  <input type="submit" name="{{.Name}}" value="{{.NameHuman}}" class="btn"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_select"}}
  <select name="{{.Name}}" class="input form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
    {{$val := .Value}}
    {{range $value := .Values}}
      <option value="{{index $value 0}}"{{if eq $val (index $value 0)}} selected{{end}}>{{index $value 1}}</option>
    {{end}}
  </select>
{{end}}

{{define "admin_item_hidden"}}
<input type="hidden" name="{{.Name}}" value="{{.Value}}">
{{end}}


{{define "admin_string"}}
{{.}}
{{end}}
{{define "admin_layout"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>Admin</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <link rel="stylesheet" href="{{.admin_header.prefix}}/admin.css">
  </head>
  <body class="admin">
    {{tmpl "admin_flash" .}}
    <div class="admin_header">
        <ul class="admin_header_list admin_header_list-right">
            <li><a href="{{.admin_header.prefix}}/user/settings">{{.currentuser.Email}}</a> | <a href="{{.admin_header.prefix}}/logout?_csrfToken={{._csrfToken}}">{{message .locale "admin_log_out"}}</a></li>
        </ul>

        <h1><a href="{{.admin_header.prefix}}">{{.admin_header.appName}}</a></h1>


        {{ $admin_resource := .admin_resource }}

        <ul class="admin_header_list">
            <li><a href="/" style="text-decoration: none;">🌐</a></li>
            {{range $item := .admin_header.menu}}
                <li class="{{if $admin_resource}}{{ if eq $admin_resource.ID $item.id }}admin_header_item-active{{end}}{{end}}">
                    <a href="{{$item.url}}">{{$item.name}}</a>
                </li>
            {{end}}
        </ul>
    </div>

    <div class="admin_content">

    {{if .template_before}}{{tmpl .template_before .}}{{end}}

    {{tmpl .admin_yield .}}

    {{if .template_after}}{{tmpl .template_after .}}{{end}}

    </div>
    
  </body>
</html>

{{end}}{{define "admin_list"}}

<h2>{{call .admin_resource.Name .locale}}</h2>

{{$adminResource := .admin_resource }}
{{$locale := .locale}}

{{$csrfToken := ._csrfToken}}

<a href="{{.admin_resource.ID}}/new" class="btn">{{message .locale "admin_new"}}</a>

<table class="admin_table">
  <tr>
  {{range $item := .admin_list_table_data.Header}}
    <th>{{$item.NameHuman}}</th>
  {{end}}
  <th colspan="2"></th>
  </tr>
{{range $item := .admin_list_table_data.Rows}}
  <tr>
    {{range $cell := $item.Items}}
    <td>
      {{ tmpl $cell.TemplateName $cell.Value }}
    </td>
    {{end}}
    <td nowrap>
      <a href="{{ $adminResource.ID}}/{{$item.ID}}" class="btn">{{message $locale "admin_edit"}}</a> 
    </td>
    <td nowrap>
      <form method="POST" action="{{ $adminResource.ID}}/{{$item.ID}}/delete?_csrfToken={{$csrfToken}}" onsubmit="return window.confirm('{{message $locale "admin_delete_confirmation"}}');">
        <input type="submit" value="{{message $locale "admin_delete"}}" class="btn">
      </form>
    </td>
  </tr>
{{end}}
</table>

{{end}}{{define "admin_login"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{.title}}</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{.admin_header_prefix}}/admin.css">
  </head>
  <body class="admin">
    {{tmpl "admin_flash" .}}

    <div class="admin_content">
    <h2>{{.title}}</h2>

    {{tmpl "admin_form" .admin_form}}

    <div style="text-align: center">
    {{Plain .bottom}}
    </div>

    </div>
  </body>
</html>

{{end}}{{define "admin_message"}}
<h1>{{.message}}</h1>
{{end}}{{define "admin_new"}}

<h2>{{message .locale "admin_new"}} - {{.admin_resource.Name}}</h2>

<a href="../{{.admin_resource.ID}}">{{message .locale "admin_back"}}</a>

{{tmpl "admin_form" .admin_form}}

{{end}}{{define "admin_new_user"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{.name}} - {{message .locale "admin_login_name"}}</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{.admin_header_prefix}}/admin.css">
  </head>
  <body class="admin">
    <div class="admin_content">
    <h2>{{.name}}CREATE - {{message .locale "admin_login_name"}}</h2>

    <form class="form" method="POST">
        <label class="form_label">
          <span class="form_label_text">{{message .locale "admin_email"}}</span>
          <input type="email" name="email" autofocus class="input form_input">
        </label>

        <label class="form_label">
          <span class="form_label_text">{{message .locale "admin_password"}}</span>
          <input type="password" name="password" class="input form_input">
        </label>

        <input type="submit" value="{{message .locale "admin_login_action"}}" class="btn">

    </form>

    <a href="{{.admin_header_prefix}}/user/login">Log In</a>

    </div>
  </body>
</html>

{{end}}{{define "admin_settings"}}

<h2>{{message .locale "admin_settings"}}</h2>

{{tmpl "admin_form" .admin_form}}

{{end}}`


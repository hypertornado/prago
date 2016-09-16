package admin
const TEMPLATES = `
{{define "admin_edit"}}

<h2>{{.admin_title}}</h2>

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

{{end}}{{define "admin_form_view"}}

<h2>{{.admin_title}}</h2>

{{tmpl "admin_form" .admin_form}}

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

{{define "admin_item_markdown"}}
  <div class="admin_markdown">
    <textarea name="{{.Name}}" class="input form_input textarea"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>{{.Value}}</textarea>
    <div class="admin_markdown_preview"></div>
  </div>
{{end}}

{{define "admin_item_checkbox"}}
  <input type="checkbox" name="{{.Name}}" {{if .Value}}checked{{end}}{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
  <span class="form_label_text-inline">{{.NameHuman}}</span>
{{end}}

{{define "admin_item_date"}}
  <input type="date" name="{{.Name}}" value="{{.Value}}" class="input form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_timestamp"}}
  {{if .Readonly}}
    <input name="{{.Name}}" value="{{.Value}}" class="input form_input"{{if .Focused}} autofocus{{end}} readonly>
  {{else}}
    <div class="admin_timestamp">
      <input type="hidden" name="{{.Name}}" value="{{.Value}}">

      <input type="date" name="_admin_timestamp_hidden" class="input form_input admin_timestamp_date"{{if .Focused}} autofocus{{end}}>

      <select class="input form_input admin_timestamp_hour"></select>
      <span class="admin_timestamp_divider">:</span>
      <select class="input form_input admin_timestamp_minute"></select>

    </div>
  {{end}}
{{end}}

{{define "admin_item_image"}}
  <div class="admin_images">
    <input name="{{.Name}}" value="{{.Value}}" type="hidden">
    <div class="admin_images_list"></div>
    <a href="#" class="btn admin_images_edit">Edit</a>
    <progress></progress>
  </div>
{{end}}

{{define "admin_item_file"}}
  <input type="file" name="{{.Name}}" class="input form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
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

{{define "admin_item_place"}}
<div class="admin_place">
  <input type="hidden" name="{{.Name}}" value="{{.Value}}">
</div>
{{end}}

{{define "admin_item_hidden"}}
<input type="hidden" name="{{.Name}}" value="{{.Value}}">
{{end}}

{{define "admin_item_link"}}
<div>
  <a href="{{.Value}}" target="_blank">{{.Value}}</a>
</div>
{{end}}

{{define "admin_item_relation"}}
<div class="admin_item_relation">
  <input type="hidden" name="{{.Name}}" value="{{.Value}}" data-relation="{{.Values}}">
  <progress></progress>
</div>
{{end}}

{{define "admin_image"}}
<img src="{{thumb .Value}}">
{{end}}

{{define "admin_link"}}
  <a href="{{.Url}}">{{.Value}}</a>
{{end}}

{{define "admin_string"}}
{{.Value}}
{{end}}
{{define "admin_layout"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{if .admin_title}}{{.admin_title}} - {{.appName}}{{else}}Admin - {{.appName}}{{end}}</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <link rel="stylesheet" href="{{.admin_header.prefix}}/normalize.css?v={{.version}}">
    <link rel="stylesheet" href="{{.admin_header.prefix}}/admin.css?v={{.version}}">
    <script type="text/javascript" src="{{.admin_header.prefix}}/jquery.js?v={{.version}}"></script>
    <script type="text/javascript" src="{{.admin_header.prefix}}/image.js?v={{.version}}"></script>
    <script type="text/javascript" src="{{.admin_header.prefix}}/order.js?v={{.version}}"></script>
    <script type="text/javascript" src="{{.admin_header.prefix}}/place.js?v={{.version}}"></script>
    <script type="text/javascript" src="{{.admin_header.prefix}}/timestamp.js?v={{.version}}"></script>
    <script type="text/javascript" src="{{.admin_header.prefix}}/markdown.js?v={{.version}}"></script>
    <script type="text/javascript" src="{{.admin_header.prefix}}/relation.js?v={{.version}}"></script>
    <script src="https://maps.googleapis.com/maps/api/js?callback=bindPlaces" async defer></script>
    <script type="text/javascript" src="{{.admin_header.prefix}}/script.js?v={{.version}}"></script>

  </head>
  <body class="admin" data-admin-prefix="{{.admin_header.prefix}}">
    {{tmpl "admin_flash" .}}
    <div class="admin_header">
        <ul class="admin_header_list admin_header_list-right">
            <li><a href="{{.admin_header.prefix}}/user/settings">{{.currentuser.Email}}</a> <a href="{{.admin_header.prefix}}/logout?_csrfToken={{._csrfToken}}">{{message .locale "admin_log_out"}}</a></li>
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

    <div class="admin_footer">
      {{.appCode}} {{.appVersion}}
    </div>

    <div id="admin_images_popup">
      <div class="admin_images_popup_box" tabindex="1">
        <div class="admin_images_popup_box_header admin_popup_section">
          <h3>Selected images</h3>
          <div class="admin_images_popup_box_content"></div>
        </div>
        <div class="admin_images_popup_box_upload admin_popup_section">
            <h3>Upload new files</h3>
            <div><input type="file" accept=".jpg,.jpeg" multiple class="admin_popup_file"></div>
            <textarea placeholder="Popis" class="admin_popup_file_description input"></textarea>
            <br><br>
            <div class="admin_images_popup_box_upload_message"></div>
            <button class="admin_images_popup_box_upload_btn btn">Upload</button>
        </div>
        <div class="admin_images_popup_box_new admin_popup_section">
          <h3>Add Image</h3>
          <input class="admin_images_popup_filter"><button class="btn admin_images_popup_filter_button">Filter</button>
          <div class="admin_images_popup_box_new_list">
          </div>
        </div>
        <div class="admin_images_popup_box_footer">
          <button class="btn admin_images_popup_save">Save</button>
          <button class="btn admin_images_popup_cancel">Cancel</button>
        </div>
      </div>
    </div>
    
  </body>
</html>

{{end}}{{define "admin_list"}}

{{$adminResource := .admin_resource }}
{{$locale := .locale}}

{{$csrfToken := ._csrfToken}}
{{$table := .admin_list}}


<h2>{{.admin_title}}</h2>

<table class="admin_table{{if .admin_list.Order}} admin_table-order{{end}}">
  <tr class="admin_table_header">
  {{range $item := .admin_list.Header}}
    <th>
      {{if $item.CanOrder}}
        <a href="{{$item.OrderPath}}" class="{{if $item.Ordered}}ordered{{end}}{{if $item.OrderedDesc}} ordered-desc{{end}}">
      {{- end -}}
        {{- $item.NameHuman -}}
      {{if $item.CanOrder -}}
        </a>
      {{end}}
    </th>
  {{end}}
  <th{{if $table.HasDelete}} colspan="2"{{end}}>
    {{if $table.HasNew}}
      <a href="{{.admin_resource.ID}}/new" class="btn">{{message .locale "admin_new"}}</a>
    {{end}}
  </th>
  </tr>
{{range $item := .admin_list.Rows}}
  <tr data-id="{{$item.ID}}" class="admin_table_row">
    {{range $cell := $item.Items}}
    <td>
      {{ tmpl $cell.TemplateName $cell }}
    </td>
    {{end}}
    <td nowrap class="center top" style="width: 0px;">
      <a href="{{ $adminResource.ID}}/{{$item.ID}}" class="btn">{{message $locale "admin_edit"}}</a> 
    </td>
    {{if $table.HasDelete}}
      <td nowrap class="center top" style="width: 0px;">
        <form method="POST" action="{{ $adminResource.ID}}/{{$item.ID}}/delete?_csrfToken={{$csrfToken}}" onsubmit="return window.confirm('{{message $locale "admin_delete_confirmation"}}');">
          <input type="submit" value="{{message $locale "admin_delete"}}" class="btn">
        </form>
      </td>
    {{end}}
  </tr>
{{end}}
</table>

<div class="pagination">
{{range $page := .admin_list.Pagination.Pages}}
  {{if $page.Current}}
    <span class="pagination_page pagination_page-current">{{$page.Name}}</span>
  {{else}}
    <a href="{{$page.Url}}" class="pagination_page">{{$page.Name}}</a>
  {{end}}
{{end}}
</div>

{{end}}{{define "admin_login"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{.title}}</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{.admin_header_prefix}}/normalize.css?v={{.version}}">
    <link rel="stylesheet" href="{{.admin_header_prefix}}/admin.css?v={{.version}}">
  </head>
  <body class="admin">
    {{tmpl "admin_flash" .}}

    <div class="admin_content">
    <h2>{{.title}}</h2>

    {{tmpl "admin_form" .admin_form}}

    {{if .bottom}}
    <div style="text-align: center">
    {{Plain .bottom}}
    </div>
    {{end}}

    </div>
  </body>
</html>

{{end}}{{define "admin_message"}}
<h1>{{.message}}</h1>
{{end}}{{define "admin_new"}}

<h2>{{.admin_title}}</h2>

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

<h2>{{.admin_title}}</h2>

<a href="password">{{message .locale "admin_password_change"}}</a>

{{tmpl "admin_form" .admin_form}}

{{end}}{{define "admin_stats"}}
<h1>Stats</h1>

<table class="admin_table">
{{range $item := .stats}}
  <tr>
    <td>{{index $item 0}}</td>
    <td>{{index $item 1}}</td>
  </tr>
{{end}}
</table>

<h2>Configuration</h2>

<table class="admin_table">
{{range $item := .configStats}}
  <tr>
    <td>{{index $item 0}}</td>
    <td>{{index $item 1}}</td>
  </tr>
{{end}}
</table>

<h2>OS</h2>

<table class="admin_table">
{{range $item := .osStats}}
  <tr>
    <td>{{index $item 0}}</td>
    <td>{{index $item 1}}</td>
  </tr>
{{end}}
</table>

<h2>Memory</h2>

<table class="admin_table">
{{range $item := .memStats}}
  <tr>
    <td>{{index $item 0}}</td>
    <td>{{index $item 1}}</td>
  </tr>
{{end}}
</table>

<h2>Environment</h2>

<table class="admin_table">
{{range $item := .environmentStats}}
  <tr>
    <td>{{index $item 0}}</td>
    <td>{{index $item 1}}</td>
  </tr>
{{end}}
</table>


{{end}}`


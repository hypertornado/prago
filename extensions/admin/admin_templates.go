package admin
const adminTemplates = `
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
  <a href="{{.URL}}">{{.Value}}</a>
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

    <link rel="stylesheet" href="{{.admin_header.prefix}}/_static/admin.css?v={{.version}}">
    <script type="text/javascript" src="{{.admin_header.prefix}}/_static/admin.js?v={{.version}}"></script>

    <!--
    <script src="https://maps.googleapis.com/maps/api/js?callback=bindPlaces" async defer></script>
    -->

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
    <a href="{{$page.URL}}" class="pagination_page">{{$page.Name}}</a>
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
    <link rel="stylesheet" href="{{.admin_header_prefix}}/_static/admin.css?v={{.version}}">
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
    <link rel="stylesheet" href="{{.admin_header.prefix}}/_static/admin.css?v={{.version}}">
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


const adminCSS = `
/*! normalize.css v3.0.2 | MIT License | git.io/normalize */
img,
legend {
  border: 0;
}
legend,
td,
th {
  padding: 0;
}
html {
  font-family: sans-serif;
  -ms-text-size-adjust: 100%;
  -webkit-text-size-adjust: 100%;
}
body {
  margin: 0;
}
article,
aside,
details,
figcaption,
figure,
footer,
header,
hgroup,
main,
menu,
nav,
section,
summary {
  display: block;
}
audio,
canvas,
progress,
video {
  display: inline-block;
  vertical-align: baseline;
}
audio:not([controls]) {
  display: none;
  height: 0;
}
[hidden],
template {
  display: none;
}
a {
  background-color: transparent;
}
a:active,
a:hover {
  outline: 0;
}
abbr[title] {
  border-bottom: 1px dotted;
}
b,
optgroup,
strong {
  font-weight: 700;
}
dfn {
  font-style: italic;
}
h1 {
  font-size: 2em;
  margin: 0.67em 0;
}
mark {
  background: #ff0;
  color: #000;
}
small {
  font-size: 80%;
}
sub,
sup {
  font-size: 75%;
  line-height: 0;
  position: relative;
  vertical-align: baseline;
}
sup {
  top: -0.5em;
}
sub {
  bottom: -0.25em;
}
svg:not(:root) {
  overflow: hidden;
}
figure {
  margin: 1em 40px;
}
hr {
  -moz-box-sizing: content-box;
  box-sizing: content-box;
  height: 0;
}
pre,
textarea {
  overflow: auto;
}
code,
kbd,
pre,
samp {
  font-family: monospace,monospace;
  font-size: 1em;
}
button,
input,
optgroup,
select,
textarea {
  color: inherit;
  font: inherit;
  margin: 0;
}
button {
  overflow: visible;
}
button,
select {
  text-transform: none;
}
button,
html input[type=button],
input[type=reset],
input[type=submit] {
  -webkit-appearance: button;
  cursor: pointer;
}
button[disabled],
html input[disabled] {
  cursor: default;
}
button::-moz-focus-inner,
input::-moz-focus-inner {
  border: 0;
  padding: 0;
}
input {
  line-height: normal;
}
input[type=checkbox],
input[type=radio] {
  box-sizing: border-box;
  padding: 0;
}
input[type=number]::-webkit-inner-spin-button,
input[type=number]::-webkit-outer-spin-button {
  height: auto;
}
input[type=search] {
  -webkit-appearance: textfield;
  -moz-box-sizing: content-box;
  -webkit-box-sizing: content-box;
  box-sizing: content-box;
}
input[type=search]::-webkit-search-cancel-button,
input[type=search]::-webkit-search-decoration {
  -webkit-appearance: none;
}
fieldset {
  border: 1px solid silver;
  margin: 0 2px;
  padding: 0.35em 0.625em 0.75em;
}
table {
  border-collapse: collapse;
  border-spacing: 0;
}
html {
  box-sizing: border-box;
}
*,
*:before,
*:after {
  box-sizing: inherit;
}
html,
body {
  height: 100%;
  font-family: Roboto, -apple-system, BlinkMacSystemFont, "Helvetica Neue", "Segoe UI", Oxygen, Ubuntu, Cantarell, "Open Sans", sans-serif;
  font-size: 15px;
  line-height: 1.4em;
  color: #333;
}
h1 {
  font-size: 1.5rem;
  line-height: 1.4em;
  margin: 0.5em 0 0.2em 0;
}
h2 {
  font-size: 1.4rem;
  line-height: 1.4em;
  margin: 0.5em 0 0.2em 0;
}
h3 {
  font-size: 1.3rem;
  line-height: 1.4em;
  margin: 0.5em 0 0.2em 0;
}
h4 {
  font-size: 1.2rem;
  line-height: 1.4em;
  margin: 0.5em 0 0.2em 0;
}
.right {
  float: right;
}
.center {
  text-align: center;
}
.clear {
  clear: both;
}
.top {
  vertical-align: top;
}
p {
  margin: 0px;
  margin-bottom: 0.5em;
}
a {
  color: #4078c0;
}
a:hover {
  text-decoration: none;
}
.admin_header {
  background: #f5f5f5;
  padding: 10px 10px;
  border: 1px solid #e5e5e5;
}
.admin_header h1 {
  display: inline-block;
  font-size: 1.1em;
  margin: 0px 5px;
}
.admin_header_list {
  margin: 0px;
  padding: 0px;
}
.admin_header_list {
  display: inline-block;
}
.admin_header_list li {
  display: inline-block;
  padding: 0px 2px;
}
.admin_header_list-right {
  float: right;
}
.admin_content {
  max-width: 600px;
  padding: 10px 10px;
  margin: 0 auto;
}
.admin_footer {
  background: #f5f5f5;
  padding: 0px 10px;
  font-size: 0.7em;
  border: 1px solid #e5e5e5;
  color: #888;
  text-align: right;
}
.btn {
  display: inline-block;
  padding: 3px 12px;
  font-size: 0.9em;
  font-weight: bold;
  line-height: 20px;
  color: #333;
  white-space: nowrap;
  vertical-align: middle;
  cursor: pointer;
  background-color: #eee;
  background-image: -webkit-linear-gradient(#fcfcfc, #eee);
  background-image: linear-gradient(#fcfcfc, #eee);
  border: 1px solid #d5d5d5;
  border-radius: 3px;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
  user-select: none;
  -webkit-appearance: none;
  outline: none;
  text-decoration: none;
}
.btn:hover {
  background-image: linear-gradient(#eee, #ddd);
  border-color: #ccc;
}
.btn:active {
  background-color: #dcdcdc;
  background-image: none;
  border-color: #d5d5d5;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.15);
}
.form_errors_error {
  border: 1px solid #dd2e4f;
  color: #dd2e4f;
  padding: 5px;
  text-align: center;
  border-radius: 3px;
}
.form_label {
  display: block;
  margin: 20px 0px;
}
.form_label-required input {
  border-width: 2px;
}
.form_label-required .form_label_text {
  font-weight: bold;
}
.form_label-errors {
  color: #dd2e4f;
}
.form_label-errors input,
.form_label-errors textarea {
  border-color: #dd2e4f !important;
}
.form_label_errors {
  font-size: 0.8em;
}
.form_label_text-checkbox {
  padding: 0px 5px;
}
.input {
  display: inline-block;
  padding: 6px 8px;
  line-height: 20px;
  color: #333;
  white-space: nowrap;
  vertical-align: middle;
  border: 1px solid #ddd;
  border-radius: 3px;
  outline: none;
  font-size: 0.9rem;
  line-height: 1.4rem;
  width: 100%;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.075);
  background-color: white;
}
select.input {
  -webkit-appearance: menulist-button;
  appearance: menulist-button;
  height: 35px;
}
input[type=date].input {
  height: 35px;
}
.input[readonly],
.textarea[readonly],
.input[disabled],
.textarea[disabled] {
  border-color: #eee;
  background: #fafafa;
  color: #888;
  box-shadow: none;
}
.input:focus {
  border-color: #009ee0;
  background-color: white;
  outline: none;
  border-color: #51a7e8;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.075), 0 0 5px rgba(81, 167, 232, 0.5);
}
.textarea {
  min-height: 150px;
}
.admin_table {
  width: 100%;
}
.admin_table thead {
  font-weight: bold;
}
.admin_table td {
  padding: 2px;
  border: 1px solid #e5e5e5;
}
.admin_table th {
  padding: 5px;
  vertical-align: bottom;
  font-weight: normal;
}
.admin_header_item-active a {
  font-weight: bold;
}
.flash {
  text-align: center;
  padding: 5px;
  background: #4078c0;
}
.flash_message {
  color: white;
  font-weight: bold;
  display: inline-block;
}
.pagination {
  text-align: center;
}
.pagination_page-current {
  font-weight: bold;
}
/* images */
#admin_images_popup {
  display: none;
  position: fixed;
  top: 0px;
  left: 0px;
  padding: 15px;
  background: rgba(0, 0, 0, 0.4);
  width: 100%;
  height: 100%;
  text-align: center;
}
.admin_images_popup_box {
  text-align: left;
  max-width: 600px;
  width: 100%;
  padding: 5px;
  background: white;
  display: inline-block;
  border-radius: 3px;
  max-height: 100%;
  overflow: auto;
}
.admin_images_popup_box:focus {
  outline: none;
}
.admin_images_img {
  display: inline-block;
  max-height: 100px;
  margin: 3px;
}
.admin_images_edit {
  display: none;
}
.admin_popup_section {
  background-color: #eee;
  padding: 5px;
  border-radius: 3px;
  margin-bottom: 10px;
}
.admin_popup_section h3 {
  margin: 5px 0px;
}
.admin_popup_file {
  margin-bottom: 15px;
}
.admin_place_map {
  height: 300px;
}
/*markdown*/
.admin_markdown textarea {
  border-bottom-left-radius: 0px;
  border-bottom-right-radius: 0px;
  border-bottom: none;
  white-space: pre-wrap;
}
.admin_markdown_preview {
  background: white;
  border: 1px dashed #e5e5e5;
  padding: 5px;
  font-size: 0.7em;
  line-height: 1.4em;
  color: #888;
  border-bottom-left-radius: 3px;
  border-bottom-right-radius: 3px;
}
.admin_timestamp_date {
  width: 150px;
}
.admin_timestamp_hour {
  width: 60px;
}
.admin_timestamp_minute {
  width: 60px;
}
.ordered {
  font-weight: bold;
}
.ordered:after {
  content: "↓";
}
.ordered-desc:after {
  content: "↑";
}
`


const adminJS = `
function bindRelations() {
    function bindRelation(el) {
        var input = el.getElementsByTagName("input")[0];
        var relationName = input.getAttribute("data-relation");
        var originalValue = input.value;
        var select = document.createElement("select");
        select.classList.add("input");
        select.classList.add("form_input");
        select.addEventListener("change", function () {
            input.value = select.value;
        });
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/resource/" + relationName, true);
        request.onload = function () {
            if (this.status >= 200 && this.status < 400) {
                var resp = JSON.parse(this.response);
                console.log(resp);
                console.log("success");
            }
            else {
                console.log("server, error");
            }
        };
        request.onerror = function () {
            console.log("error");
        };
        request.send();
    }
    function addOption(select, value, description, selected) {
        var option = $("<option></option>");
        if (selected) {
            option.attr("selected", "selected");
        }
        option.attr("value", value);
        option.text(description);
        select.append(option);
    }
    var elements = document.querySelectorAll(".admin_item_relation");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindRelation(el);
    });
}
function bindPlaces() {
    function bindPlace(el) {
        var mapEl = document.createElement("div");
        mapEl.classList.add("admin_place_map");
        el.appendChild(mapEl);
        var position = { lat: 50.0796284, lng: 14.4292577 };
        var zoom = 1;
        var visible = false;
        var input = el.getElementsByTagName("input")[0];
        var inVal = input.value;
        var inVals = inVal.split(",");
        if (inVals.length == 2) {
            inVals[0] = parseFloat(inVals[0]);
            inVals[1] = parseFloat(inVals[1]);
            if (!isNaN(inVals[0]) && !isNaN(inVals[1])) {
                position.lat = inVals[0];
                position.lng = inVals[1];
                zoom = 11;
                visible = true;
            }
        }
        var map = new google.maps.Map(mapEl, {
            center: position,
            zoom: zoom
        });
        var marker = new google.maps.Marker({
            position: position,
            map: map,
            draggable: true,
            title: "",
            visible: visible
        });
        marker.addListener("position_changed", function () {
            var p = marker.getPosition();
            var str = stringifyPosition(p.lat(), p.lng());
            input.value = str;
        });
        marker.addListener("click", function () {
            marker.setVisible(false);
            input.value = "";
        });
        map.addListener('click', function (e) {
            position.lat = e.latLng.lat();
            position.lng = e.latLng.lng();
            marker.setPosition(position);
            marker.setVisible(true);
        });
        function stringifyPosition(lat, lng) {
            return lat + "," + lng;
        }
    }
    var elements = document.querySelectorAll(".admin_place");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindPlace(el);
    });
}
window.onload = function () {
    bindRelations();
};
`


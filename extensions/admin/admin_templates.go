package admin
const adminTemplates = `
{{define "admin_edit"}}

<div class="admin_box">

  <h2>{{.admin_title}}</h2>

  <a href="../../{{.admin_resource.ID}}">{{message .locale "admin_back"}}</a>

  {{tmpl "admin_form" .admin_form}}

</div>

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

<div class="admin_box">

<h2>{{.admin_title}}</h2>

{{tmpl "admin_form" .admin_form}}

</div>

{{end}}{{define "admin_home"}}

<div class="admin_box">

  <h2>{{.admin_header.Name}}</h2>

  <table class="admin_table">
  {{range $item := .admin_header.Items}}
    <tr>
      <td>
        <a href="{{$item.Url}}">{{$item.Name}}</a>
      </td>
    </tr>
  {{end}}
  </table>
</div>

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
    <input name="{{.Name}}" value="{{.Value}}" type="hidden" class="admin_images_hidden">
    <div class="admin_images_loaded hidden">
      <input type="file" accept=".jpg,.jpeg,.png" multiple class="admin_images_fileinput">
      <div class="admin_images_preview"></div>
    </div>
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
  <div class="admin_thumb">
    <img src="{{thumb .Value}}">
  </div>
{{end}}

{{define "admin_link"}}
  <a href="{{.URL}}">{{.Value}}</a>
{{end}}

{{define "admin_string"}}
{{.Value}}
{{end}}

{{define "admin_cell_checkbox"}}
<div class="center">
  {{if .Value}}
    ✅
  {{else}}
    -
  {{end}}
</div>
{{end}}
{{define "admin_layout"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{if .admin_title}}{{.admin_title}} — {{.appName}}{{else}}Admin — {{.appName}}{{end}}</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <link rel="stylesheet" href="{{.admin_header.UrlPrefix}}/_static/admin.css?v={{.version}}">
    {{range $c := .css}}
        <link rel="stylesheet" href="{{$c}}">
    {{end}}

    <script type="text/javascript" src="{{.admin_header.UrlPrefix}}/_static/admin.js?v={{.version}}"></script>
    {{range $javascript := .javascripts}}
        <script type="text/javascript" src="{{$javascript}}"></script>
    {{end}}
    <script src="https://maps.googleapis.com/maps/api/js?callback=bindPlaces&key={{.google}}" async defer></script>

  </head>
  <body class="admin" data-csrf-token="{{._csrfToken}}" data-admin-prefix="{{.admin_header.UrlPrefix}}"
    {{if .admin_header.Background}}
      style="background: linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(255,255,255, 0.9) 100%), url('{{.admin_header.Background}}'); background-size: cover; background-attachment: fixed;" 
    {{end}}
    >
    {{tmpl "admin_flash" .}}
    <div class="admin_header">
        <div class="admin_header_top">
            <a href="{{.admin_header.UrlPrefix}}" class="admin_header_name admin_header_top_item">
              {{if .admin_header.Logo}}
                <div class="admin_logo" style="background-image: url('{{.admin_header.Logo}}');"></div>
              {{end}}
            {{message .locale "admin_admin"}} — {{.admin_header.Name}}</a>
            <a href="/" class="admin_header_top_item">{{.admin_header.HomepageUrl}}</a>
            <div class="admin_header_top_item admin_header_top_space"></div>
            <div class="admin_header_top_item">{{.currentuser.Email}}</div>
            <a href="{{.admin_header.UrlPrefix}}/user/settings" class="admin_header_top_item">{{message .locale "admin_settings"}}</a>
            <a href="{{.admin_header.UrlPrefix}}/logout?_csrfToken={{._csrfToken}}" class="admin_header_top_item">{{message .locale "admin_log_out"}}</a>
        </div>


        {{ $admin_resource := .admin_resource }}

        <div class="admin_header_resources">
            {{range $item := .admin_header.Items}}
                <a href="{{$item.Url}}" class="admin_header_resource {{if $admin_resource}}{{ if eq $admin_resource.ID $item.ID }}admin_header_resource-active{{end}}{{end}}">{{$item.Name}}</a>
            {{end}}
        </div>
    </div>

    <div class="admin_content">

    {{if .template_before}}{{tmpl .template_before .}}{{end}}

    {{tmpl .admin_yield .}}

    {{if .template_after}}{{tmpl .template_after .}}{{end}}

    </div>
  </body>
</html>

{{end}}{{define "admin_layout_nologin"}}
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
  <body class="admin_nologin"
    {{if .background}}
      style="background: linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(255,255,255, 0.9) 100%), url('{{.background}}'); background-size: cover; background-attachment: fixed;" 
    {{end}}
  >
    {{tmpl "admin_flash" .}}

    {{tmpl .yield .}}

  </body>
</html>

{{end}}{{define "admin_list"}}

{{$adminResource := .admin_resource }}
{{$locale := .locale}}

{{$csrfToken := ._csrfToken}}
{{$table := .admin_list}}

{{$global := .}}

{{range $snippet := .admin_resource.Snippets}}
  {{tmpl $snippet.Template $global}}
{{end}}

{{$list := .admin_list}}

<table class="admin_table admin_table-list {{if .admin_list.CanChangeOrder}} admin_table-order{{end}}" data-type="{{.admin_list.TypeID}}" data-order-column="{{.admin_list.OrderColumn}}" data-order-desc="{{.admin_list.OrderDesc}}">
  <thead>
  <tr>
    <td colspan="{{.admin_list.Colspan}}">
      <div class="admin_table_listheader">
        <h2>{{.admin_title}}</h2>

        <div class="btngrp">
        {{range $item := .admin_list.Actions}}
          <a href="{{$item.Url}}" class="btn">{{$item.Name}}</a>
        {{end}}
        </div>
      </div>
    </td>
  </tr>
  <tr>
  {{range $item := .admin_list.Header}}
    <th>
      {{if $item.CanOrder}}
        <a href="#" class="admin_table_orderheader" data-name="{{$item.ColumnName}}">
      {{- end -}}
        {{- $item.NameHuman -}}
      {{if $item.CanOrder -}}
        </a>
      {{end}}
    </th>
  {{end}}
  <th></th>
  </tr>
  <tr>
    {{range $item := .admin_list.Header}}
      <th>
        {{if $item.CanFilter}}
          <input class="input admin_table_filter_item" data-typ="{{$item.ColumnName}}">
        {{end}}
      </th>
    {{end}}
    <th></th>
  </tr>
  </thead>
  <tbody></tbody>
</table>

{{end}}

{{define "admin_list_cells"}}
{{range $item := .admin_list.Rows}}
  <tr data-id="{{$item.ID}}" class="admin_table_row">
    {{range $cell := $item.Items}}
    <td>
      {{ tmpl $cell.TemplateName $cell }}
    </td>
    {{end}}
    <td nowrap class="center top">
      <div class="btngrp">
        {{range $action := $item.Actions}}
          {{if $action.Url}}
            <a href="{{$action.Url}}" class="btn">{{$action.Name}}</a>
          {{else}}
            <div{{range $k, $v := $action.Params}} {{HTMLAttr $k}}="{{$v}}"{{end}}>{{$action.Name}}</div>
          {{end}}
        {{end}}
      </div>
    </td>
  </tr>
{{end}}
{{if .admin_list.Pagination.Pages}}
  <tr>
    <td colspan="{{.admin_list.Colspan}}" class="pagination">
      {{range $page := .admin_list.Pagination.Pages}}
        {{if $page.Current}}
          <span class="pagination_page_current">{{$page.Page}}</span>
        {{else}}
          <a href="#" class="pagination_page" data-page="{{$page.Page}}">{{$page.Page}}</a>
        {{end}}
      {{end}}
    </td>
  </tr>
{{end}}
{{end}}
{{define "admin_login"}}
<div class="admin_box">
    <h2>{{.title}}</h2>

    {{tmpl "admin_form" .admin_form}}

    {{if .bottom}}
    <div style="text-align: center">
    {{HTML .bottom}}
    </div>
    {{end}}
</div>
{{end}}{{define "admin_message"}}

<div class="admin_box">
  <h1>{{.message}}</h1>
</div>
{{end}}{{define "admin_new"}}

<div class="admin_box">
  <h2>{{.admin_title}}</h2>

  <a href="../{{.admin_resource.ID}}">{{message .locale "admin_back"}}</a>

  {{tmpl "admin_form" .admin_form}}
</div>

{{end}}{{define "admin_settings"}}

<div class="admin_box">

  <h2>{{.admin_title}}</h2>

  <a href="password">{{message .locale "admin_password_change"}}</a>

  {{tmpl "admin_form" .admin_form}}
</div>

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


{{end}}{{define "elastic_index"}}
<h1>Elastic Index</h1>

<h2>Global Stats</h2>
  <ul>
    {{range $url := .urls}}
    <li><a href="{{index $url 0}}">{{index $url 1}}</a></li>
    {{end}}
  </ul>

<h2>Indexes</h2>
  <ul>
    {{range $index := .indexes}}
    <li><a href="elastic/index/{{$index}}">{{$index}}</a></li>
    {{end}}
  </ul>

{{end}}

{{define "elastic_pre"}}
<h1>{{.admin_title}}</h1>
  <pre>
    {{.data}}
  </pre>
{{end}}

{{define "elastic_detail"}}
  <h1>Index {{.index_name}}</h1>
  
  <h2>Field stats</h2>

  {{.fieldStats}}

{{end}}{{define "newsletter_empty"}}{{end}}{{define "newsletter_layout"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{.title}}</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <style type="text/css">
      body {
        font-family: Roboto, -apple-system, BlinkMacSystemFont, "Helvetica Neue", "Segoe UI", Oxygen, Ubuntu, Cantarell, "Open Sans", sans-serif;
      }

      label {
        display: block;
        margin: 10px 0px;
      }
      .box {
        margin: 0 auto;
        padding: 5px;
        border: 0px solid red;
        max-width: 500px;
      }

      input {
        max-width: 200px;
        display: block;
      }
    </style>

  </head>
  <body>
    <div class="box">
      <a href="/">{{.site}}</a>
      <h1>{{.title}}</h1>
    {{tmpl .yield .}}
    </div>
  </body>
</html>

{{end}}{{define "newsletter_send"}}

<div class="admin_box">
<form method="POST" action="send">
<h1>Odeslat newsletter - {{.title}}</h1>

<b>Emailové adresy ({{.recipients_count}})</b>
{{range $item := .recipients}}
  <div>{{$item}}</div>
{{end}}


<input type="submit" class="btn" value="Odeslat newsletter">
</form>
</div>

{{end}}{{define "newsletter_send_preview"}}
<div class="admin_box">
<form method="POST" action="send-preview">
<h1>Odeslat náhled newsletteru</h1>
<label>
  Seznam emailů na poslání preview (jeden email na řádek)
  <textarea class="input" name="emails"></textarea>
</label>

<input type="submit" class="btn">
</form>
</div>


{{end}}{{define "newsletter_sent"}}

<div class="admin_box">
  <h1>Newsletter odeslán</h1>

  <b>Emailové adresy ({{.recipients_count}})</b>
  {{range $item := .recipients}}
    <div>{{$item}}</div>
  {{end}}
  
</div>

{{end}}{{define "newsletter_snippet"}}

<div class="admin_box">
Počet odběratelů newsletteru: {{.recipients_count}}
</div>

{{end}}{{define "newsletter_subscribe"}}
<form method="post" action="/newsletter-subscribe">
<label>
  Vaše jméno
  <input type="text" name="name">
</label>
<label>
  Email
  <input type="email" name="email">
</label>
<input type="submit" value="Přihlásit se k odběru newsletteru">
<input type="hidden" name="csrf" value="{{.csrf}}">
</form>
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
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
*,
*:before,
*:after {
  box-sizing: inherit;
}
html,
body {
  font-family: Roboto, -apple-system, BlinkMacSystemFont, "Helvetica Neue", "Segoe UI", Oxygen, Ubuntu, Cantarell, "Open Sans", sans-serif;
  font-size: 13px;
  line-height: 1.4em;
  color: #444;
}
body {
  background-color: #f3f3f3;
  background-size: cover;
  background-attachment: fixed;
}
.admin_nologin > .admin_box {
  margin-top: 20px;
}
.shadow {
  box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.2);
}
h1 {
  font-size: 1.5rem;
  line-height: 1.4em;
  margin: 0.5em 0 0.2em 0;
}
h2 {
  font-size: 1.4rem;
  line-height: 1.4em;
  margin: 0px;
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
.hidden {
  display: none;
}
.top {
  vertical-align: top;
}
p {
  margin: 0px;
  margin-bottom: 0.5em;
}
progress {
  display: block;
  margin: 5px auto;
}
a {
  color: #4078c0;
}
a:hover {
  text-decoration: none;
}
.admin_content {
  padding: 5px;
}
.admin_box {
  box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.2);
  margin: 5px auto;
  background-color: white;
  padding: 10px;
  border-radius: 3px;
  max-width: 600px;
}
.admin_footer {
  padding: 0px 10px;
  font-size: 0.7em;
  border-top: 1px solid #e5e5e5;
  color: #888;
  text-align: right;
  box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.2);
  position: relative;
  z-index: 2;
  background-color: white;
  display: none;
}
.btn {
  display: inline-block;
  padding: 3px 12px;
  font-size: 0.9em;
  font-weight: bold;
  line-height: 1.2em;
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
.btngrp {
  display: flex;
  text-align: right;
  justify-content: flex-end;
}
.btngrp > :not(:last-child) {
  border-right: none;
  border-top-right-radius: 0px;
  border-bottom-right-radius: 0px;
}
.btngrp > :not(:first-child) {
  border-top-left-radius: 0px;
  border-bottom-left-radius: 0px;
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
  margin: 10px 0px;
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
  padding: 2px 2px;
  line-height: 1.2em;
  color: #333;
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
.input:focus,
.btn:focus {
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
.admin_table-list {
  background-color: white;
  margin: 10px 0px;
  box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.2);
}
.admin_table_listheader {
  display: flex;
  justify-content: space-between;
  margin: 5px;
  align-items: center;
}
.admin_table thead {
  font-weight: bold;
}
.admin_table td {
  padding: 2px;
  border: 1px solid #f1f1f1;
}
.admin_table-list td {
  border-left: none;
  border-bottom: none;
}
.admin_table-list tr td:last-child {
  border-right: none;
}
.admin_table th {
  padding: 5px;
  vertical-align: bottom;
  font-weight: normal;
  border: 1px solid #f1f1f1;
  border-left: none;
  background-color: #fafafa;
}
.admin_table th a {
  display: block;
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
.pagination_page_current {
  font-weight: bold;
}
/* images */
.admin_images_preview {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}
.admin_images_image {
  padding: 3px;
  border: 1px solid #eee;
  border-radius: 3px;
  margin: 2px;
  display: inline-block;
  text-align: center;
  vertical-align: middle;
  position: relative;
}
.admin_images_image_delete {
  position: absolute;
  top: 1px;
  right: 0px;
  background-color: white;
  width: 20px;
  height: 20px;
  font-size: 17px;
  border-bottom-left-radius: 3px;
}
.admin_images_image_delete:hover {
  font-weight: bold;
}
.admin_images_image img {
  max-height: 150px;
  max-width: 150px;
  margin: 0 auto;
  vertical-align: middle;
  display: inline-block;
}
.admin_images_fileinput {
  display: block;
  margin: 0px auto;
  border: 1px solid #eee;
  padding: 3px;
  border-radius: 3px;
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
  overflow: auto;
}
.admin_thumb {
  text-align: center;
}
.admin_thumb img {
  max-height: 30px;
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
.admin-action-order {
  background: #fafafa !important;
  cursor: move;
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
.admin_header {
  background: white;
  font-size: 14px;
  padding-bottom: 0px;
  box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.2);
  position: relative;
  z-index: 2;
  line-height: 1.6em;
}
.admin_logo {
  width: 40px;
  height: 20px;
  border: 0px solid red;
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
  margin-right: 10px;
}
.admin_header_top {
  display: flex;
  padding: 0px 5px;
  flex-wrap: wrap;
  align-items: center;
}
.admin_header_top_item {
  padding: 5px;
}
.admin_header a {
  text-decoration: none;
}
.admin_header a:hover {
  background-color: #eee;
}
.admin_header_top_space {
  flex-grow: 2;
}
.admin_header_name {
  font-weight: bold;
  display: flex;
  align-items: center;
}
.admin_header_resources {
  border-top: 1px solid #e5e5e5;
  margin: 0px;
  padding: 0px;
  clear: both;
  display: flex;
  margin-left: -2px;
  font-size: 0px;
  padding-left: 5px;
  flex-wrap: wrap;
}
.admin_header_resource {
  text-transform: uppercase;
  padding: 5px 5px;
  margin: 0px;
  margin-right: 2px;
  font-size: 12px;
  border-bottom: 2px solid none;
}
.admin_header_resource:hover {
  background-color: #eee;
}
.admin_header_resource-active {
  font-weight: bold;
  border-bottom: 2px solid #4078c0;
}
`


const adminJS = `
function DOMinsertChildAtIndex(parent, child, index) {
    if (index >= parent.children.length) {
        parent.appendChild(child);
    }
    else {
        parent.insertBefore(child, parent.children[index]);
    }
}
function bindImagePickers() {
    var els = document.querySelectorAll(".admin_images");
    for (var i = 0; i < els.length; i++) {
        new ImagePicker(els[i]);
    }
}
var ImagePicker = (function () {
    function ImagePicker(el) {
        var _this = this;
        this.el = el;
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.hiddenInput = el.querySelector(".admin_images_hidden");
        this.preview = el.querySelector(".admin_images_preview");
        this.fileInput = this.el.querySelector(".admin_images_fileinput");
        this.progress = this.el.querySelector("progress");
        this.el.querySelector(".admin_images_loaded").classList.remove("hidden");
        this.hideProgress();
        var ids = this.hiddenInput.value.split(",");
        for (var i = 0; i < ids.length; i++) {
            var id = ids[i];
            if (id) {
                this.addImage(id);
            }
        }
        this.fileInput.addEventListener("change", function () {
            var files = _this.fileInput.files;
            var formData = new FormData();
            if (files.length == 0) {
                return;
            }
            for (var i = 0; i < files.length; i++) {
                formData.append("file", files[i]);
            }
            var request = new XMLHttpRequest();
            request.open("POST", _this.adminPrefix + "/_api/image/upload");
            request.addEventListener("load", function (e) {
                _this.hideProgress();
                if (request.status == 200) {
                    var data = JSON.parse(request.response);
                    for (var i = 0; i < data.length; i++) {
                        _this.addImage(data[i].UID);
                    }
                }
                else {
                    alert("Error while uploading image.");
                    console.error("Error while loading item.");
                }
            });
            _this.fileInput.type = "";
            _this.fileInput.type = "file";
            _this.showProgress();
            request.send(formData);
        });
    }
    ImagePicker.prototype.updateHiddenData = function () {
        var ids = [];
        for (var i = 0; i < this.preview.children.length; i++) {
            var item = this.preview.children[i];
            var uuid = item.getAttribute("data-uuid");
            ids.push(uuid);
        }
        this.hiddenInput.value = ids.join(",");
    };
    ImagePicker.prototype.addImage = function (id) {
        var _this = this;
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("data-uuid", id);
        container.setAttribute("draggable", "true");
        container.setAttribute("target", "_blank");
        container.setAttribute("href", this.adminPrefix + "/file/uuid/" + id);
        container.addEventListener("dragstart", function (e) {
            _this.draggedElement = e.target;
        });
        container.addEventListener("drop", function (e) {
            var droppedElement = e.toElement;
            for (var i = 0; i < 3; i++) {
                if (droppedElement.nodeName == "A") {
                    break;
                }
                else {
                    droppedElement = droppedElement.parentElement;
                }
            }
            var draggedIndex = -1;
            var droppedIndex = -1;
            var parent = _this.draggedElement.parentElement;
            for (var i = 0; i < parent.children.length; i++) {
                var child = parent.children[i];
                if (child == _this.draggedElement) {
                    draggedIndex = i;
                }
                if (child == droppedElement) {
                    droppedIndex = i;
                }
            }
            if (draggedIndex == -1 || droppedIndex == -1) {
                return;
            }
            if (draggedIndex <= droppedIndex) {
                droppedIndex += 1;
            }
            DOMinsertChildAtIndex(parent, _this.draggedElement, droppedIndex);
            _this.updateHiddenData();
        });
        container.addEventListener("dragover", function (e) {
            e.preventDefault();
        });
        container.addEventListener("click", function (e) {
            var target = e.target;
            if (target.classList.contains("admin_images_image_delete")) {
                var parent = e.currentTarget.parentNode;
                parent.removeChild(e.currentTarget);
                _this.updateHiddenData();
                e.preventDefault();
                return false;
            }
        });
        var img = document.createElement("img");
        img.setAttribute("src", this.adminPrefix + "/_api/image/thumb/" + id);
        img.setAttribute("draggable", "false");
        container.appendChild(img);
        var del = document.createElement("div");
        del.textContent = "×";
        del.classList.add("admin_images_image_delete");
        container.appendChild(del);
        this.preview.appendChild(container);
        this.updateHiddenData();
    };
    ImagePicker.prototype.hideProgress = function () {
        this.progress.classList.add("hidden");
    };
    ImagePicker.prototype.showProgress = function () {
        this.progress.classList.remove("hidden");
    };
    return ImagePicker;
}());
function bindLists() {
    var els = document.getElementsByClassName("admin_table-list");
    for (var i = 0; i < els.length; i++) {
        new List(els[i]);
    }
}
var List = (function () {
    function List(el) {
        this.el = el;
        this.page = 1;
        this.typeName = el.getAttribute("data-type");
        if (!this.typeName) {
            return;
        }
        this.tbody = el.querySelector("tbody");
        this.tbody.textContent = "";
        this.bindFilter();
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.orderColumn = el.getAttribute("data-order-column");
        if (el.getAttribute("data-order-desc") == "true") {
            this.orderDesc = true;
        }
        else {
            this.orderDesc = false;
        }
        this.bindOrder();
        this.load();
    }
    List.prototype.load = function () {
        var _this = this;
        var request = new XMLHttpRequest();
        request.open("POST", this.adminPrefix + "/_api/list/" + this.typeName + document.location.search, true);
        request.addEventListener("load", function () {
            _this.tbody.innerHTML = "";
            if (request.status == 200) {
                _this.tbody.innerHTML = request.response;
                bindOrder();
                bindDelete();
                _this.bindPage();
            }
            else {
                console.error("error while loading list");
            }
        });
        var requestData = this.getListRequest();
        request.send(JSON.stringify(requestData));
    };
    List.prototype.bindPage = function () {
        var _this = this;
        var pages = this.el.querySelectorAll(".pagination_page");
        for (var i = 0; i < pages.length; i++) {
            var pageEl = pages[i];
            pageEl.addEventListener("click", function (e) {
                var el = e.target;
                var page = parseInt(el.getAttribute("data-page"));
                _this.page = page;
                _this.load();
                e.preventDefault();
                return false;
            });
        }
    };
    List.prototype.bindOrder = function () {
        var _this = this;
        this.renderOrder();
        var headers = this.el.querySelectorAll(".admin_table_orderheader");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.addEventListener("click", function (e) {
                var el = e.target;
                var name = el.getAttribute("data-name");
                if (name == _this.orderColumn) {
                    if (_this.orderDesc) {
                        _this.orderDesc = false;
                    }
                    else {
                        _this.orderDesc = true;
                    }
                }
                else {
                    _this.orderColumn = name;
                    _this.orderDesc = false;
                }
                _this.renderOrder();
                _this.load();
                e.preventDefault();
                return false;
            });
        }
    };
    List.prototype.renderOrder = function () {
        var headers = this.el.querySelectorAll(".admin_table_orderheader");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.classList.remove("ordered");
            header.classList.remove("ordered-desc");
            var name = header.getAttribute("data-name");
            if (name == this.orderColumn) {
                header.classList.add("ordered");
                if (this.orderDesc) {
                    header.classList.add("ordered-desc");
                }
            }
        }
    };
    List.prototype.getListRequest = function () {
        var ret = {};
        ret.Page = this.page;
        ret.OrderBy = this.orderColumn;
        ret.OrderDesc = this.orderDesc;
        ret.Filter = this.getFilterData();
        return ret;
    };
    List.prototype.getFilterData = function () {
        var ret = {};
        var items = this.el.querySelectorAll(".admin_table_filter_item");
        for (var i = 0; i < items.length; i++) {
            var item = items[i];
            var typ = item.getAttribute("data-typ");
            var val = item.value.trim();
            if (val) {
                ret[typ] = val;
            }
        }
        return ret;
    };
    List.prototype.bindFilter = function () {
        this.filterInputs = this.el.querySelectorAll(".admin_table_filter_item");
        for (var i = 0; i < this.filterInputs.length; i++) {
            var input = this.filterInputs[i];
            input.addEventListener("change", this.inputListener.bind(this));
            input.addEventListener("keyup", this.inputListener.bind(this));
        }
        this.inputPeriodicListener();
    };
    List.prototype.inputListener = function () {
        this.page = 1;
        this.changed = true;
        this.changedTimestamp = Date.now();
    };
    List.prototype.inputPeriodicListener = function () {
        var _this = this;
        setInterval(function () {
            if (_this.changed == true && Date.now() - _this.changedTimestamp > 500) {
                _this.changed = false;
                _this.load();
            }
        }, 200);
    };
    return List;
}());
var getParams = function (query) {
    if (!query) {
        return {};
    }
    return (/^[?#]/.test(query) ? query.slice(1) : query)
        .split('&')
        .reduce(function (params, param) {
        var _a = param.split('='), key = _a[0], value = _a[1];
        params[key] = value ? decodeURIComponent(value.replace(/\+/g, ' ')) : '';
        return params;
    }, {});
};
function bindOrder() {
    function orderTable(el) {
        var rows = el.getElementsByClassName("admin_table_row");
        Array.prototype.forEach.call(rows, function (item, i) {
            bindDraggable(item);
        });
        var draggedElement;
        function bindDraggable(row) {
            row.setAttribute("draggable", "true");
            row.addEventListener("dragstart", function (ev) {
                draggedElement = this;
                ev.dataTransfer.setData('text/plain', '');
            });
            row.addEventListener("drop", function (ev) {
                var targetEl = this;
                if (this != draggedElement) {
                    var draggedIndex = -1;
                    var thisIndex = -1;
                    Array.prototype.forEach.call(el.getElementsByClassName("admin_table_row"), function (item, i) {
                        if (item == draggedElement) {
                            draggedIndex = i;
                        }
                        if (item == targetEl) {
                            thisIndex = i;
                        }
                    });
                    if (draggedIndex < thisIndex) {
                        thisIndex += 1;
                    }
                    DOMinsertChildAtIndex(targetEl.parentElement, draggedElement, thisIndex);
                    saveOrder();
                }
                return false;
            });
            row.addEventListener("dragover", function (ev) {
                ev.preventDefault();
            });
        }
        function saveOrder() {
            var ajaxPath = document.location.pathname + "/order";
            var order = [];
            var rows = el.getElementsByClassName("admin_table_row");
            Array.prototype.forEach.call(rows, function (item, i) {
                order.push(parseInt(item.getAttribute("data-id")));
            });
            var request = new XMLHttpRequest();
            request.open("POST", ajaxPath, true);
            request.addEventListener("load", function () {
                if (request.status != 200) {
                    console.error("Error while saving order.");
                }
            });
            request.send(JSON.stringify({ "order": order }));
        }
    }
    var elements = document.querySelectorAll(".admin_table-order");
    Array.prototype.forEach.call(elements, function (el, i) {
        orderTable(el);
    });
}
function bindMarkdowns() {
    function bindMarkdown(el) {
        var textarea = el.getElementsByTagName("textarea")[0];
        var lastChanged = Date.now();
        var changed = false;
        setInterval(function () {
            if (changed && (Date.now() - lastChanged > 500)) {
                loadPreview();
            }
        }, 100);
        textarea.addEventListener("change", textareaChanged);
        textarea.addEventListener("keyup", textareaChanged);
        function textareaChanged() {
            changed = true;
            lastChanged = Date.now();
        }
        loadPreview();
        function loadPreview() {
            changed = false;
            var request = new XMLHttpRequest();
            request.open("POST", document.body.getAttribute("data-admin-prefix") + "/_api/markdown", true);
            request.addEventListener("load", function () {
                if (request.status == 200) {
                    var previewEl = el.getElementsByClassName("admin_markdown_preview")[0];
                    previewEl.innerHTML = JSON.parse(request.response);
                }
                else {
                    console.error("Error while loading markdown preview.");
                }
            });
            request.send(textarea.value);
        }
    }
    var elements = document.querySelectorAll(".admin_markdown");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindMarkdown(el);
    });
}
function bindTimestamps() {
    function bindTimestamp(el) {
        var hidden = el.getElementsByTagName("input")[0];
        var v = hidden.value;
        if (v == "0001-01-01 00:00") {
            var d = new Date();
            var month = d.getMonth() + 1;
            var monthStr = String(month);
            if (month < 10) {
                monthStr = "0" + monthStr;
            }
            var day = d.getUTCDate();
            var dayStr = String(day);
            if (day < 10) {
                dayStr = "0" + dayStr;
            }
            v = d.getFullYear() + "-" + monthStr + "-" + dayStr + " " + d.getHours() + ":" + d.getMinutes();
        }
        var date = v.split(" ")[0];
        var hour = parseInt(v.split(" ")[1].split(":")[0]);
        var minute = parseInt(v.split(" ")[1].split(":")[1]);
        var timestampEl = el.getElementsByClassName("admin_timestamp_date")[0];
        timestampEl.value = date;
        var hourEl = el.getElementsByClassName("admin_timestamp_hour")[0];
        for (var i = 0; i < 24; i++) {
            var newEl = document.createElement("option");
            var addVal = "" + i;
            if (i < 10) {
                addVal = "0" + addVal;
            }
            newEl.innerText = addVal;
            newEl.setAttribute("value", addVal);
            if (hour == i) {
                newEl.setAttribute("selected", "selected");
            }
            hourEl.appendChild(newEl);
        }
        var minEl = el.getElementsByClassName("admin_timestamp_minute")[0];
        for (var i = 0; i < 60; i++) {
            var newEl = document.createElement("option");
            var addVal = "" + i;
            if (i < 10) {
                addVal = "0" + addVal;
            }
            newEl.innerText = addVal;
            newEl.setAttribute("value", addVal);
            if (minute == i) {
                newEl.setAttribute("selected", "selected");
            }
            minEl.appendChild(newEl);
        }
        var elTsDate = el.getElementsByClassName("admin_timestamp_date")[0];
        var elTsHour = el.getElementsByClassName("admin_timestamp_hour")[0];
        var elTsMinute = el.getElementsByClassName("admin_timestamp_minute")[0];
        var elTsInput = el.getElementsByTagName("input")[0];
        function saveValue() {
            var str = elTsDate.value + " " + elTsHour.value + ":" + elTsMinute.value;
            elTsInput.value = str;
        }
        saveValue();
        elTsDate.addEventListener("change", saveValue);
        elTsHour.addEventListener("change", saveValue);
        elTsMinute.addEventListener("change", saveValue);
    }
    var elements = document.querySelectorAll(".admin_timestamp");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindTimestamp(el);
    });
}
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
        var progress = el.getElementsByTagName("progress")[0];
        request.addEventListener("load", function () {
            if (request.status >= 200 && request.status < 400) {
                var resp = JSON.parse(request.response);
                addOption(select, "0", "", false);
                Array.prototype.forEach.call(resp, function (item, i) {
                    var selected = false;
                    if (originalValue == item["id"]) {
                        selected = true;
                    }
                    addOption(select, item["id"], item["name"], selected);
                });
                el.appendChild(select);
            }
            else {
                console.error("Error wile loading relation " + relationName + ".");
            }
            progress.style.display = 'none';
        });
        request.onerror = function () {
            console.error("Error wile loading relation " + relationName + ".");
            progress.style.display = 'none';
        };
        request.send();
    }
    function addOption(select, value, description, selected) {
        var option = document.createElement("option");
        if (selected) {
            option.setAttribute("selected", "selected");
        }
        option.setAttribute("value", value);
        option.innerText = description;
        select.appendChild(option);
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
function bindDelete() {
    var deleteButtons = document.querySelectorAll(".admin-action-delete");
    for (var i = 0; i < deleteButtons.length; i++) {
        bindDeleteButton(deleteButtons[i]);
    }
}
function bindDeleteButton(btn) {
    var _this = this;
    var csrfToken = document.body.getAttribute("data-csrf-token");
    btn.addEventListener("click", function () {
        var message = btn.getAttribute("data-confirm-message");
        var url = btn.getAttribute("data-action") + csrfToken;
        if (confirm(message)) {
            var request = new XMLHttpRequest();
            request.open("POST", url, true);
            request.addEventListener("load", function () {
                if (_this.status == 200) {
                    document.location.reload();
                }
                else {
                    console.error("Error while deleting item");
                }
            });
            request.send();
        }
    });
}
document.addEventListener("DOMContentLoaded", function () {
    bindMarkdowns();
    bindTimestamps();
    bindRelations();
    bindImagePickers();
    bindClickAndStay();
    bindLists();
});
function bindClickAndStay() {
    var els = document.getElementsByName("_submit_and_stay");
    var elsClicked = document.getElementsByName("_submit_and_stay_clicked");
    if (els.length == 1 && elsClicked.length == 1) {
        els[0].addEventListener("click", function () {
            elsClicked[0].value = "true";
        });
    }
}
`


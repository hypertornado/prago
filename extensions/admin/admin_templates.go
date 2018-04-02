package admin
const adminTemplates = `
{{define "admin_export"}}
  
  <form method="POST" action="export">

    <h2>Fields</h2>
    {{range $field := .Fields}}
      <label class="form_label">
        <input type="checkbox" name="_field" value="{{$field.ColumnName}}" checked>
        <span class="form_label_text-inline">{{$field.NameHuman}}</span>
      </label>
    {{end}}

    <h2>Format</h2>
    <select name="_format" class="input">
    {{range $format := .Formats}}
      <option value="{{$format}}">{{$format}}</option>
    {{end}}
    </select>


    <h2>Limit</h2>
    <input name="_limit" type="number" class="input">

    <h2>Fields</h2>
    {{range $field := .Fields}}
      <label class="form_label">
        <span class="form_label_text">{{$field.NameHuman}}</span>
        <input name="{{$field.ColumnName}}" class="input">
      </label>
    {{end}}

    <input type="submit" class="btn">

  </form>
{{end}}
{{define "admin_flash"}}
  {{if .flash_messages}}
    <div class="flash_messages">
      {{range $message := .flash_messages}}
        <div class="flash_message">
          <div class="flash_message_content">{{$message}}</div>
          <div class="flash_message_close">✕</div>
        </div>
      {{end}}
    </div>
  {{end}}
{{end}}
{{define "admin_form"}}

<form method="{{.Method}}" action="{{.Action}}" class="form{{range $class := .Classes}} {{$class}}{{end}}" enctype="multipart/form-data" novalidate>

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

{{end}}{{define "admin_help_markdown"}}

<div class="admin_box">
  <h1>Nápověda k Markdown</h1>

  <p>
    Markdown je jazyk na zápis formátovaného textu na webové stránky.
  </p>

  <h2>Odstavce</h2>
  <p>Odstavce textu vytvoříte stejně jako v programu MS Word - stačí odřádkovat.</p>

  <h2>Tučný text a kurzíva</h2>
  <p>Text můžete označit <b>tučně</b> pomocí dvou hvězdiček (<code>Toto je **tučný** text</code>). U kurzívy stačí jedna hvězdička</p>

  <h2>Odkazy</h2>
  <p>Odkazy jdou vložit dvěma způsoby. Pokud vložíte text s webovou adresou, sám se změní na odkaz (např <code>http://www.seznam.cz</code> se změní na <a href="http://www.seznam.cz">http://www.seznam.cz</a>).</p>

  <p>Pokud chcete použit odkaz na konkrétních slovech, dáte tyto slova do hranatých závorek a adresu odkazu za to do závorek kulatých (např. <code>[Seznam](http://www.seznam.cz)</code> se vykreslí jako <a href="http://www.seznam.cz">Seznam</a>).</p>

  <h2>Nadpisy</h2>
  <p>Nejčastěji používaný odkaz druhé úrovně vytvoříte tak, že jeho text dáte na samotný řádek a před něj dáte dva znaky "#" (<code>## Toto je nadpis</code>).</p>

  <h2>Seznam</h2>
  <p>Seznam s odrážkami vytvoříte tak, že každou položku seznamu dáte na zvláštní řádek a přidáte na začátek řádku "* " (např. <code>* první odrážka</code>)</p>

  <h2>Obrázky</h2>
  <p>Obrázky se vloží takto: <code>![popis obrázku](http://odkaz/na/obrazek)</code>.</p>

  <h2>Ostatní</h2>
  <p>Může být potřeba vložit nějaké jiné formátování, nebo prvek. Například chcete vložit YouTube video. Ve většině případů jde do Markdown textu vložit i HTML prvky. Pokud HTML příliš neovládáte, je ve většině případů lepší se tomuto způsobu vyhnout.</p>

</div>

{{end}}{{define "admin_history"}}

  <table class="admin_table">
    <tr>
      <th>#</th>
      <th>Typ Akce</th>
      <th>Položka</th>
      <th>Uživatel</th>
      <th>Datum</th>
    </tr>
    {{range $item := .Items}}
      <tr>
        <td><a href="{{$item.ActivityURL}}">{{$item.ID}}</a></td>
        <td>{{$item.ActionType}}</td>
        <td><a href="{{$item.ItemURL}}">{{$item.ItemName}}</a></td>
        <td><a href="{{$item.UserURL}}">{{$item.UserName}}</a></td>
        <td>{{$item.CreatedAt}}</td>
      </tr>
    {{end}}
  </table>
{{end}}{{define "admin_home"}}
  {{tmpl "admin_navigation_page_content" .navigation}}

  {{$global := .}}
  {{range $snippet := .snippets}}
    {{tmpl $snippet.Template $global}}
  {{end}}
{{end}}


{{define "admin_home_navigation"}}
  <table class="admin_table">
    {{range $item := .}}
      <tr>
        <td>
          <b>{{$item.Name}}</b> ({{$item.Count}}x)
        </td>
        <td>
          <div class="btngroup">
            {{range $action := $item.Actions}}
              <a class="btn" href="{{$action.Url}}">{{$action.Name}}</a>
            {{end}}
          </div>
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
<div class="form_label">
  <span class="form_label_text">{{.NameHuman}}</span>
  <div class="admin_markdown">

    <div class="btngroup">
      <div class="btn btn-small admin_markdown_command" data-cmd="b" title="ctrl+b">B</div>
      <div class="btn btn-small admin_markdown_command" data-cmd="i" title="ctrl+i">I</div>
      <div class="btn btn-small admin_markdown_command" data-cmd="a" title="ctrl+u">Odkaz</div>
      <div class="btn btn-small admin_markdown_command" data-cmd="h2" title="ctrl+k">Nadpis</div>
    </div>

    &nbsp;&nbsp;<a href="" target="_blank" class="admin_markdown_show_help">Zobrazit nápovědu</a>&nbsp;&nbsp;

    <label>
      <input type="checkbox" class="admin_markdown_preview_show"> Zobrazit náhled
    </label>

    <textarea name="{{.Name}}" class="input form_input textarea"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>{{.Value}}</textarea>
    <div class="admin_markdown_preview hidden"></div>
  </div>
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
  <input type="submit" name="{{.Name}}" value="{{.NameHuman}}" class="btn btn-primary"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
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
<html lang="{{.admin_header.Language}}">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{.admin_title}}</title>
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
    <script src="https://maps.googleapis.com/maps/api/js?callback=bindPlaces&libraries=places&key={{.google}}" async defer></script>

  </head>
  <body class="admin" data-csrf-token="{{._csrfToken}}" data-admin-prefix="{{.admin_header.UrlPrefix}}"
    {{if .admin_header.Background}}
      style="background: linear-gradient(180deg, rgba(0, 0, 0, 0.0), rgba(0,0,0, 0.6) 100%), url('{{.admin_header.Background}}'); background-size: cover; background-attachment: fixed;" 
    {{end}}
    >
    {{tmpl "admin_flash" .}}
    <div class="admin_header">
        <div class="admin_header_top">
            <a href="{{.admin_header.UrlPrefix}}" class="admin_header_name admin_header_top_item{{if .admin_header_home_selected}} admin_header_top_item-active{{end}}">
              {{if .admin_header.Logo}}
                <div class="admin_logo" style="background-image: url('{{.admin_header.Logo}}');"></div>
              {{end}}
            {{message .locale "admin_admin"}} — {{.admin_header.Name}}</a>
            <a href="/" class="admin_header_top_item">{{.admin_header.HomepageUrl}}</a>
            <div class="admin_header_top_item admin_header_top_space"></div>
            <div class="admin_header_top_item">{{.currentuser.Email}}</div>
            <a href="{{.admin_header.UrlPrefix}}/user/settings"
                class="admin_header_top_item{{if .admin_header_settings_selected}} admin_header_top_item-active{{end}}">
                {{message .locale "admin_settings"}}
            </a>
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
        {{if .template_before}}
            {{tmpl .template_before .}}
        {{end}}
        {{tmpl .admin_yield .}}
        {{if .template_after}}
            {{tmpl .template_after .}}
        {{end}}
    </div>
  </body>
</html>

{{end}}{{define "admin_layout_nologin"}}
<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>{{.admin_title}}</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="{{.admin_header_prefix}}/normalize.css?v={{.version}}">
    <link rel="stylesheet" href="{{.admin_header_prefix}}/_static/admin.css?v={{.version}}">
  </head>
  <body class="admin_nologin"
    {{if .background}}
      style="background: linear-gradient(180deg, rgba(255, 255, 255, 0.0), rgba(255,255,255, 0.0) 100%), url('{{.background}}'); background-size: cover; background-attachment: fixed;" 
    {{end}}
  >
    {{tmpl "admin_flash" .}}


    {{tmpl .admin_yield .}}

  </body>
</html>

{{end}}{{define "admin_list"}}

{{$locale := .locale}}

{{$csrfToken := ._csrfToken}}
{{$table := .admin_list}}

{{$global := .}}
{{range $snippet := .admin_resource.Snippets}}
  {{tmpl $snippet.Template $global}}
{{end}}

{{$list := .admin_list}}


{{tmpl "admin_navigation" .navigation}}
<table class="admin_table admin_table-list {{if .admin_list.CanChangeOrder}} admin_table-order{{end}}" data-type="{{.admin_list.TypeID}}" data-order-column="{{.admin_list.OrderColumn}}" data-order-desc="{{.admin_list.OrderDesc}}">
  <thead>
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
  <th>
    <span class="admin_table_count"></span>
  </th>
  </tr>
  <tr>
    {{range $item := .admin_list.Header}}
      <th>
        {{if $item.FilterLayout}}
          {{tmpl $item.FilterLayout $item}}
        {{end}}
      </th>
    {{end}}
    <th>
      <progress class="admin_table_progress"></progress>
    </th>
  </tr>
  </thead>
  <tbody></tbody>
</table>
</div>
{{end}}

{{define "filter_layout_text"}}
  <input class="input input-small admin_table_filter_item" data-typ="{{.ColumnName}}">
{{end}}

{{define "filter_layout_relation"}}
  <select class="input input-small admin_table_filter_item admin_table_filter_item-relations" data-typ="{{.ColumnName}}">
    <option value="" selected=""></option>
  </select>
{{end}}

{{define "filter_layout_number"}}
  <input class="input input-small admin_table_filter_item" data-typ="{{.ColumnName}}">
{{end}}

{{define "filter_layout_boolean"}}
  <select class="input input-small admin_table_filter_item" data-typ="{{.ColumnName}}">
    <option value=""></option>
    <option value="true">✅</option>
    <option value="false">-</option>
  </select>
{{end}}

{{define "admin_list_cells"}}
  {{range $item := .admin_list.Rows}}
    <tr data-id="{{$item.ID}}" data-url="{{$item.URL}}" class="admin_table_row">
      {{range $cell := $item.Items}}
      <td>
        {{ tmpl $cell.TemplateName $cell }}
      </td>
      {{end}}
      <td nowrap class="top align-right">
        <div class="btngroup admin_list_buttons">
          {{range $action := $item.Actions.VisibleButtons}}
            <a href="{{$action.Url}}" class="btn btn-small"
              {{range $k, $v := $action.Params}} {{HTMLAttr $k}}="{{$v}}"{{end}}
            >{{$action.Name}}</a>
          {{end}}
          {{if $item.Actions.ShowOrderButton}}
            <a href="" class="btn btn-small admin-action-order preventredirect">☰</a>
          {{end}}
          {{if $item.Actions.MenuButtons}}
            <button class="btn preventredirect btn-small btn-more">
              <div class="preventredirect">▼</div>
              <div class="btn-more_content preventredirect">
                {{range $action := $item.Actions.MenuButtons}}
                  <a href="{{$action.Url}}" class="btn btn-small btn-more_content_item">{{$action.Name}}</a>
                {{end}}
              </div>
            </button>
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
{{define "admin_message"}}

<div class="admin_box">
  <h1>{{.message}}</h1>
</div>
{{end}}{{define "admin_navigation"}}
  <div class="admin_navigation_tabs{{if .Wide}} admin_navigation_tabs-wide{{end}}">
    {{range $item := .Tabs}}
      <div class="admin_navigation_tab{{if $item.Selected}} admin_navigation_tab-selected{{end}}">
        <a href="{{$item.URL}}">{{$item.Name}}</a>
      </div>
    {{end}}
  </div>

  <div class="admin_box admin_box-navigation{{if .Wide}} admin_box-wide{{end}}">
    <div class="admin_box_header">
      {{if .Breadcrumbs}}
        <div class="admin_navigation_breadcrumbs">
        {{range $item := .Breadcrumbs}}
          <div class="admin_navigation_breadcrumb">
            <a href="{{$item.URL}}">{{$item.Name}}</a>
          </div>
        {{end}}
        </div>
      {{end}}

      <h1>{{.Name}}</h1>
    </div>
{{end}}

{{define "admin_navigation_page"}}
    {{tmpl "admin_navigation_page_content" .admin_page}}
  </div>
{{end}}

{{define "admin_navigation_page_content"}}
    {{tmpl "admin_navigation" .Navigation}}
    {{tmpl .PageTemplate .PageData}}
  </div>
{{end}}{{define "admin_settings_OLD"}}

<div class="admin_box">
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


{{end}}{{define "admin_view"}}
  {{range $item := .Items}}
    <div class="view_name">
      {{$item.Name}}
    </div>
    <div class="view_content">
      {{tmpl $item.Template $item.Value}}
    </div>
  {{end}}
  </div>
{{end}}

{{define "admin_item_view_text"}}
  {{.}}
{{end}}

{{define "admin_item_view_boolean"}}
  {{if .}}✅{{else}}-{{end}}
{{end}}

{{define "admin_item_view_markdown"}}
  {{markdown .}}
{{end}}

{{define "admin_item_view_image"}}
  <div class="admin_item_view_image_content" data-images="{{.}}">
    <progress value="" max=""></progress>
  </div>
{{end}}

{{define "admin_item_view_place"}}
  <div class="admin_item_view_place" data-value="{{.}}">
    <progress value="" max=""></progress>
  </div>
{{end}}

{{define "admin_item_view_relation"}}
  <div class="admin_item_view_relation" data-type="{{.Typ}}" data-id="{{.ID}}">
    <progress value="" max=""></progress>
  </div>
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

    {{if .show_back_button}}
      <br><br>
      <a href="/">Vrátit se zpět na stránky {{.site}}</a>
    {{end}}
    </div>
  </body>
</html>

{{end}}{{define "newsletter_send"}}
  <form method="POST" action="send">

    <div>
      <b>Emailové adresy ({{.recipients_count}})</b>
    </div>
    {{if false}}
      {{range $item := .recipients}}
        <div>{{$item}}</div>
      {{end}}
    {{end}}

    <input type="submit" class="btn" value="Odeslat newsletter">
  </form>
{{end}}{{define "newsletter_send_preview"}}
  <form method="POST" action="send-preview">
    <label>
      Seznam emailů na poslání preview (jeden email na řádek)
      <textarea class="input" name="emails"></textarea>
    </label>

    <input type="submit" class="btn">
  </form>
{{end}}{{define "newsletter_sent"}}

<div class="admin_box">
  <h1>Newsletter odeslán</h1>

  <b>Emailové adresy ({{.recipients_count}})</b>
  {{if false}}
    {{range $item := .recipients}}
      <div>{{$item}}</div>
    {{end}}
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
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
  font-size: 14px;
  line-height: 1.4em;
  color: #444;
}
body {
  background-color: #f3f3f3;
  background-size: cover;
  background-attachment: fixed;
}
.admin_nologin .admin_navigation_tabs {
  margin-top: 50px;
}
.shadow {
  box-shadow: 0px 1px 2px 0px rgba(0, 0, 0, 0.1);
}
code {
  background-color: #fafafa;
  padding: 0px 3px;
  border-radius: 3px;
  border: 1px solid #ddd;
}
h1 {
  font-size: 2rem;
  line-height: 1.4em;
  margin: 0.2rem 0 0.2rem 0;
}
h2 {
  font-size: 1.4rem;
  line-height: 1.4em;
  margin: 1rem 0 0.2rem 0;
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
.align-right {
  text-align: right;
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
  display: none !important;
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
  padding-bottom: 50px;
}
.admin_box {
  box-shadow: 0px 1px 2px 0px rgba(0, 0, 0, 0.1);
  margin: 5px auto;
  background-color: white;
  padding: 10px;
  border-radius: 2px;
  max-width: 600px;
}
.admin_box-wide {
  max-width: none;
  width: 100%;
  padding: 0px;
}
.admin_box-wide .admin_box_header {
  padding: 10px;
  padding-bottom: 0px;
}
.btn {
  display: inline-block;
  padding: 6px 12px;
  font-size: 14px;
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
.btn-small {
  padding: 3px 12px;
  font-size: 12px;
}
.btngroup {
  display: inline-flex;
}
.btngroup > .btn:not(:last-child) {
  border-right: none;
  border-top-right-radius: 0px;
  border-bottom-right-radius: 0px;
}
.btngroup > .btn:not(:first-child) {
  border-top-left-radius: 0px;
  border-bottom-left-radius: 0px;
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
.btn-primary {
  background-image: linear-gradient(-180deg, #1E90FF, #005A9C 90%);
  color: white;
  border-color: #1E90FF;
}
.btn-primary:hover {
  background-image: linear-gradient(-180deg, #1E90FF, #005A9C 50%);
  border-color: #1E90FF;
}
.btn-primary:active {
  background-image: linear-gradient(-180deg, #1E90FF, #005A9C 0%);
  border-color: #005A9C;
}
.form {
  margin-top: 10px;
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
.form_label_text {
  font-weight: 600;
  line-height: 2em;
}
.form_label-required .form_label_text {
  font-weight: 800;
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
  padding: 8px 6px;
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
  background-color: #fafafa;
}
.input-small {
  padding: 2px 6px;
}
.input-placesearch {
  max-width: calc(100% - 200px);
  margin-left: 10px;
}
select.input {
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
}
select.admin_table_filter_item {
  width: auto;
}
.input[readonly],
.textarea[readonly],
.input[disabled],
.textarea[disabled],
.input[readonly]:focus,
.textarea[readonly]:focus,
.input[disabled]:focus,
.textarea[disabled]:focus {
  border-color: #eee;
  background: #fafafa;
  color: #999;
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
  background-color: white;
  width: 100%;
}
.admin_table thead {
  font-weight: bold;
}
.admin_table td {
  padding: 2px;
  border: 1px solid #f1f1f1;
}
.admin_table-list {
  margin-top: 10px;
}
.admin_table-list td {
  border-left: none;
  border-bottom: none;
  border-top: none;
}
.admin_table_row:nth-child(even) {
  background-color: #fafafa;
}
.admin_table_row:hover {
  background-color: rgba(64, 120, 192, 0.1);
  cursor: pointer;
}
.admin_table_row td {
  padding: 2px 5px;
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
.admin_list_buttons {
  opacity: 0;
}
.admin_table_row:hover .admin_list_buttons {
  opacity: 1;
}
.flash_messages {
  text-align: center;
  position: fixed;
  top: 0px;
  left: 0px;
  right: 0px;
  z-index: 10;
}
@keyframes example {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
.flash_message {
  font-weight: bold;
  color: #444;
  display: inline-block;
  margin: 5px 10px;
  box-shadow: 0px 1px 2px 0px rgba(0, 0, 0, 0.1);
  border-radius: 3px;
  animation-name: example;
  animation-duration: 300ms;
  animation-timing-function: ease-in;
  background-color: #FFD800;
  display: inline-flex;
  align-items: center;
}
.flash_message_content {
  border-radius: 3px;
  padding: 5px 20px;
  flex-grow: 2;
}
.flash_message_close {
  flex-shrink: 0;
  padding: 5px;
  font-weight: normal;
  color: #4078c0;
  opacity: .2;
  cursor: pointer;
}
.flash_message:hover .flash_message_close {
  opacity: 1;
}
.pagination {
  text-align: center;
  background-color: white;
}
.pagination_page,
.pagination_page_current {
  color: #4078c0;
  text-transform: uppercase;
  padding: 5px;
  margin: 0px;
  margin-right: 1px;
  font-size: 16px;
  line-height: 1.4em;
  display: inline-block;
  border-bottom: 2px solid none;
  text-decoration: none;
}
.pagination_page_current,
.pagination_page:hover {
  border-bottom: 2px solid #4078c0;
  padding-bottom: 3px;
  background-color: rgba(64, 120, 192, 0.1);
}
.pagination_page_current {
  font-weight: bold;
  cursor: default;
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
  border: 1px dashed #eee;
  padding: 3px;
  border-radius: 3px;
  padding: 10px;
  background-color: #fafafa;
  box-shadow: 0px 1px 2px 0px rgba(0, 0, 0, 0.1);
}
.admin_images_fileinput-droparea {
  border: 3px dashed #aaa;
  background-color: #fafafa;
}
.admin_place_map {
  height: 300px;
}
/*markdown*/
.admin_markdown textarea {
  margin-top: 5px;
  white-space: pre-wrap;
}
.admin_markdown_preview {
  margin-top: 5px;
  background: white;
  border: 1px dashed #e5e5e5;
  padding: 5px;
  font-size: 0.9em;
  line-height: 1.4em;
  color: #333;
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
.view_name {
  font-weight: bold;
  margin-left: 5px;
  margin-top: 10px;
}
.view_content {
  margin-bottom: 10px;
  background-color: rgba(64, 120, 192, 0.1);
  padding: 5px;
  border-radius: 3px;
  word-wrap: break-word;
}
.admin_item_view_place {
  height: 200px;
}
progress {
  padding: 8px;
  margin: 0px auto;
  line-height: 100px;
  display: block;
  background: none;
  -webkit-appearance: none;
  appearance: none;
  border: 4px solid #eee;
  border-top: 4px solid #bbb;
  border-radius: 50%;
  width: 5px;
  height: 5px;
  animation: spin 800ms linear infinite;
}
@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
.admin_navigation_breadcrumbs {
  display: flex;
  margin-bottom: 5px;
  font-size: 1.1rem;
}
.admin_navigation_breadcrumb:after {
  content: ">";
  color: #999;
  color: #4078c0;
  margin-right: 3px;
  font-weight: bold;
}
.admin_box-navigation {
  border-top-right-radius: 0px;
  margin-top: 0px;
}
.admin_navigation_tabs {
  border: 0px solid red;
  margin: 0 auto;
  margin-top: 5px;
  max-width: 600px;
  display: flex;
  justify-content: flex-end;
  vertical-align: bottom;
}
.admin_navigation_tabs-wide {
  max-width: none;
}
.admin_navigation_tab {
  margin-top: 3px;
  display: flex;
  background-color: rgba(255, 255, 255, 0.9);
  border-bottom: none;
  margin-left: 1px;
  border-bottom: 1px solid #eee;
  border-top-left-radius: 3px;
  border-top-right-radius: 3px;
}
.admin_navigation_tab:hover a {
  background-color: rgba(64, 120, 192, 0.1);
}
.admin_navigation_tab-selected {
  margin-top: 0px;
  background-color: #fff;
  border-bottom: 1px solid #fff;
  font-weight: bold;
  border-top: 2px solid #4078c0;
  padding: 0px 10px;
}
.admin_navigation_tab-selected:hover a {
  background-color: white;
}
.admin_navigation_tab a {
  text-decoration: none;
  padding: 7px 10px;
  display: inline-block;
}
.btn-more {
  position: relative;
  padding-left: 5px;
  padding-right: 5px;
  color: #999;
  font-weight: normal;
}
.btn-more_content {
  border: 0px solid red;
  display: none;
  position: absolute;
  right: -1px;
  border-radius: 3px;
  top: 21px;
  box-shadow: 0px 1px 2px 0px rgba(0, 0, 0, 0.1);
  z-index: 2;
  flex-flow: column;
}
.btn-more:hover {
  color: #444;
  border-bottom-right-radius: 0px;
}
.btn-more:hover .btn-more_content {
  display: flex;
}
.btn-more_content_item {
  display: block;
}
.btn-more_content_item:not(:last-child) {
  border-bottom-right-radius: 0px;
  border-bottom-left-radius: 0px;
}
.btn-more_content_item {
  border-top: none;
  border-top-right-radius: 0px;
  border-top-left-radius: 0px;
}
.btn-more_content_item:hover {
  background-color: rgba(64, 120, 192, 0.1);
}
.admin_header {
  background: white;
  font-size: 14px;
  padding-bottom: 0px;
  box-shadow: 0px 1px 2px 0px rgba(0, 0, 0, 0.1);
  position: relative;
  z-index: 2;
  line-height: 1.6em;
  flex-grow: 0;
  flex-shrink: 0;
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
.admin_header_top_item-active {
  font-weight: bold;
  border-bottom: 2px solid #4078c0;
  background-color: rgba(64, 120, 192, 0.1);
  padding-bottom: 3px;
}
.admin_header a {
  text-decoration: none;
}
.admin_header a:hover {
  background-color: rgba(64, 120, 192, 0.1);
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
  padding: 0px 5px;
  flex-wrap: wrap;
  overflow-x: auto;
}
.admin_header_resource {
  text-transform: uppercase;
  padding: 5px;
  margin: 0px;
  margin-right: 1px;
  font-size: 12px;
  border-bottom: 2px solid none;
  flex-shrink: 0;
}
.admin_header_resource-active {
  font-weight: bold;
  border-bottom: 2px solid #4078c0;
  padding-bottom: 3px;
  background-color: rgba(64, 120, 192, 0.1);
}
`


const adminJS = `
var Autoresize = (function () {
    function Autoresize(el) {
        return;
    }
    Autoresize.prototype.delayedResize = function () {
        var self = this;
        setTimeout(function () { self.resizeIt(); }, 0);
    };
    Autoresize.prototype.resizeIt = function () {
        this.el.style.height = 'auto';
        this.el.style.height = this.el.scrollHeight + 'px';
    };
    return Autoresize;
}());
function DOMinsertChildAtIndex(parent, child, index) {
    if (index >= parent.children.length) {
        parent.appendChild(child);
    }
    else {
        parent.insertBefore(child, parent.children[index]);
    }
}
function bindImageViews() {
    var els = document.querySelectorAll(".admin_item_view_image_content");
    for (var i = 0; i < els.length; i++) {
        new ImageView(els[i]);
    }
}
var ImageView = (function () {
    function ImageView(el) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.el = el;
        var ids = el.getAttribute("data-images").split(",");
        this.addImages(ids);
    }
    ImageView.prototype.addImages = function (ids) {
        this.el.innerHTML = "";
        for (var i = 0; i < ids.length; i++) {
            this.addImage(ids[i]);
        }
    };
    ImageView.prototype.addImage = function (id) {
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("target", "_blank");
        container.setAttribute("href", this.adminPrefix + "/file/uuid/" + id);
        var img = document.createElement("img");
        img.setAttribute("src", this.adminPrefix + "/_api/image/thumb/" + id);
        img.setAttribute("draggable", "false");
        container.appendChild(img);
        this.el.appendChild(container);
    };
    return ImageView;
}());
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
        this.fileInput.addEventListener("dragenter", function (ev) {
            _this.fileInput.classList.add("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragleave", function (ev) {
            _this.fileInput.classList.remove("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragover", function (ev) {
            ev.preventDefault();
        });
        this.fileInput.addEventListener("drop", function (ev) {
            var text = ev.dataTransfer.getData('Text');
            return;
        });
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
            if (!droppedElement) {
                droppedElement = e.originalTarget;
            }
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
            e.preventDefault();
            return false;
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
        this.progress = el.querySelector(".admin_table_progress");
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
        this.progress.classList.remove("hidden");
        var request = new XMLHttpRequest();
        request.open("POST", this.adminPrefix + "/_api/list/" + this.typeName + document.location.search, true);
        request.addEventListener("load", function () {
            _this.tbody.innerHTML = "";
            if (request.status == 200) {
                _this.tbody.innerHTML = request.response;
                var count = request.getResponseHeader("X-Count");
                var totalCount = request.getResponseHeader("X-Total-Count");
                var countStr = count + " / " + totalCount;
                _this.el.querySelector(".admin_table_count").textContent = countStr;
                bindOrder();
                _this.bindPagination();
                _this.bindClick();
            }
            else {
                console.error("error while loading list");
            }
            _this.progress.classList.add("hidden");
        });
        var requestData = this.getListRequest();
        request.send(JSON.stringify(requestData));
    };
    List.prototype.bindPagination = function () {
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
    List.prototype.bindClick = function () {
        var rows = this.el.querySelectorAll(".admin_table_row");
        for (var i = 0; i < rows.length; i++) {
            var row = rows[i];
            var id = row.getAttribute("data-id");
            row.addEventListener("click", function (e) {
                var target = e.target;
                if (target.classList.contains("preventredirect")) {
                    return;
                }
                var el = e.currentTarget;
                var url = el.getAttribute("data-url");
                window.location = url;
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
        this.bindFilterRelations();
        this.filterInputs = this.el.querySelectorAll(".admin_table_filter_item");
        for (var i = 0; i < this.filterInputs.length; i++) {
            var input = this.filterInputs[i];
            input.addEventListener("change", this.inputListener.bind(this));
            input.addEventListener("keyup", this.inputListener.bind(this));
        }
        this.inputPeriodicListener();
    };
    List.prototype.inputListener = function (e) {
        if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
            return;
        }
        this.page = 1;
        this.changed = true;
        this.changedTimestamp = Date.now();
        this.progress.classList.remove("hidden");
    };
    List.prototype.bindFilterRelations = function () {
        var els = this.el.querySelectorAll(".admin_table_filter_item-relations");
        for (var i = 0; i < els.length; i++) {
            this.bindFilterRelation(els[i]);
        }
    };
    List.prototype.bindFilterRelation = function (select) {
        var typ = select.getAttribute("data-typ");
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/resource/" + typ, true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                var resp = JSON.parse(request.response);
                for (var _i = 0, resp_1 = resp; _i < resp_1.length; _i++) {
                    var item = resp_1[_i];
                    var option = document.createElement("option");
                    option.setAttribute("value", item.id);
                    option.innerText = item.name;
                    select.appendChild(option);
                }
            }
            else {
                console.error("Error wile loading relation " + typ + ".");
            }
        });
        request.send();
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
    var elements = document.querySelectorAll(".admin_markdown");
    Array.prototype.forEach.call(elements, function (el, i) {
        new MarkdownEditor(el);
    });
}
var MarkdownEditor = (function () {
    function MarkdownEditor(el) {
        var _this = this;
        this.el = el;
        this.textarea = el.querySelector(".textarea");
        this.preview = el.querySelector(".admin_markdown_preview");
        new Autoresize(this.textarea);
        var prefix = document.body.getAttribute("data-admin-prefix");
        var helpLink = el.querySelector(".admin_markdown_show_help");
        helpLink.setAttribute("href", prefix + "/_help/markdown");
        this.lastChanged = Date.now();
        this.changed = false;
        var showChange = el.querySelector(".admin_markdown_preview_show");
        showChange.addEventListener("change", function () {
            _this.preview.classList.toggle("hidden");
        });
        setInterval(function () {
            if (_this.changed && (Date.now() - _this.lastChanged > 500)) {
                _this.loadPreview();
            }
        }, 100);
        this.textarea.addEventListener("change", this.textareaChanged.bind(this));
        this.textarea.addEventListener("keyup", this.textareaChanged.bind(this));
        this.loadPreview();
        this.bindCommands();
        this.bindShortcuts();
    }
    MarkdownEditor.prototype.bindCommands = function () {
        var _this = this;
        var btns = this.el.querySelectorAll(".admin_markdown_command");
        for (var i = 0; i < btns.length; i++) {
            btns[i].addEventListener("mousedown", function (e) {
                var cmd = e.target.getAttribute("data-cmd");
                _this.executeCommand(cmd);
                e.preventDefault();
                return false;
            });
        }
    };
    MarkdownEditor.prototype.bindShortcuts = function () {
        var _this = this;
        this.textarea.addEventListener("keydown", function (e) {
            if (e.metaKey == false && e.ctrlKey == false) {
                return;
            }
            switch (e.keyCode) {
                case 66:
                    _this.executeCommand("b");
                    break;
                case 73:
                    _this.executeCommand("i");
                    break;
                case 75:
                    _this.executeCommand("h2");
                    break;
                case 85:
                    _this.executeCommand("a");
                    break;
            }
        });
    };
    MarkdownEditor.prototype.executeCommand = function (commandName) {
        switch (commandName) {
            case "b":
                this.setAroundMarkdown("**", "**");
                break;
            case "i":
                this.setAroundMarkdown("*", "*");
                break;
            case "a":
                this.setAroundMarkdown("[", "]()");
                var newEnd = this.textarea.selectionEnd + 2;
                this.textarea.selectionStart = newEnd;
                this.textarea.selectionEnd = newEnd;
                break;
            case "h2":
                var start = "## ";
                var end = "";
                var text = this.textarea.value;
                if (text[this.textarea.selectionStart - 1] !== "\n") {
                    start = "\n" + start;
                }
                if (text[this.textarea.selectionEnd] !== "\n") {
                    end = "\n";
                }
                this.setAroundMarkdown(start, end);
                break;
        }
        this.textareaChanged();
    };
    MarkdownEditor.prototype.setAroundMarkdown = function (before, after) {
        var text = this.textarea.value;
        var selected = text.substr(this.textarea.selectionStart, this.textarea.selectionEnd - this.textarea.selectionStart);
        var newText = text.substr(0, this.textarea.selectionStart);
        newText += before;
        var newStart = newText.length;
        newText += selected;
        var newEnd = newText.length;
        newText += after;
        newText += text.substr(this.textarea.selectionEnd, text.length);
        this.textarea.value = newText;
        this.textarea.selectionStart = newStart;
        this.textarea.selectionEnd = newEnd;
        this.textarea.focus();
    };
    MarkdownEditor.prototype.textareaChanged = function () {
        this.changed = true;
        this.lastChanged = Date.now();
    };
    MarkdownEditor.prototype.loadPreview = function () {
        var _this = this;
        this.changed = false;
        var request = new XMLHttpRequest();
        request.open("POST", document.body.getAttribute("data-admin-prefix") + "/_api/markdown", true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                _this.preview.innerHTML = JSON.parse(request.response);
            }
            else {
                console.error("Error while loading markdown preview.");
            }
        });
        request.send(this.textarea.value);
    };
    return MarkdownEditor;
}());
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
function bindRelationsView() {
    var els = document.querySelectorAll(".admin_item_view_relation");
    for (var i = 0; i < els.length; i++) {
        new RelationsView(els[i]);
    }
}
var RelationsView = (function () {
    function RelationsView(el) {
        var idStr = el.getAttribute("data-id");
        var typ = el.getAttribute("data-type");
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/resource/" + typ + "/" + idStr, true);
        request.addEventListener("load", function () {
            el.innerHTML = "";
            if (request.status == 200) {
                var resp = JSON.parse(request.response);
                var link = document.createElement("a");
                link.setAttribute("href", adminPrefix + "/" + typ + "/" + idStr);
                var name = resp.name;
                if (name == "") {
                    name += " ";
                }
                link.textContent = name;
                el.appendChild(link);
            }
            else {
                el.textContent = "Error while loading";
            }
        });
        request.send();
    }
    return RelationsView;
}());
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
function bindPlacesView() {
    var els = document.querySelectorAll(".admin_item_view_place");
    for (var i = 0; i < els.length; i++) {
        new PlacesView(els[i]);
    }
}
var PlacesView = (function () {
    function PlacesView(el) {
        var val = el.getAttribute("data-value");
        el.innerText = "";
        var coords = val.split(",");
        if (coords.length != 2) {
            return;
        }
        var position = { lat: parseFloat(coords[0]), lng: parseFloat(coords[1]) };
        var zoom = 11;
        var map = new google.maps.Map(el, {
            center: position,
            zoom: zoom
        });
        var marker = new google.maps.Marker({
            position: position,
            map: map,
            draggable: true,
            title: ""
        });
    }
    return PlacesView;
}());
function bindPlaces() {
    bindPlacesView();
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
        var searchInput = document.createElement("input");
        searchInput.classList.add("input", "input-placesearch");
        var searchBox = new google.maps.places.SearchBox(searchInput);
        map.controls[google.maps.ControlPosition.LEFT_TOP].push(searchInput);
        searchBox.addListener('places_changed', function () {
            var places = searchBox.getPlaces();
            if (places.length > 0) {
                map.fitBounds(places[0].geometry.viewport);
                marker.setPosition({ lat: places[0].geometry.location.lat(), lng: places[0].geometry.location.lng() });
                marker.setVisible(true);
            }
        });
        searchInput.addEventListener("keydown", function (e) {
            if (e.keyCode == 13) {
                e.preventDefault();
                return false;
            }
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
function bindForm() {
    var els = document.querySelectorAll(".form_leavealert");
    for (var i = 0; i < els.length; i++) {
        new Form(els[i]);
    }
}
var Form = (function () {
    function Form(el) {
        var _this = this;
        this.allow = false;
        el.addEventListener("submit", function () {
            _this.allow = true;
        });
        window.addEventListener("beforeunload", function (e) {
            if (_this.allow) {
                return;
            }
            var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
            e.returnValue = confirmationMessage;
            return confirmationMessage;
        });
    }
    return Form;
}());
document.addEventListener("DOMContentLoaded", function () {
    bindMarkdowns();
    bindTimestamps();
    bindRelationsView();
    bindRelations();
    bindImagePickers();
    bindClickAndStay();
    bindLists();
    bindForm();
    bindImageViews();
    bindFlashMessages();
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
function bindFlashMessages() {
    var messages = document.querySelectorAll(".flash_message");
    for (var i = 0; i < messages.length; i++) {
        var message = messages[i];
        message.addEventListener("click", function (e) {
            var target = e.currentTarget;
            if (target.classList.contains("flash_message_close")) {
                target.classList.add("hidden");
            }
        });
    }
}
`


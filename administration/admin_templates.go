package administration
const adminTemplates = `
{{define "admin_delete"}}
  <h1>Chcete smazat tuto položku?</h1>
  {{template "admin_form" .form}}
{{end}}{{define "admin_export"}}
  
  <form method="POST" action="export">

    <h2>Fields</h2>
    {{range $field := .Fields}}
      <label class="form_label">
        <input type="checkbox" name="_field" value="{{$field.ColumnName}}" checked>
        <span class="form_label_text-inline">{{$field.NameHuman}}</span>
      </label>
    {{end}}

    <h2>Order By</h2>
    <select name="_order" class="input">
      {{$default := .DefaultOrderColumnName}}
      {{range $field := .Fields}}
        <option value="{{$field.ColumnName}}"{{if eq $field.ColumnName $default}} selected{{end}}>{{$field.NameHuman}}</option>
      {{end}}
    </select>
    <label class="form_label">
      <input type="checkbox" name="_desc" {{if .DefaultOrderDesc}} checked{{end}}>
      <span class="form_label_text-inline">Descending order</span>
    </label>

    <h2>Limit</h2>
    <input name="_limit" type="number" class="input">

    <h2>Filter</h2>
    {{range $field := .Fields}}
      {{if $field.FilterLayout}}
        <label class="form_label">
          <span class="form_label_text">{{$field.NameHuman}}</span>
          <div>
            {{tmpl $field.FilterLayout $field}}
          </div>
        </label>
      {{end}}
    {{end}}

    <div class="primarybtncontainer">
      <button class="btn btn-primary">Create Export</button>
    </div>
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

{{if .CSRFToken}}
  <input type="hidden" name="_csrfToken" value="{{.CSRFToken}}">
{{end}}

{{if .Errors}}
  <div class="form_errors">
    {{range $error := .Errors}}
      <div class="form_errors_error">{{$error}}</div>
    {{end}}
  </div>
{{end}}

{{range $item := .Items}}
  <div class="form_label{{if .Errors}} form_label-errors{{end}}{{if .Required}} form_label-required{{end}}">
    {{if eq .HiddenName false}}
      <label for="{{.UUID}}" class="form_label_text">{{.NameHuman}}</label>
    {{end}}
    {{if .Errors}}
      <div class="form_label_errors">
        {{range $error := .Errors}}
          <div class="form_label_errors_error">{{$error}}</div>
        {{end}}
      </div>
    {{end}}
    <div>
      {{tmpl $item.Template $item}}
    </div>
  </div>
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
{{end}}{{define "admin_home_navigation"}}
  <div class="admin_box_padding">  
    {{range $item := .}}
      <h2><a href="{{$item.URL}}">{{$item.Name}}</a></h2>
      <ul>
        {{range $action := $item.Actions}}
          <li><a href="{{$action.URL}}">{{$action.Name}}</a></li>
        {{end}}
      </ul>
    {{end}}
  </div>
{{end}}{{define "admin_item_input"}}
  <input name="{{.Name}}" value="{{.Value}}" id="{{.UUID}}" class="input form_watcher form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_email"}}
  <input name="{{.Name}}" value="{{.Value}}" type="email" class="input form_watcher form_input" spellcheck="false"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_password"}}
  <input name="{{.Name}}" value="{{.Value}}" type="password" class="input form_watcher form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_textarea"}}
  <textarea name="{{.Name}}" id="{{.UUID}}" class="input form_watcher form_input textarea"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>{{.Value}}</textarea>
{{end}}

{{define "admin_item_markdown"}}
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

    <textarea name="{{.Name}}" id="{{.UUID}}" class="input form_watcher form_input textarea"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>{{.Value}}</textarea>
    <div class="admin_markdown_preview hidden"></div>
  </div>
{{end}}

{{define "admin_item_checkbox"}}
  <label>
    <input type="checkbox" name="{{.Name}}" {{if .Value}}checked{{end}}{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}} class="form_watcher">
    <span class="form_label_text-inline">{{.NameHuman}}</span>
  </label>
{{end}}

{{define "admin_item_date"}}
  <input type="date" name="{{.Name}}" value="{{.Value}}" id="{{.UUID}}" class="input form_watcher form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_item_timestamp"}}
  {{if .Readonly}}
    <input name="{{.Name}}" value="{{.Value}}" class="input form_input"{{if .Focused}} autofocus{{end}} readonly>
  {{else}}
    <div class="admin_timestamp">
      <input type="hidden" id="{{.UUID}}" name="{{.Name}}" value="{{.Value}}">

      <input type="date" name="_admin_timestamp_hidden" class="input form_input admin_timestamp_date"{{if .Focused}} autofocus{{end}}>

      <select class="input form_watcher form_input admin_timestamp_hour"></select>
      <span class="admin_timestamp_divider">:</span>
      <select class="input form_watcher form_input admin_timestamp_minute"></select>

    </div>
  {{end}}
{{end}}

{{define "admin_item_image"}}
  <div class="admin_images">
    <input name="{{.Name}}" value="{{.Value}}" type="hidden" class="admin_images_hidden form_watcher">
    <div class="admin_images_loaded hidden">
      <input type="file" id="{{.UUID}}" accept=".jpg,.jpeg,.png" multiple class="admin_images_fileinput form_watcher">
      <div class="admin_images_preview"></div>
    </div>
    <progress></progress>
  </div>
{{end}}

{{define "admin_item_file"}}
  <input type="file" id="{{.UUID}}" name="{{.Name}}" class="input form_watcher form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
{{end}}

{{define "admin_file"}}
  {{if .Value}}
    <img src="{{thumb .Value}}">
  {{else}}
    <input type="file" id="{{.UUID}}" name="{{.Name}}" class="input form_watcher form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
  {{end}}
{{end}}

{{define "admin_item_submit"}}
  <div class="primarybtncontainer">
    <button id="{{.UUID}}" name="{{.Name}}" class="btn btn-primary"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>{{.NameHuman}}</button>
  </div>
{{end}}

{{define "admin_item_delete"}}
  <button id="{{.UUID}}" name="{{.Name}}" class="btn btn-primary btn-delete"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>{{.NameHuman}}</button>
{{end}}

{{define "admin_item_select"}}
  <select name="{{.Name}}" id="{{.UUID}}" class="input form_watcher form_input"{{if .Focused}} autofocus{{end}}{{if .Readonly}} readonly{{end}}>
    {{$val := .Value}}
    {{range $value := .Data}}
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
<div class="admin_item_relation" data-relation="{{.Data}}">
  <input type="hidden" name="{{.Name}}" value="{{.Value}}">
  <progress></progress>
  <div class="admin_item_relation_preview hidden"></div>
  <div class="admin_item_relation_change hidden">
    <div class="btn btn-small admin_item_relation_change_btn">×</div>
  </div>
  <div class="admin_item_relation_picker hidden">
    <input class="input">
    <div class="admin_item_relation_picker_suggestions">
      <div class="admin_item_relation_picker_suggestions_content">

      </div>
    </div>
  </div>
</div>
{{end}}

{{define "admin_list_image"}}
  <div class="admin_thumb">
    <img src="{{thumb .}}">
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
  <body class="admin" data-csrf-token="{{._csrfToken}}" data-admin-prefix="{{.admin_header.UrlPrefix}}">
    {{template "admin_flash" .}}
    <div class="admin_header">
        <div class="admin_header_top">
            <a href="{{.admin_header.UrlPrefix}}" class="admin_header_top_item admin_header_name">
                {{message .locale "admin_admin"}} – {{.admin_header.Name}}
            </a>
            <div class="admin_header_top_item admin_header_top_space"></div>
            <div class="admin_header_top_item">{{.currentuser.Email}}</div>
            <a href="{{.admin_header.UrlPrefix}}/user/settings" class="admin_header_top_item">
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
        {{tmpl .admin_yield .}}
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
  <body class="admin_nologin">
    {{tmpl "admin_flash" .}}
    {{tmpl .admin_yield .}}
  </body>
</html>

{{end}}{{define "admin_list"}}
  <table class="admin_table admin_table-list {{if .CanChangeOrder}} admin_table-order{{end}}"
  data-type="{{.TypeID}}"
  data-order-column="{{.OrderColumn}}"
  data-order-desc="{{.OrderDesc}}"
  data-prefilter-field="{{.PrefilterField}}"
  data-prefilter-value="{{.PrefilterValue}}"
  >
    <thead>
    <tr>
      {{range $item := .Header}}
        {{if $item.ShouldShow}}
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
      {{end}}
      <th>
        <span class="admin_table_count"></span>
      </th>
    </tr>
    <tr>
      {{range $item := .Header}}
        {{if $item.ShouldShow}}
          <th>
            {{if $item.FilterLayout}}
              {{tmpl $item.FilterLayout $item}}
            {{end}}
          </th>
        {{end}}
      {{end}}
      <th>
        <progress class="admin_table_progress"></progress>
      </th>
    </tr>
    </thead>
    <tbody></tbody>
  </table>
{{end}}

{{define "filter_layout_text"}}
  <input class="input input-small admin_table_filter_item" name="{{.ColumnName}}" data-typ="{{.ColumnName}}">
{{end}}

{{define "filter_layout_relation"}}
  <select class="input input-small admin_table_filter_item admin_table_filter_item-relations" name="{{.ColumnName}}" data-typ="{{.ColumnName}}">
    <option value="" selected=""></option>
  </select>
{{end}}

{{define "filter_layout_number"}}
  <input class="input input-small admin_table_filter_item" type="number" name="{{.ColumnName}}" data-typ="{{.ColumnName}}">
{{end}}

{{define "filter_layout_boolean"}}
  <select class="input input-small admin_table_filter_item" name="{{.ColumnName}}" data-typ="{{.ColumnName}}">
    <option value=""></option>
    <option value="true">{{index .FilterData 0}}</option>
    <option value="false">{{index .FilterData 1}}</option>
  </select>
{{end}}

{{define "filter_layout_date"}}
  <div class="admin_filter_layout_date">
    <input class="admin_table_filter_item admin_filter_layout_date_value" name="{{.ColumnName}}" data-typ="{{.ColumnName}}">
    <div class="admin_filter_layout_date_content">
      <input type="date" class="input input-small admin_filter_layout_date_from">
      <div class="admin_filter_layout_date_divider">–</div>
      <input type="date" class="input input-small admin_filter_layout_date_to">
    </div>
  </div>
{{end}}

{{define "admin_list_cells"}}
  {{range $item := .admin_list.Rows}}
    <tr data-id="{{$item.ID}}" data-url="{{$item.URL}}" class="admin_table_row">
      {{range $cell := $item.Items}}
      <td>
        {{tmpl $cell.Template $cell.Value}}
      </td>
      {{end}}
      <td nowrap class="top align-right">
        <div class="btngroup admin_list_buttons">
          {{range $action := $item.Actions.VisibleButtons}}
            <a href="{{$action.URL}}" class="btn btn-small"
              {{range $k, $v := $action.Params}} {{HTMLAttr $k}}="{{$v}}"{{end}}
            >{{$action.Name}}</a>
          {{end}}
          {{if $item.Actions.ShowOrderButton}}
            <a href="" class="btn btn-small admin-action-order preventredirect">☰</a>
          {{end}}
          {{if $item.Actions.MenuButtons}}
            <div class="btn preventredirect btn-small btn-more">
              <div class="preventredirect">▼</div>
              <div class="btn-more_content preventredirect">
                {{range $action := $item.Actions.MenuButtons}}
                  <a href="{{$action.URL}}" class="btn btn-small btn-more_content_item">{{$action.Name}}</a>
                {{end}}
              </div>
            </div>
          {{end}}
        </div>
      </td>
    </tr>
  {{end}}
  {{if .admin_list.Message}}
    <tr>
      <td colspan="{{.admin_list.Colspan}}" class="admin_list_message">
        {{.admin_list.Message}}
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
  <div class="admin_navigation_breadcrumbs">
    {{range $item := .Breadcrumbs}}
      <a href="{{$item.URL}}" class="admin_navigation_breadcrumb">
        {{- if $item.Logo -}}
          <div style="background-image: url('{{CSS $item.Logo}}');"
          class="admin_navigation_breadcrumb_image admin_navigation_breadcrumb_image-logo"></div>
        {{- end -}}
        {{- if $item.Image -}}
          <div style="background-image: url('{{CSS $item.Image}}');"
          class="admin_navigation_breadcrumb_image"></div>
        {{- end -}}
        {{- $item.Name -}}
        <div class="admin_navigation_breadcrumb_divider"></div>
      </a>
    {{end}}
  </div>

  <div class="admin_navigation_tabs">
    {{range $item := .Tabs}}
      <a href="{{$item.URL}}" class="admin_navigation_tab{{if $item.Selected}} admin_navigation_tab-selected{{end}}">
        {{$item.Name}}
      </a>
    {{end}}
  </div>

  <div class="admin_box{{if .Wide}} admin_box-wide{{end}}">
{{end}}

{{define "admin_navigation_page"}}
    {{tmpl "admin_navigation" .admin_page.Navigation}}
    {{tmpl .admin_page.PageTemplate .admin_page.PageData}}
  </div>
{{end}}{{define "admin_settings_OLD"}}

<div class="admin_box">
  {{tmpl "admin_form" .admin_form}}
</div>

{{end}}{{define "admin_stats"}}

<h2>Access view</h2>

<table class="admin_table">
  <tr>
    <td></td>
    {{range $role := .accessView.Roles}}
      <td>{{$role}}</td>
    {{end}}
  </tr>

  {{range $resource := .accessView.Resources}}
    <tr>
      <td>{{$resource.Name}}</td>
      {{range $role := $resource.Roles}}
        <td style="font-family: monospace;" nowrap>{{$role.Value}}</td>
      {{end}}
    </tr>
  {{end}}

</table>

<h2>Auth roles</h2>

<table class="admin_table">
{{range $role, $permissions := .roles}}
  <tr>
    <td>{{$role}}</td>
    <td>{{range $permission, $_ := $permissions}} {{$permission}}{{end}}</td>
  </tr>
{{end}}
</table>


<h2>Base stats</h2>

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
      {{- tmpl $item.Template $item.Value -}}
    </div>
  {{end}}
  </div>
{{end}}

{{define "admin_item_view_text"}}
  {{.}}
{{end}}

{{define "admin_item_view_markdown"}}
  {{markdown .}}
{{end}}

{{define "admin_item_view_file"}}
  {{if .MediumURL}}
    <div>
      <a href="{{.OriginalURL}}"><img src="{{.MediumURL}}"></a>
    </div>
  {{end}}
  <div>
    Download:
    {{range $path := .Paths}}
      <a href="{{$path.URL}}">{{$path.Name}}</a>
    {{end}}
  </div>
  <div>UUID: {{.UUID}}</div>
{{end}}

{{define "admin_item_view_file_cell"}}
  {{if .SmallURL}}
    <div class="admin_thumb">
      <img src="{{.SmallURL}}">
    </div>
  {{end}}
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
  <div class="admin_item_view_relation">
    {{if .}}
      <a class="admin_preview" href="{{.URL}}">
        <div class="admin_preview_image" style="background-image: url('{{CSS .Image}}') ;"></div>
        <div class="admin_preview_right">
          <div class="admin_preview_name">{{.Name}}</div>
          <div class="admin_preview_description">{{.Description}}</div>
        </div>
      </a>
    {{else}}
      –
    {{end}}
  </div>
{{end}}


{{define "admin_item_view_relation_cell"}}
  {{if .}}
    {{.Name}}
  {{else}}
    –
  {{end}}
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
@font-face {
  font-family: 'Glyphicons Regular';
  src: url('/fonts/glyphicons-regular.woff2') format('woff2'), url('/fonts/glyphicons-regular.woff') format('woff');
}
.glyphicon {
  display: inline-block;
  font-family: 'Glyphicons Regular';
  font-style: normal;
  font-weight: normal;
  line-height: 1;
  vertical-align: middle;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: red;
}
.glyphicons {
  position: relative;
  top: 1px;
  display: inline-block;
  font-family: 'Glyphicons Regular';
  font-style: normal;
  font-weight: normal;
  line-height: 1;
  vertical-align: middle;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
.glyphicons.x05 {
  font-size: 12px;
}
.glyphicons.x2 {
  font-size: 48px;
}
.glyphicons.x3 {
  font-size: 72px;
}
.glyphicons.x4 {
  font-size: 96px;
}
.glyphicons.x5 {
  font-size: 120px;
}
.glyphicons.light:before {
  color: #f2f2f2;
}
.glyphicons.drop:before {
  text-shadow: -1px 1px 3px rgba(0, 0, 0, 0.3);
}
.glyphicons.flip {
  -moz-transform: scaleX(-1);
  -o-transform: scaleX(-1);
  -webkit-transform: scaleX(-1);
  transform: scaleX(-1);
  filter: FlipH;
  -ms-filter: "FlipH";
}
.glyphicons.flipv {
  -moz-transform: scaleY(-1);
  -o-transform: scaleY(-1);
  -webkit-transform: scaleY(-1);
  transform: scaleY(-1);
  filter: FlipV;
  -ms-filter: "FlipV";
}
.glyphicons.rotate90 {
  -webkit-transform: rotate(90deg);
  -moz-transform: rotate(90deg);
  -ms-transform: rotate(90deg);
  transform: rotate(90deg);
}
.glyphicons.rotate180 {
  -webkit-transform: rotate(180deg);
  -moz-transform: rotate(180deg);
  -ms-transform: rotate(180deg);
  transform: rotate(180deg);
}
.glyphicons.rotate270 {
  -webkit-transform: rotate(270deg);
  -moz-transform: rotate(270deg);
  -ms-transform: rotate(270deg);
  transform: rotate(270deg);
}
.glyphicons-glass:before {
  content: "\E001";
}
.glyphicons-leaf:before {
  content: "\E002";
}
.glyphicons-dog:before {
  content: "\E003";
}
.glyphicons-user:before {
  content: "\E004";
}
.glyphicons-girl:before {
  content: "\E005";
}
.glyphicons-car:before {
  content: "\E006";
}
.glyphicons-user-add:before {
  content: "\E007";
}
.glyphicons-user-remove:before {
  content: "\E008";
}
.glyphicons-film:before {
  content: "\E009";
}
.glyphicons-magic:before {
  content: "\E010";
}
.glyphicons-envelope:before {
  content: "\2709";
}
.glyphicons-camera:before {
  content: "\E011";
}
.glyphicons-heart:before {
  content: "\E013";
}
.glyphicons-beach-umbrella:before {
  content: "\E014";
}
.glyphicons-train:before {
  content: "\E015";
}
.glyphicons-print:before {
  content: "\E016";
}
.glyphicons-bin:before {
  content: "\E017";
}
.glyphicons-music:before {
  content: "\E018";
}
.glyphicons-note:before {
  content: "\E019";
}
.glyphicons-heart-empty:before {
  content: "\E020";
}
.glyphicons-home:before {
  content: "\E021";
}
.glyphicons-snowflake:before {
  content: "\2744";
}
.glyphicons-fire:before {
  content: "\E023";
}
.glyphicons-magnet:before {
  content: "\E024";
}
.glyphicons-parents:before {
  content: "\E025";
}
.glyphicons-binoculars:before {
  content: "\E026";
}
.glyphicons-road:before {
  content: "\E027";
}
.glyphicons-search:before {
  content: "\E028";
}
.glyphicons-cars:before {
  content: "\E029";
}
.glyphicons-notes-2:before {
  content: "\E030";
}
.glyphicons-pencil:before {
  content: "\270F";
}
.glyphicons-bus:before {
  content: "\E032";
}
.glyphicons-wifi-alt:before {
  content: "\E033";
}
.glyphicons-luggage:before {
  content: "\E034";
}
.glyphicons-old-man:before {
  content: "\E035";
}
.glyphicons-woman:before {
  content: "\E036";
}
.glyphicons-file:before {
  content: "\E037";
}
.glyphicons-coins:before {
  content: "\E038";
}
.glyphicons-airplane:before {
  content: "\2708";
}
.glyphicons-notes:before {
  content: "\E040";
}
.glyphicons-stats:before {
  content: "\E041";
}
.glyphicons-charts:before {
  content: "\E042";
}
.glyphicons-pie-chart:before {
  content: "\E043";
}
.glyphicons-group:before {
  content: "\E044";
}
.glyphicons-keys:before {
  content: "\E045";
}
.glyphicons-calendar:before {
  content: "\E046";
}
.glyphicons-router:before {
  content: "\E047";
}
.glyphicons-camera-small:before {
  content: "\E048";
}
.glyphicons-star-empty:before {
  content: "\E049";
}
.glyphicons-star:before {
  content: "\E050";
}
.glyphicons-link:before {
  content: "\E051";
}
.glyphicons-eye-open:before {
  content: "\E052";
}
.glyphicons-eye-close:before {
  content: "\E053";
}
.glyphicons-alarm:before {
  content: "\E054";
}
.glyphicons-clock:before {
  content: "\E055";
}
.glyphicons-stopwatch:before {
  content: "\E056";
}
.glyphicons-projector:before {
  content: "\E057";
}
.glyphicons-history:before {
  content: "\E058";
}
.glyphicons-truck:before {
  content: "\E059";
}
.glyphicons-cargo:before {
  content: "\E060";
}
.glyphicons-compass:before {
  content: "\E061";
}
.glyphicons-keynote:before {
  content: "\E062";
}
.glyphicons-paperclip:before {
  content: "\E063";
}
.glyphicons-power:before {
  content: "\E064";
}
.glyphicons-lightbulb:before {
  content: "\E065";
}
.glyphicons-tag:before {
  content: "\E066";
}
.glyphicons-tags:before {
  content: "\E067";
}
.glyphicons-cleaning:before {
  content: "\E068";
}
.glyphicons-ruler:before {
  content: "\E069";
}
.glyphicons-gift:before {
  content: "\E070";
}
.glyphicons-umbrella:before {
  content: "\2602";
}
.glyphicons-book:before {
  content: "\E072";
}
.glyphicons-bookmark:before {
  content: "\E073";
}
.glyphicons-wifi:before {
  content: "\E074";
}
.glyphicons-cup:before {
  content: "\E075";
}
.glyphicons-stroller:before {
  content: "\E076";
}
.glyphicons-headphones:before {
  content: "\E077";
}
.glyphicons-headset:before {
  content: "\E078";
}
.glyphicons-warning-sign:before {
  content: "\E079";
}
.glyphicons-signal:before {
  content: "\E080";
}
.glyphicons-retweet:before {
  content: "\E081";
}
.glyphicons-refresh:before {
  content: "\E082";
}
.glyphicons-roundabout:before {
  content: "\E083";
}
.glyphicons-random:before {
  content: "\E084";
}
.glyphicons-heat:before {
  content: "\E085";
}
.glyphicons-repeat:before {
  content: "\E086";
}
.glyphicons-display:before {
  content: "\E087";
}
.glyphicons-log-book:before {
  content: "\E088";
}
.glyphicons-address-book:before {
  content: "\E089";
}
.glyphicons-building:before {
  content: "\E090";
}
.glyphicons-eyedropper:before {
  content: "\E091";
}
.glyphicons-adjust:before {
  content: "\E092";
}
.glyphicons-tint:before {
  content: "\E093";
}
.glyphicons-crop:before {
  content: "\E094";
}
.glyphicons-vector-path-square:before {
  content: "\E095";
}
.glyphicons-vector-path-circle:before {
  content: "\E096";
}
.glyphicons-vector-path-polygon:before {
  content: "\E097";
}
.glyphicons-vector-path-line:before {
  content: "\E098";
}
.glyphicons-vector-path-curve:before {
  content: "\E099";
}
.glyphicons-vector-path-all:before {
  content: "\E100";
}
.glyphicons-font:before {
  content: "\E101";
}
.glyphicons-italic:before {
  content: "\E102";
}
.glyphicons-bold:before {
  content: "\E103";
}
.glyphicons-text-underline:before {
  content: "\E104";
}
.glyphicons-text-strike:before {
  content: "\E105";
}
.glyphicons-text-height:before {
  content: "\E106";
}
.glyphicons-text-width:before {
  content: "\E107";
}
.glyphicons-text-resize:before {
  content: "\E108";
}
.glyphicons-left-indent:before {
  content: "\E109";
}
.glyphicons-right-indent:before {
  content: "\E110";
}
.glyphicons-align-left:before {
  content: "\E111";
}
.glyphicons-align-center:before {
  content: "\E112";
}
.glyphicons-align-right:before {
  content: "\E113";
}
.glyphicons-justify:before {
  content: "\E114";
}
.glyphicons-list:before {
  content: "\E115";
}
.glyphicons-text-smaller:before {
  content: "\E116";
}
.glyphicons-text-bigger:before {
  content: "\E117";
}
.glyphicons-embed:before {
  content: "\E118";
}
.glyphicons-embed-close:before {
  content: "\E119";
}
.glyphicons-table:before {
  content: "\E120";
}
.glyphicons-message-full:before {
  content: "\E121";
}
.glyphicons-message-empty:before {
  content: "\E122";
}
.glyphicons-message-in:before {
  content: "\E123";
}
.glyphicons-message-out:before {
  content: "\E124";
}
.glyphicons-message-plus:before {
  content: "\E125";
}
.glyphicons-message-minus:before {
  content: "\E126";
}
.glyphicons-message-ban:before {
  content: "\E127";
}
.glyphicons-message-flag:before {
  content: "\E128";
}
.glyphicons-message-lock:before {
  content: "\E129";
}
.glyphicons-message-new:before {
  content: "\E130";
}
.glyphicons-inbox:before {
  content: "\E131";
}
.glyphicons-inbox-plus:before {
  content: "\E132";
}
.glyphicons-inbox-minus:before {
  content: "\E133";
}
.glyphicons-inbox-lock:before {
  content: "\E134";
}
.glyphicons-inbox-in:before {
  content: "\E135";
}
.glyphicons-inbox-out:before {
  content: "\E136";
}
.glyphicons-cogwheel:before {
  content: "\E137";
}
.glyphicons-cogwheels:before {
  content: "\E138";
}
.glyphicons-picture:before {
  content: "\E139";
}
.glyphicons-adjust-alt:before {
  content: "\E140";
}
.glyphicons-database-lock:before {
  content: "\E141";
}
.glyphicons-database-plus:before {
  content: "\E142";
}
.glyphicons-database-minus:before {
  content: "\E143";
}
.glyphicons-database-ban:before {
  content: "\E144";
}
.glyphicons-folder-open:before {
  content: "\E145";
}
.glyphicons-folder-plus:before {
  content: "\E146";
}
.glyphicons-folder-minus:before {
  content: "\E147";
}
.glyphicons-folder-lock:before {
  content: "\E148";
}
.glyphicons-folder-flag:before {
  content: "\E149";
}
.glyphicons-folder-new:before {
  content: "\E150";
}
.glyphicons-edit:before {
  content: "\E151";
}
.glyphicons-new-window:before {
  content: "\E152";
}
.glyphicons-check:before {
  content: "\E153";
}
.glyphicons-unchecked:before {
  content: "\E154";
}
.glyphicons-more-windows:before {
  content: "\E155";
}
.glyphicons-show-big-thumbnails:before {
  content: "\E156";
}
.glyphicons-show-thumbnails:before {
  content: "\E157";
}
.glyphicons-show-thumbnails-with-lines:before {
  content: "\E158";
}
.glyphicons-show-lines:before {
  content: "\E159";
}
.glyphicons-playlist:before {
  content: "\E160";
}
.glyphicons-imac:before {
  content: "\E161";
}
.glyphicons-macbook:before {
  content: "\E162";
}
.glyphicons-ipad:before {
  content: "\E163";
}
.glyphicons-iphone:before {
  content: "\E164";
}
.glyphicons-iphone-transfer:before {
  content: "\E165";
}
.glyphicons-iphone-exchange:before {
  content: "\E166";
}
.glyphicons-ipod:before {
  content: "\E167";
}
.glyphicons-ipod-shuffle:before {
  content: "\E168";
}
.glyphicons-ear-plugs:before {
  content: "\E169";
}
.glyphicons-record:before {
  content: "\E170";
}
.glyphicons-step-backward:before {
  content: "\E171";
}
.glyphicons-fast-backward:before {
  content: "\E172";
}
.glyphicons-rewind:before {
  content: "\E173";
}
.glyphicons-play:before {
  content: "\E174";
}
.glyphicons-pause:before {
  content: "\E175";
}
.glyphicons-stop:before {
  content: "\E176";
}
.glyphicons-forward:before {
  content: "\E177";
}
.glyphicons-fast-forward:before {
  content: "\E178";
}
.glyphicons-step-forward:before {
  content: "\E179";
}
.glyphicons-eject:before {
  content: "\E180";
}
.glyphicons-facetime-video:before {
  content: "\E181";
}
.glyphicons-download-alt:before {
  content: "\E182";
}
.glyphicons-mute:before {
  content: "\E183";
}
.glyphicons-volume-down:before {
  content: "\E184";
}
.glyphicons-volume-up:before {
  content: "\E185";
}
.glyphicons-screenshot:before {
  content: "\E186";
}
.glyphicons-move:before {
  content: "\E187";
}
.glyphicons-more:before {
  content: "\E188";
}
.glyphicons-brightness-reduce:before {
  content: "\E189";
}
.glyphicons-brightness-increase:before {
  content: "\E190";
}
.glyphicons-circle-plus:before {
  content: "\E191";
}
.glyphicons-circle-minus:before {
  content: "\E192";
}
.glyphicons-circle-remove:before {
  content: "\E193";
}
.glyphicons-circle-ok:before {
  content: "\E194";
}
.glyphicons-circle-question-mark:before {
  content: "\E195";
}
.glyphicons-circle-info:before {
  content: "\E196";
}
.glyphicons-circle-exclamation-mark:before {
  content: "\E197";
}
.glyphicons-remove:before {
  content: "\E198";
}
.glyphicons-ok:before {
  content: "\E199";
}
.glyphicons-ban:before {
  content: "\E200";
}
.glyphicons-download:before {
  content: "\E201";
}
.glyphicons-upload:before {
  content: "\E202";
}
.glyphicons-shopping-cart:before {
  content: "\E203";
}
.glyphicons-lock:before {
  content: "\E204";
}
.glyphicons-unlock:before {
  content: "\E205";
}
.glyphicons-electricity:before {
  content: "\E206";
}
.glyphicons-ok-2:before {
  content: "\E207";
}
.glyphicons-remove-2:before {
  content: "\E208";
}
.glyphicons-cart-out:before {
  content: "\E209";
}
.glyphicons-cart-in:before {
  content: "\E210";
}
.glyphicons-left-arrow:before {
  content: "\E211";
}
.glyphicons-right-arrow:before {
  content: "\E212";
}
.glyphicons-down-arrow:before {
  content: "\E213";
}
.glyphicons-up-arrow:before {
  content: "\E214";
}
.glyphicons-resize-small:before {
  content: "\E215";
}
.glyphicons-resize-full:before {
  content: "\E216";
}
.glyphicons-circle-arrow-left:before {
  content: "\E217";
}
.glyphicons-circle-arrow-right:before {
  content: "\E218";
}
.glyphicons-circle-arrow-top:before {
  content: "\E219";
}
.glyphicons-circle-arrow-down:before {
  content: "\E220";
}
.glyphicons-play-button:before {
  content: "\E221";
}
.glyphicons-unshare:before {
  content: "\E222";
}
.glyphicons-share:before {
  content: "\E223";
}
.glyphicons-chevron-right:before {
  content: "\E224";
}
.glyphicons-chevron-left:before {
  content: "\E225";
}
.glyphicons-bluetooth:before {
  content: "\E226";
}
.glyphicons-euro:before {
  content: "\20AC";
}
.glyphicons-usd:before {
  content: "\E228";
}
.glyphicons-gbp:before {
  content: "\E229";
}
.glyphicons-retweet-2:before {
  content: "\E230";
}
.glyphicons-moon:before {
  content: "\E231";
}
.glyphicons-sun:before {
  content: "\2609";
}
.glyphicons-cloud:before {
  content: "\2601";
}
.glyphicons-direction:before {
  content: "\E234";
}
.glyphicons-brush:before {
  content: "\E235";
}
.glyphicons-pen:before {
  content: "\E236";
}
.glyphicons-zoom-in:before {
  content: "\E237";
}
.glyphicons-zoom-out:before {
  content: "\E238";
}
.glyphicons-pin:before {
  content: "\E239";
}
.glyphicons-albums:before {
  content: "\E240";
}
.glyphicons-rotation-lock:before {
  content: "\E241";
}
.glyphicons-flash:before {
  content: "\E242";
}
.glyphicons-google-maps:before {
  content: "\E243";
}
.glyphicons-anchor:before {
  content: "\2693";
}
.glyphicons-conversation:before {
  content: "\E245";
}
.glyphicons-chat:before {
  content: "\E246";
}
.glyphicons-male:before {
  content: "\E247";
}
.glyphicons-female:before {
  content: "\E248";
}
.glyphicons-asterisk:before {
  content: "\002A";
}
.glyphicons-divide:before {
  content: "\00F7";
}
.glyphicons-snorkel-diving:before {
  content: "\E251";
}
.glyphicons-scuba-diving:before {
  content: "\E252";
}
.glyphicons-oxygen-bottle:before {
  content: "\E253";
}
.glyphicons-fins:before {
  content: "\E254";
}
.glyphicons-fishes:before {
  content: "\E255";
}
.glyphicons-boat:before {
  content: "\E256";
}
.glyphicons-delete:before {
  content: "\E257";
}
.glyphicons-sheriffs-star:before {
  content: "\E258";
}
.glyphicons-qrcode:before {
  content: "\E259";
}
.glyphicons-barcode:before {
  content: "\E260";
}
.glyphicons-pool:before {
  content: "\E261";
}
.glyphicons-buoy:before {
  content: "\E262";
}
.glyphicons-spade:before {
  content: "\E263";
}
.glyphicons-bank:before {
  content: "\E264";
}
.glyphicons-vcard:before {
  content: "\E265";
}
.glyphicons-electrical-plug:before {
  content: "\E266";
}
.glyphicons-flag:before {
  content: "\E267";
}
.glyphicons-credit-card:before {
  content: "\E268";
}
.glyphicons-keyboard-wireless:before {
  content: "\E269";
}
.glyphicons-keyboard-wired:before {
  content: "\E270";
}
.glyphicons-shield:before {
  content: "\E271";
}
.glyphicons-ring:before {
  content: "\02DA";
}
.glyphicons-cake:before {
  content: "\E273";
}
.glyphicons-drink:before {
  content: "\E274";
}
.glyphicons-beer:before {
  content: "\E275";
}
.glyphicons-fast-food:before {
  content: "\E276";
}
.glyphicons-cutlery:before {
  content: "\E277";
}
.glyphicons-pizza:before {
  content: "\E278";
}
.glyphicons-birthday-cake:before {
  content: "\E279";
}
.glyphicons-tablet:before {
  content: "\E280";
}
.glyphicons-settings:before {
  content: "\E281";
}
.glyphicons-bullets:before {
  content: "\E282";
}
.glyphicons-cardio:before {
  content: "\E283";
}
.glyphicons-t-shirt:before {
  content: "\E284";
}
.glyphicons-pants:before {
  content: "\E285";
}
.glyphicons-sweater:before {
  content: "\E286";
}
.glyphicons-fabric:before {
  content: "\E287";
}
.glyphicons-leather:before {
  content: "\E288";
}
.glyphicons-scissors:before {
  content: "\E289";
}
.glyphicons-bomb:before {
  content: "\E290";
}
.glyphicons-skull:before {
  content: "\E291";
}
.glyphicons-celebration:before {
  content: "\E292";
}
.glyphicons-tea-kettle:before {
  content: "\E293";
}
.glyphicons-french-press:before {
  content: "\E294";
}
.glyphicons-coffee-cup:before {
  content: "\E295";
}
.glyphicons-pot:before {
  content: "\E296";
}
.glyphicons-grater:before {
  content: "\E297";
}
.glyphicons-kettle:before {
  content: "\E298";
}
.glyphicons-hospital:before {
  content: "\E299";
}
.glyphicons-hospital-h:before {
  content: "\E300";
}
.glyphicons-microphone:before {
  content: "\E301";
}
.glyphicons-webcam:before {
  content: "\E302";
}
.glyphicons-temple-christianity-church:before {
  content: "\E303";
}
.glyphicons-temple-islam:before {
  content: "\E304";
}
.glyphicons-temple-hindu:before {
  content: "\E305";
}
.glyphicons-temple-buddhist:before {
  content: "\E306";
}
.glyphicons-bicycle:before {
  content: "\E307";
}
.glyphicons-life-preserver:before {
  content: "\E308";
}
.glyphicons-share-alt:before {
  content: "\E309";
}
.glyphicons-comments:before {
  content: "\E310";
}
.glyphicons-flower:before {
  content: "\2698";
}
.glyphicons-baseball:before {
  content: "\26BE";
}
.glyphicons-rugby:before {
  content: "\E313";
}
.glyphicons-ax:before {
  content: "\E314";
}
.glyphicons-table-tennis:before {
  content: "\E315";
}
.glyphicons-bowling:before {
  content: "\E316";
}
.glyphicons-tree-conifer:before {
  content: "\E317";
}
.glyphicons-tree-deciduous:before {
  content: "\E318";
}
.glyphicons-more-items:before {
  content: "\E319";
}
.glyphicons-sort:before {
  content: "\E320";
}
.glyphicons-filter:before {
  content: "\E321";
}
.glyphicons-gamepad:before {
  content: "\E322";
}
.glyphicons-playing-dices:before {
  content: "\E323";
}
.glyphicons-calculator:before {
  content: "\E324";
}
.glyphicons-tie:before {
  content: "\E325";
}
.glyphicons-wallet:before {
  content: "\E326";
}
.glyphicons-piano:before {
  content: "\E327";
}
.glyphicons-sampler:before {
  content: "\E328";
}
.glyphicons-podium:before {
  content: "\E329";
}
.glyphicons-soccer-ball:before {
  content: "\E330";
}
.glyphicons-blog:before {
  content: "\E331";
}
.glyphicons-dashboard:before {
  content: "\E332";
}
.glyphicons-certificate:before {
  content: "\E333";
}
.glyphicons-bell:before {
  content: "\E334";
}
.glyphicons-candle:before {
  content: "\E335";
}
.glyphicons-pushpin:before {
  content: "\E336";
}
.glyphicons-iphone-shake:before {
  content: "\E337";
}
.glyphicons-pin-flag:before {
  content: "\E338";
}
.glyphicons-turtle:before {
  content: "\E339";
}
.glyphicons-rabbit:before {
  content: "\E340";
}
.glyphicons-globe:before {
  content: "\E341";
}
.glyphicons-briefcase:before {
  content: "\E342";
}
.glyphicons-hdd:before {
  content: "\E343";
}
.glyphicons-thumbs-up:before {
  content: "\E344";
}
.glyphicons-thumbs-down:before {
  content: "\E345";
}
.glyphicons-hand-right:before {
  content: "\E346";
}
.glyphicons-hand-left:before {
  content: "\E347";
}
.glyphicons-hand-up:before {
  content: "\E348";
}
.glyphicons-hand-down:before {
  content: "\E349";
}
.glyphicons-fullscreen:before {
  content: "\E350";
}
.glyphicons-shopping-bag:before {
  content: "\E351";
}
.glyphicons-book-open:before {
  content: "\E352";
}
.glyphicons-nameplate:before {
  content: "\E353";
}
.glyphicons-nameplate-alt:before {
  content: "\E354";
}
.glyphicons-vases:before {
  content: "\E355";
}
.glyphicons-bullhorn:before {
  content: "\E356";
}
.glyphicons-dumbbell:before {
  content: "\E357";
}
.glyphicons-suitcase:before {
  content: "\E358";
}
.glyphicons-file-import:before {
  content: "\E359";
}
.glyphicons-file-export:before {
  content: "\E360";
}
.glyphicons-bug:before {
  content: "\E361";
}
.glyphicons-crown:before {
  content: "\E362";
}
.glyphicons-smoking:before {
  content: "\E363";
}
.glyphicons-cloud-download:before {
  content: "\E364";
}
.glyphicons-cloud-upload:before {
  content: "\E365";
}
.glyphicons-restart:before {
  content: "\E366";
}
.glyphicons-security-camera:before {
  content: "\E367";
}
.glyphicons-expand:before {
  content: "\E368";
}
.glyphicons-collapse:before {
  content: "\E369";
}
.glyphicons-collapse-top:before {
  content: "\E370";
}
.glyphicons-globe-af:before {
  content: "\E371";
}
.glyphicons-global:before {
  content: "\E372";
}
.glyphicons-spray:before {
  content: "\E373";
}
.glyphicons-nails:before {
  content: "\E374";
}
.glyphicons-claw-hammer:before {
  content: "\E375";
}
.glyphicons-classic-hammer:before {
  content: "\E376";
}
.glyphicons-hand-saw:before {
  content: "\E377";
}
.glyphicons-riflescope:before {
  content: "\E378";
}
.glyphicons-electrical-socket-eu:before {
  content: "\E379";
}
.glyphicons-electrical-socket-us:before {
  content: "\E380";
}
.glyphicons-message-forward:before {
  content: "\E381";
}
.glyphicons-coat-hanger:before {
  content: "\E382";
}
.glyphicons-dress:before {
  content: "\E383";
}
.glyphicons-bathrobe:before {
  content: "\E384";
}
.glyphicons-shirt:before {
  content: "\E385";
}
.glyphicons-underwear:before {
  content: "\E386";
}
.glyphicons-log-in:before {
  content: "\E387";
}
.glyphicons-log-out:before {
  content: "\E388";
}
.glyphicons-exit:before {
  content: "\E389";
}
.glyphicons-new-window-alt:before {
  content: "\E390";
}
.glyphicons-video-sd:before {
  content: "\E391";
}
.glyphicons-video-hd:before {
  content: "\E392";
}
.glyphicons-subtitles:before {
  content: "\E393";
}
.glyphicons-sound-stereo:before {
  content: "\E394";
}
.glyphicons-sound-dolby:before {
  content: "\E395";
}
.glyphicons-sound-5-1:before {
  content: "\E396";
}
.glyphicons-sound-6-1:before {
  content: "\E397";
}
.glyphicons-sound-7-1:before {
  content: "\E398";
}
.glyphicons-copyright-mark:before {
  content: "\E399";
}
.glyphicons-registration-mark:before {
  content: "\E400";
}
.glyphicons-radar:before {
  content: "\E401";
}
.glyphicons-skateboard:before {
  content: "\E402";
}
.glyphicons-golf-course:before {
  content: "\E403";
}
.glyphicons-sorting:before {
  content: "\E404";
}
.glyphicons-sort-by-alphabet:before {
  content: "\E405";
}
.glyphicons-sort-by-alphabet-alt:before {
  content: "\E406";
}
.glyphicons-sort-by-order:before {
  content: "\E407";
}
.glyphicons-sort-by-order-alt:before {
  content: "\E408";
}
.glyphicons-sort-by-attributes:before {
  content: "\E409";
}
.glyphicons-sort-by-attributes-alt:before {
  content: "\E410";
}
.glyphicons-compressed:before {
  content: "\E411";
}
.glyphicons-package:before {
  content: "\E412";
}
.glyphicons-cloud-plus:before {
  content: "\E413";
}
.glyphicons-cloud-minus:before {
  content: "\E414";
}
.glyphicons-disk-save:before {
  content: "\E415";
}
.glyphicons-disk-open:before {
  content: "\E416";
}
.glyphicons-disk-saved:before {
  content: "\E417";
}
.glyphicons-disk-remove:before {
  content: "\E418";
}
.glyphicons-disk-import:before {
  content: "\E419";
}
.glyphicons-disk-export:before {
  content: "\E420";
}
.glyphicons-tower:before {
  content: "\E421";
}
.glyphicons-send:before {
  content: "\E422";
}
.glyphicons-git-branch:before {
  content: "\E423";
}
.glyphicons-git-create:before {
  content: "\E424";
}
.glyphicons-git-private:before {
  content: "\E425";
}
.glyphicons-git-delete:before {
  content: "\E426";
}
.glyphicons-git-merge:before {
  content: "\E427";
}
.glyphicons-git-pull-request:before {
  content: "\E428";
}
.glyphicons-git-compare:before {
  content: "\E429";
}
.glyphicons-git-commit:before {
  content: "\E430";
}
.glyphicons-construction-cone:before {
  content: "\E431";
}
.glyphicons-shoe-steps:before {
  content: "\E432";
}
.glyphicons-plus:before {
  content: "\002B";
}
.glyphicons-minus:before {
  content: "\2212";
}
.glyphicons-redo:before {
  content: "\E435";
}
.glyphicons-undo:before {
  content: "\E436";
}
.glyphicons-golf:before {
  content: "\E437";
}
.glyphicons-hockey:before {
  content: "\E438";
}
.glyphicons-pipe:before {
  content: "\E439";
}
.glyphicons-wrench:before {
  content: "\E440";
}
.glyphicons-folder-closed:before {
  content: "\E441";
}
.glyphicons-phone-alt:before {
  content: "\E442";
}
.glyphicons-earphone:before {
  content: "\E443";
}
.glyphicons-floppy-disk:before {
  content: "\E444";
}
.glyphicons-floppy-saved:before {
  content: "\E445";
}
.glyphicons-floppy-remove:before {
  content: "\E446";
}
.glyphicons-floppy-save:before {
  content: "\E447";
}
.glyphicons-floppy-open:before {
  content: "\E448";
}
.glyphicons-translate:before {
  content: "\E449";
}
.glyphicons-fax:before {
  content: "\E450";
}
.glyphicons-factory:before {
  content: "\E451";
}
.glyphicons-shop-window:before {
  content: "\E452";
}
.glyphicons-shop:before {
  content: "\E453";
}
.glyphicons-kiosk:before {
  content: "\E454";
}
.glyphicons-kiosk-wheels:before {
  content: "\E455";
}
.glyphicons-kiosk-light:before {
  content: "\E456";
}
.glyphicons-kiosk-food:before {
  content: "\E457";
}
.glyphicons-transfer:before {
  content: "\E458";
}
.glyphicons-money:before {
  content: "\E459";
}
.glyphicons-header:before {
  content: "\E460";
}
.glyphicons-blacksmith:before {
  content: "\E461";
}
.glyphicons-saw-blade:before {
  content: "\E462";
}
.glyphicons-basketball:before {
  content: "\E463";
}
.glyphicons-server:before {
  content: "\E464";
}
.glyphicons-server-plus:before {
  content: "\E465";
}
.glyphicons-server-minus:before {
  content: "\E466";
}
.glyphicons-server-ban:before {
  content: "\E467";
}
.glyphicons-server-flag:before {
  content: "\E468";
}
.glyphicons-server-lock:before {
  content: "\E469";
}
.glyphicons-server-new:before {
  content: "\E470";
}
.glyphicons-charging-station:before {
  content: "\F471";
}
.glyphicons-gas-station:before {
  content: "\E472";
}
.glyphicons-target:before {
  content: "\E473";
}
.glyphicons-bed-alt:before {
  content: "\E474";
}
.glyphicons-mosquito-net:before {
  content: "\E475";
}
.glyphicons-dining-set:before {
  content: "\E476";
}
.glyphicons-plate-of-food:before {
  content: "\E477";
}
.glyphicons-hygiene-kit:before {
  content: "\E478";
}
.glyphicons-blackboard:before {
  content: "\E479";
}
.glyphicons-marriage:before {
  content: "\E480";
}
.glyphicons-bucket:before {
  content: "\E481";
}
.glyphicons-none-color-swatch:before {
  content: "\E482";
}
.glyphicons-bring-forward:before {
  content: "\E483";
}
.glyphicons-bring-to-front:before {
  content: "\E484";
}
.glyphicons-send-backward:before {
  content: "\E485";
}
.glyphicons-send-to-back:before {
  content: "\E486";
}
.glyphicons-fit-frame-to-image:before {
  content: "\E487";
}
.glyphicons-fit-image-to-frame:before {
  content: "\E488";
}
.glyphicons-multiple-displays:before {
  content: "\E489";
}
.glyphicons-handshake:before {
  content: "\E490";
}
.glyphicons-child:before {
  content: "\E491";
}
.glyphicons-baby-formula:before {
  content: "\E492";
}
.glyphicons-medicine:before {
  content: "\E493";
}
.glyphicons-atv-vehicle:before {
  content: "\E494";
}
.glyphicons-motorcycle:before {
  content: "\E495";
}
.glyphicons-bed:before {
  content: "\E496";
}
.glyphicons-tent:before {
  content: "\26FA";
}
.glyphicons-glasses:before {
  content: "\E498";
}
.glyphicons-sunglasses:before {
  content: "\E499";
}
.glyphicons-family:before {
  content: "\E500";
}
.glyphicons-education:before {
  content: "\E501";
}
.glyphicons-shoes:before {
  content: "\E502";
}
.glyphicons-map:before {
  content: "\E503";
}
.glyphicons-cd:before {
  content: "\E504";
}
.glyphicons-alert:before {
  content: "\E505";
}
.glyphicons-piggy-bank:before {
  content: "\E506";
}
.glyphicons-star-half:before {
  content: "\E507";
}
.glyphicons-cluster:before {
  content: "\E508";
}
.glyphicons-flowchart:before {
  content: "\E509";
}
.glyphicons-commodities:before {
  content: "\E510";
}
.glyphicons-duplicate:before {
  content: "\E511";
}
.glyphicons-copy:before {
  content: "\E512";
}
.glyphicons-paste:before {
  content: "\E513";
}
.glyphicons-bath-bathtub:before {
  content: "\E514";
}
.glyphicons-bath-shower:before {
  content: "\E515";
}
.glyphicons-shower:before {
  content: "\1F6BF";
}
.glyphicons-menu-hamburger:before {
  content: "\E517";
}
.glyphicons-option-vertical:before {
  content: "\E518";
}
.glyphicons-option-horizontal:before {
  content: "\E519";
}
.glyphicons-currency-conversion:before {
  content: "\E520";
}
.glyphicons-user-ban:before {
  content: "\E521";
}
.glyphicons-user-lock:before {
  content: "\E522";
}
.glyphicons-user-flag:before {
  content: "\E523";
}
.glyphicons-user-asterisk:before {
  content: "\E524";
}
.glyphicons-user-alert:before {
  content: "\E525";
}
.glyphicons-user-key:before {
  content: "\E526";
}
.glyphicons-user-conversation:before {
  content: "\E527";
}
.glyphicons-database:before {
  content: "\E528";
}
.glyphicons-database-search:before {
  content: "\E529";
}
.glyphicons-list-alt:before {
  content: "\E530";
}
.glyphicons-hazard-sign:before {
  content: "\E531";
}
.glyphicons-hazard:before {
  content: "\E532";
}
.glyphicons-stop-sign:before {
  content: "\E533";
}
.glyphicons-lab:before {
  content: "\E534";
}
.glyphicons-lab-alt:before {
  content: "\E535";
}
.glyphicons-ice-cream:before {
  content: "\E536";
}
.glyphicons-ice-lolly:before {
  content: "\E537";
}
.glyphicons-ice-lolly-tasted:before {
  content: "\E538";
}
.glyphicons-invoice:before {
  content: "\E539";
}
.glyphicons-cart-tick:before {
  content: "\E540";
}
.glyphicons-hourglass:before {
  content: "\231B";
}
.glyphicons-cat:before {
  content: "\1F408";
}
.glyphicons-lamp:before {
  content: "\E543";
}
.glyphicons-scale-classic:before {
  content: "\E544";
}
.glyphicons-eye-plus:before {
  content: "\E545";
}
.glyphicons-eye-minus:before {
  content: "\E546";
}
.glyphicons-quote:before {
  content: "\E547";
}
.glyphicons-bitcoin:before {
  content: "\E548";
}
.glyphicons-yen:before {
  content: "\00A5";
}
.glyphicons-ruble:before {
  content: "\20BD";
}
.glyphicons-erase:before {
  content: "\E551";
}
.glyphicons-podcast:before {
  content: "\E552";
}
.glyphicons-firework:before {
  content: "\E553";
}
.glyphicons-scale:before {
  content: "\E554";
}
.glyphicons-king:before {
  content: "\E555";
}
.glyphicons-queen:before {
  content: "\E556";
}
.glyphicons-pawn:before {
  content: "\E557";
}
.glyphicons-bishop:before {
  content: "\E558";
}
.glyphicons-knight:before {
  content: "\E559";
}
.glyphicons-mic-mute:before {
  content: "\E560";
}
.glyphicons-voicemail:before {
  content: "\E561";
}
.glyphicons-paragraph:before {
  content: "\00B6";
}
.glyphicons-person-walking:before {
  content: "\E563";
}
.glyphicons-person-wheelchair:before {
  content: "\E564";
}
.glyphicons-underground:before {
  content: "\E565";
}
.glyphicons-car-hov:before {
  content: "\E566";
}
.glyphicons-car-rental:before {
  content: "\E567";
}
.glyphicons-transport:before {
  content: "\E568";
}
.glyphicons-taxi:before {
  content: "\1F695";
}
.glyphicons-ice-cream-no:before {
  content: "\E570";
}
.glyphicons-uk-rat-u:before {
  content: "\E571";
}
.glyphicons-uk-rat-pg:before {
  content: "\E572";
}
.glyphicons-uk-rat-12a:before {
  content: "\E573";
}
.glyphicons-uk-rat-12:before {
  content: "\E574";
}
.glyphicons-uk-rat-15:before {
  content: "\E575";
}
.glyphicons-uk-rat-18:before {
  content: "\E576";
}
.glyphicons-uk-rat-r18:before {
  content: "\E577";
}
.glyphicons-tv:before {
  content: "\E578";
}
.glyphicons-sms:before {
  content: "\E579";
}
.glyphicons-mms:before {
  content: "\E580";
}
.glyphicons-us-rat-g:before {
  content: "\E581";
}
.glyphicons-us-rat-pg:before {
  content: "\E582";
}
.glyphicons-us-rat-pg-13:before {
  content: "\E583";
}
.glyphicons-us-rat-restricted:before {
  content: "\E584";
}
.glyphicons-us-rat-no-one-17:before {
  content: "\E585";
}
.glyphicons-equalizer:before {
  content: "\E586";
}
.glyphicons-speakers:before {
  content: "\E587";
}
.glyphicons-remote-control:before {
  content: "\E588";
}
.glyphicons-remote-control-tv:before {
  content: "\E589";
}
.glyphicons-shredder:before {
  content: "\E590";
}
.glyphicons-folder-heart:before {
  content: "\E591";
}
.glyphicons-person-running:before {
  content: "\E592";
}
.glyphicons-person:before {
  content: "\E593";
}
.glyphicons-voice:before {
  content: "\E594";
}
.glyphicons-stethoscope:before {
  content: "\E595";
}
.glyphicons-hotspot:before {
  content: "\E596";
}
.glyphicons-activity:before {
  content: "\E597";
}
.glyphicons-watch:before {
  content: "\231A";
}
.glyphicons-scissors-alt:before {
  content: "\E599";
}
.glyphicons-car-wheel:before {
  content: "\E600";
}
.glyphicons-chevron-up:before {
  content: "\E601";
}
.glyphicons-chevron-down:before {
  content: "\E602";
}
.glyphicons-superscript:before {
  content: "\E603";
}
.glyphicons-subscript:before {
  content: "\E604";
}
.glyphicons-text-size:before {
  content: "\E605";
}
.glyphicons-text-color:before {
  content: "\E606";
}
.glyphicons-text-background:before {
  content: "\E607";
}
.glyphicons-modal-window:before {
  content: "\E608";
}
.glyphicons-newspaper:before {
  content: "\1F4F0";
}
.glyphicons-tractor:before {
  content: "\1F69C";
}
/* 
* 
* THIS IS A SMALL BONUS FOR ALL CURIOUS PEOPLE :) 
* Just add class .animated and .pulse, .rotateIn, .bounce, .swing or .tada to you HTML element with icons. You may find other great css animations here: http://coveloping.com/tools/css-animation-generator 
* 
*/
.animated {
  -webkit-animation-duration: 1s;
  animation-duration: 1s;
  -webkit-animation-fill-mode: both;
  animation-fill-mode: both;
  -webkit-animation-timing-function: ease-in-out;
  animation-timing-function: ease-in-out;
  animation-iteration-count: infinite;
  -webkit-animation-iteration-count: infinite;
}
@-webkit-keyframes pulse {
  0% {
    -webkit-transform: scale(1);
  }
  50% {
    -webkit-transform: scale(1.1);
  }
  100% {
    -webkit-transform: scale(1);
  }
}
@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}
.pulse {
  -webkit-animation-name: pulse;
  animation-name: pulse;
}
@-webkit-keyframes rotateIn {
  0% {
    -webkit-transform-origin: center center;
    -webkit-transform: rotate(-200deg);
    opacity: 0;
  }
  100% {
    -webkit-transform-origin: center center;
    -webkit-transform: rotate(0);
    opacity: 1;
  }
}
@keyframes rotateIn {
  0% {
    transform-origin: center center;
    transform: rotate(-200deg);
    opacity: 0;
  }
  100% {
    transform-origin: center center;
    transform: rotate(0);
    opacity: 1;
  }
}
.rotateIn {
  -webkit-animation-name: rotateIn;
  animation-name: rotateIn;
}
@-webkit-keyframes bounce {
  0%,
  20%,
  50%,
  80%,
  100% {
    -webkit-transform: translateY(0);
  }
  40% {
    -webkit-transform: translateY(-30px);
  }
  60% {
    -webkit-transform: translateY(-15px);
  }
}
@keyframes bounce {
  0%,
  20%,
  50%,
  80%,
  100% {
    transform: translateY(0);
  }
  40% {
    transform: translateY(-30px);
  }
  60% {
    transform: translateY(-15px);
  }
}
.bounce {
  -webkit-animation-name: bounce;
  animation-name: bounce;
}
@-webkit-keyframes swing {
  20%,
  40%,
  60%,
  80%,
  100% {
    -webkit-transform-origin: top center;
  }
  20% {
    -webkit-transform: rotate(15deg);
  }
  40% {
    -webkit-transform: rotate(-10deg);
  }
  60% {
    -webkit-transform: rotate(5deg);
  }
  80% {
    -webkit-transform: rotate(-5deg);
  }
  100% {
    -webkit-transform: rotate(0deg);
  }
}
@keyframes swing {
  20% {
    transform: rotate(15deg);
  }
  40% {
    transform: rotate(-10deg);
  }
  60% {
    transform: rotate(5deg);
  }
  80% {
    transform: rotate(-5deg);
  }
  100% {
    transform: rotate(0deg);
  }
}
.swing {
  -webkit-transform-origin: top center;
  transform-origin: top center;
  -webkit-animation-name: swing;
  animation-name: swing;
}
@-webkit-keyframes tada {
  0% {
    -webkit-transform: scale(1);
  }
  10%,
  20% {
    -webkit-transform: scale(0.9) rotate(-3deg);
  }
  30%,
  50%,
  70%,
  90% {
    -webkit-transform: scale(1.1) rotate(3deg);
  }
  40%,
  60%,
  80% {
    -webkit-transform: scale(1.1) rotate(-3deg);
  }
  100% {
    -webkit-transform: scale(1) rotate(0);
  }
}
@keyframes tada {
  0% {
    transform: scale(1);
  }
  10%,
  20% {
    transform: scale(0.9) rotate(-3deg);
  }
  30%,
  50%,
  70%,
  90% {
    transform: scale(1.1) rotate(3deg);
  }
  40%,
  60%,
  80% {
    transform: scale(1.1) rotate(-3deg);
  }
  100% {
    transform: scale(1) rotate(0);
  }
}
.tada {
  -webkit-animation-name: tada;
  animation-name: tada;
}
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
  background-color: #fafafa;
  background-size: cover;
  background-attachment: fixed;
}
.shadow {
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
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
h1,
h2,
h3,
h4 {
  font-weight: normal;
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
  background-color: rgba(64, 120, 192, 0.05);
}
a:active {
  background-color: rgba(64, 120, 192, 0.1);
}
ul {
  padding: 0px 20px;
  margin: 0px;
  margin-bottom: 1rem;
}
.admin_box {
  box-shadow: 0px 1px 5px rgba(0, 0, 0, 0.1);
  margin: 5px auto;
  background-color: white;
  border-radius: 2px;
  max-width: 600px;
  width: 100%;
  margin-bottom: 20px;
  background-color: #fff;
  border-radius: 3px;
}
.admin_box_padding {
  padding: 10px;
}
.admin_box-wide {
  box-shadow: none;
  max-width: none;
  width: 100%;
  padding: 0px;
  margin-top: 0px;
  background-color: rgba(0, 0, 0, 0);
}
.btn {
  display: inline-block;
  padding: 5px 30px;
  font-size: 1.1rem;
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
.primarybtncontainer {
  text-align: right;
}
.btn-primary {
  background: linear-gradient(#42a1ec, #0070c9);
  color: white;
  border-color: #1E90FF;
}
.btn-primary:after {
  content: " >";
}
.btn-primary:hover {
  background-color: #147bcd;
  background: linear-gradient(#51a9ee, #147bcd);
  border-color: #1482d0;
  outline: none;
}
.btn-primary:active {
  background-color: #0067b9;
  background: linear-gradient(#3d94d9, #0067b9);
  border-color: #006dbc;
  outline: none;
}
.btn-delete,
.btn-delete:hover,
.btn-delete:active {
  border: 1px solid red;
  background: linear-gradient(#dd2e4f, #dd2e4f);
}
.btn-delete:hover,
.btn-delete:active {
  background: linear-gradient(red, red);
}
.form {
  padding: 10px;
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
.form_label:first-of-type {
  margin-top: 0px;
}
.form_label:last-child {
  margin-bottom: 0px;
}
.form_label-required input {
  border-width: 2px;
}
.form_label_text {
  line-height: 2em;
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
.inputzone {
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.075);
  background-color: #fafafa;
  outline: none;
  border: 1px solid #ddd;
  padding: 8px 6px;
  display: inline-block;
  border-radius: 3px;
  font-size: 0.9rem;
  line-height: 1.4rem;
  color: #333;
  vertical-align: middle;
  width: 100%;
}
.input {
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.075);
  background-color: #fafafa;
  outline: none;
  border: 1px solid #ddd;
  padding: 8px 6px;
  display: inline-block;
  border-radius: 3px;
  font-size: 0.9rem;
  line-height: 1.4rem;
  color: #333;
  vertical-align: middle;
  width: 100%;
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
  padding-right: 24px;
  background: #fafafa url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAUCAMAAACzvE1FAAAADFBMVEUzMzMzMzMzMzMzMzMKAG/3AAAAA3RSTlMAf4C/aSLHAAAAPElEQVR42q3NMQ4AIAgEQTn//2cLdRKppSGzBYwzVXvznNWs8C58CiussPJj8h6NwgorrKRdTvuV9v16Afn0AYFOB7aYAAAAAElFTkSuQmCC") no-repeat right 8px center;
  background-size: 8px 10px;
  width: auto;
  max-width: 100%;
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
  min-width: 100%;
  margin: 0 auto;
}
.admin_table_orderheader {
  text-decoration: none;
}
.admin_table td {
  padding: 2px;
  border: 1px solid #f1f1f1;
}
.admin_table-list td {
  border-left: none;
  border-bottom: none;
  border-top: none;
}
.admin_table_row:nth-child(odd) {
  background-color: #fafafa;
}
.admin_table_row:hover {
  background-color: rgba(64, 120, 192, 0.1);
  cursor: pointer;
}
.admin_table_row td {
  padding: 2px 5px;
}
.admin_table-list tr:last-child td {
  border-bottom: 1px solid #f1f1f1;
}
.admin_table-list tr td:first-child,
.admin_table-list tr th:first-child {
  border-left: 1px solid #f1f1f1;
}
.admin_table th {
  padding: 5px;
  vertical-align: bottom;
  font-weight: normal;
  border: 1px solid #f1f1f1;
  border-left: none;
}
.admin_table th a {
  display: block;
}
.admin_table_loading {
  opacity: .4;
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
  pointer-events: none;
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
  pointer-events: auto;
  color: #444;
  display: inline-block;
  margin: 5px 10px;
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
  border-radius: 3px;
  animation-name: example;
  animation-duration: 300ms;
  animation-timing-function: ease-in;
  background-color: #FFD800;
  display: inline-flex;
  align-items: center;
  cursor: default;
}
.flash_message_content {
  border-radius: 3px;
  padding: 5px 20px;
  flex-grow: 2;
}
.flash_message_close {
  flex-shrink: 0;
  padding: 5px;
  color: #4078c0;
  opacity: .2;
  cursor: pointer;
}
.flash_message:hover .flash_message_close {
  opacity: 1;
}
td.pagination {
  text-align: center;
  background-color: white;
  padding: 10px 5px;
}
.pagination_page,
.pagination_page_current {
  color: #4078c0;
  padding: 2px 10px;
  margin: 0px;
  margin-right: 1px;
  font-size: 16px;
  line-height: 1.4em;
  display: inline-block;
  text-decoration: none;
  border-radius: 3px;
}
.pagination_page:hover {
  background-color: rgba(64, 120, 192, 0.1);
}
.pagination_page_current,
.pagination_page_current:hover {
  background-color: #4078c0;
  color: white;
}
.pagination_page_current {
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
  border-radius: 3px;
  margin: 2px;
  display: inline-block;
  text-align: center;
  vertical-align: middle;
  position: relative;
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
}
.admin_images_image:hover {
  background-color: rgba(64, 120, 192, 0.1);
}
.admin_images_image:hover .admin_images_image_delete {
  background-color: #eee;
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
  color: white;
  background: red !important;
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
  margin-bottom: 10px;
  padding: 3px;
  border-radius: 3px;
  padding: 10px;
  background-color: #fff;
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
  cursor: pointer;
}
.admin_images_fileinput:hover {
  background-color: rgba(64, 120, 192, 0.1);
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
.admin_timestamp {
  display: inline-flex;
  align-items: center;
}
select.admin_timestamp_date {
  width: 150px;
  flex-grow: 0;
}
select.admin_timestamp_hour {
  width: 60px;
  flex-grow: 0;
  margin: 0px 5px;
}
select.admin_timestamp_minute {
  width: 60px;
  flex-grow: 0;
  margin-left: 5px;
}
.admin-action-order {
  background: #fafafa !important;
  cursor: move;
}
.ordered,
.ordered:hover {
  background-color: rgba(64, 120, 192, 0.1);
  border-radius: 3px;
}
.ordered:after,
.ordered-desc:after {
  font-weight: bold;
}
.ordered:after {
  content: " ↓";
}
.ordered-desc:after {
  content: " ↑";
}
.view_name {
  font-size: 1.2rem;
  margin-left: 0px;
  padding: 10px 10px 5px 10px;
  border-top: 1px solid rgba(64, 120, 192, 0.1);
  border-top: 2px solid #fafafa;
}
.view_name:first-child {
  border-top: none;
}
.view_content {
  margin-bottom: 10px;
  padding: 5px 10px;
  border-radius: 3px;
  word-wrap: break-word;
  color: #888;
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
  border: 4px solid #4078c0;
  border-top: 4px solid rgba(64, 120, 192, 0.1);
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
  flex-wrap: wrap;
  margin: 10px 10px;
  font-size: 1.1rem;
  align-items: center;
}
.admin_navigation_breadcrumb {
  text-decoration: none;
  border-radius: 3px;
  display: flex;
  align-items: center;
  line-height: 30px;
  padding-left: 5px;
  padding: 0px 5px 0px 5px;
}
.admin_navigation_breadcrumb_image {
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
  width: 30px;
  height: 30px;
  margin: 3px 10px 3px 0px;
  display: inline-block;
  border-radius: 300px;
  background-color: white;
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
}
.admin_navigation_breadcrumb_image-logo {
  background-size: contain;
}
.admin_navigation_breadcrumb_divider {
  font-size: 1.4rem;
  vertical-align: center;
  display: inline-flex;
  color: #4078c0;
  margin: 0px 0px 0px 10px;
  font-weight: 100;
}
.admin_navigation_breadcrumb_divider:after {
  content: ">";
}
.admin_navigation_tabs {
  display: flex;
  justify-content: center;
  vertical-align: bottom;
  flex-wrap: wrap;
  margin: 10px 5px 20px 5px;
}
.admin_navigation_tab {
  white-space: nowrap;
  overflow: hidden;
  margin: 0px;
  display: flex;
  border-bottom: none;
  font-size: 1rem;
  text-decoration: none;
  padding: 3px 10px;
  display: inline-block;
  background-color: white;
  border-top: 1px solid #4078c0;
  border-bottom: 1px solid #4078c0;
  border-right: 1px solid #4078c0;
}
.admin_navigation_tab:first-of-type {
  border-top-left-radius: 5px;
  border-bottom-left-radius: 5px;
  border-left: 1px solid #4078c0;
}
.admin_navigation_tab:last-of-type {
  border-top-right-radius: 5px;
  border-bottom-right-radius: 5px;
}
.admin_navigation_tab-selected,
.admin_navigation_tab-selected:hover {
  background-color: #4078c0;
  color: white;
  box-shadow: none;
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
  top: 19px;
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
  z-index: 2;
  flex-flow: column;
  background-color: white;
}
.btn-more:hover {
  color: #444;
  border-bottom-right-radius: 0px;
  display: inline-block;
}
.btn-more:hover .btn-more_content,
.btn-more:active .btn-more_content {
  display: flex;
}
.btn-more_content_item {
  display: block;
  background: none;
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
.admin_filter_layout_date_value {
  width: 1px;
  height: 1px;
  opacity: 0;
  margin: 0px;
  position: absolute;
}
.admin_filter_layout_date_content {
  display: flex;
}
.admin_filter_layout_date_divider {
  padding: 0px 5px;
}
td.admin_list_message {
  padding: 20px 5px;
  text-align: center;
  font-size: 1.3rem;
  color: #999;
}
@media (max-width: 600px) {
  .admin_box {
    box-shadow: none;
    margin-bottom: 0px;
    padding-bottom: 20px;
  }
  .admin_header {
    position: relative !important;
  }
}
.admin_preview {
  padding: 5px;
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
  border-radius: 5px;
  text-decoration: none;
  color: #444;
  display: inline-block;
  margin: 5px 0px;
  font-size: 1.1rem;
  display: flex;
  align-items: flex-start;
}
.admin_preview_image {
  flex-grow: 0;
  flex-shrink: 0;
  width: 50px;
  height: 50px;
  margin-right: 10px;
  border-radius: 300px;
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
  background-color: #eee;
}
.admin_preview_right {
  flex-grow: 2;
  flex-shrink: 2;
}
.admin_preview_description {
  font-size: .9rem;
  line-height: 1.2em;
  color: #888;
}
.admin_item_relation {
  display: flex;
  align-items: center;
}
.admin_item_relation_change {
  margin: 10px 0px;
  flex-grow: 0;
  flex-shrink: 0;
}
.admin_item_relation_change_btn {
  display: inline-block;
  width: 40px;
  text-align: center;
  margin: 0px 20px;
  font-size: 1.2rem;
  line-height: 1.2em;
}
.admin_item_relation_picker {
  display: flex;
  flex-direction: column;
  flex-grow: 2;
}
.admin_item_relation_picker_suggestions {
  position: relative;
}
.admin_item_relation_picker_suggestions_content {
  position: absolute;
  top: 0px;
  z-index: 20;
  background-color: white;
  margin-bottom: 10px;
  border-bottom-right-radius: 5px;
  border-bottom-left-radius: 5px;
  box-shadow: 0px 1px 10px 0px rgba(0, 0, 0, 0.1);
  width: 400px;
}
.admin_item_relation_picker_suggestion {
  box-shadow: none;
  margin: 0px;
  border-bottom: 1px solid #eee;
  cursor: pointer;
  border-radius: 0px;
}
.admin_item_relation_picker_suggestion:last-child {
  border-bottom-left-radius: 3px;
  border-bottom-right-radius: 3px;
  border-bottom: none;
}
.admin_item_relation_picker_suggestion-selected {
  background-color: rgba(64, 120, 192, 0.05);
}
.admin_header {
  background: white;
  padding-bottom: 0px;
  position: relative;
  z-index: 2;
  line-height: 1.6em;
  flex-grow: 0;
  flex-shrink: 0;
  position: sticky;
  top: 0px;
  box-shadow: 0px 2px 3px rgba(0, 0, 0, 0.05);
}
.admin_header_top_item {
  border-radius: 3px;
  padding: 2px 5px;
  margin: 2px 2px;
  font-size: 1.1rem;
}
.admin_header a {
  text-decoration: none;
}
.admin_header_top_space {
  flex-grow: 2;
}
.admin_header_top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  padding: 5px 10px;
  background-color: #fafafa;
  background: linear-gradient(#fafafa, #fff);
}
.admin_header_name {
  font-size: 1.4rem;
  font-weight: 500;
  line-height: 1.2em;
}
.admin_header_resources {
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
  padding: 3px 10px;
  margin: 0px;
  font-size: .9rem;
  border-bottom: 2px solid none;
  flex-shrink: 0;
  border-radius: 3px;
}
.admin_header_resource-active,
.admin_header_resource-active:hover {
  background-color: #4078c0;
  color: white;
}
.admin_header_sitename {
  font-size: .9rem;
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
        this.el.addEventListener("click", function (e) {
            if (e.altKey) {
                var ids = window.prompt("IDs of images", _this.hiddenInput.value);
                _this.hiddenInput.value = ids;
                e.preventDefault();
                return false;
            }
        });
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
        this.prefilterField = el.getAttribute("data-prefilter-field");
        this.prefilterValue = el.getAttribute("data-prefilter-value");
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
                _this.tbody.classList.remove("admin_table_loading");
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
                window.location.href = url;
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
        ret.PrefilterField = this.prefilterField;
        ret.PrefilterValue = this.prefilterValue;
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
            input.addEventListener("input", this.inputListener.bind(this));
        }
        this.inputPeriodicListener();
    };
    List.prototype.inputListener = function (e) {
        if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
            return;
        }
        this.tbody.classList.add("admin_table_loading");
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
function bindRelations() {
    var elements = document.querySelectorAll(".admin_item_relation");
    Array.prototype.forEach.call(elements, function (el, i) {
        new RelationPicker(el);
    });
}
var RelationPicker = (function () {
    function RelationPicker(el) {
        var _this = this;
        this.selectedClass = "admin_item_relation_picker_suggestion-selected";
        this.input = el.getElementsByTagName("input")[0];
        this.previewContainer = el.querySelector(".admin_item_relation_preview");
        this.relationName = el.getAttribute("data-relation");
        this.progress = el.querySelector("progress");
        this.changeSection = el.querySelector(".admin_item_relation_change");
        this.changeButton = el.querySelector(".admin_item_relation_change_btn");
        this.changeButton.addEventListener("click", function () {
            _this.showSearch();
            _this.pickerInput.focus();
        });
        this.suggestionsEl = el.querySelector(".admin_item_relation_picker_suggestions_content");
        this.suggestions = [];
        this.picker = el.querySelector(".admin_item_relation_picker");
        this.pickerInput = this.picker.querySelector("input");
        this.pickerInput.addEventListener("input", function () {
            _this.getSuggestions(_this.pickerInput.value);
        });
        this.pickerInput.addEventListener("blur", function () {
            _this.suggestionsEl.classList.add("hidden");
        });
        this.pickerInput.addEventListener("focus", function () {
            _this.suggestionsEl.classList.remove("hidden");
        });
        this.pickerInput.addEventListener("keydown", this.suggestionInput.bind(this));
        this.getData();
    }
    RelationPicker.prototype.getData = function () {
        var _this = this;
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/preview/" + this.relationName + "/" + this.input.value, true);
        request.addEventListener("load", function () {
            _this.progress.classList.add("hidden");
            if (request.status == 200) {
                _this.showPreview(JSON.parse(request.response));
            }
            else {
                _this.showSearch();
            }
        });
        request.send();
    };
    RelationPicker.prototype.showPreview = function (data) {
        this.previewContainer.textContent = "";
        this.input.value = data.ID;
        var el = this.createPreview(data);
        this.previewContainer.appendChild(el);
        this.previewContainer.classList.remove("hidden");
        this.changeSection.classList.remove("hidden");
        this.picker.classList.add("hidden");
    };
    RelationPicker.prototype.showSearch = function () {
        this.previewContainer.classList.add("hidden");
        this.changeSection.classList.add("hidden");
        this.picker.classList.remove("hidden");
        this.suggestions = [];
        this.suggestionsEl.innerText = "";
        this.pickerInput.value = "";
    };
    RelationPicker.prototype.getSuggestions = function (q) {
        var _this = this;
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/search/" + this.relationName + "?q=" + encodeURIComponent(q), true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                if (q != _this.pickerInput.value) {
                    return;
                }
                var data = JSON.parse(request.response);
                _this.suggestions = data;
                _this.suggestionsEl.innerText = "";
                for (var i = 0; i < data.length; i++) {
                    var item = data[i];
                    var el = _this.createPreview(item);
                    el.classList.add("admin_item_relation_picker_suggestion");
                    el.setAttribute("data-position", i + "");
                    el.addEventListener("mousedown", _this.suggestionClick.bind(_this));
                    el.addEventListener("mouseenter", _this.suggestionSelect.bind(_this));
                    _this.suggestionsEl.appendChild(el);
                }
            }
            else {
                console.log("Error while searching");
            }
        });
        request.send();
    };
    RelationPicker.prototype.suggestionClick = function () {
        var selected = this.getSelected();
        if (selected >= 0) {
            this.showPreview(this.suggestions[selected]);
        }
    };
    RelationPicker.prototype.suggestionSelect = function (e) {
        var target = e.currentTarget;
        var position = parseInt(target.getAttribute("data-position"));
        this.select(position);
    };
    RelationPicker.prototype.getSelected = function () {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        return parseInt(selected.getAttribute("data-position"));
    };
    RelationPicker.prototype.unselect = function () {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        selected.classList.remove(this.selectedClass);
        return parseInt(selected.getAttribute("data-position"));
    };
    RelationPicker.prototype.select = function (i) {
        this.unselect();
        this.suggestionsEl.querySelectorAll(".admin_preview")[i].classList.add(this.selectedClass);
    };
    RelationPicker.prototype.suggestionInput = function (e) {
        switch (e.keyCode) {
            case 13:
                this.suggestionClick();
                e.preventDefault();
                return true;
            case 38:
                var i = this.getSelected();
                if (i < 1) {
                    i = this.suggestions.length - 1;
                }
                else {
                    i = i - 1;
                }
                this.select(i);
                e.preventDefault();
                return false;
            case 40:
                var i = this.getSelected();
                if (i >= 0) {
                    i += 1;
                    i = i % this.suggestions.length;
                }
                else {
                    i = 0;
                }
                this.select(i);
                e.preventDefault();
                return false;
        }
    };
    RelationPicker.prototype.createPreview = function (data) {
        var ret = document.createElement("div");
        ret.classList.add("admin_preview");
        var image = document.createElement("div");
        image.classList.add("admin_preview_image");
        image.setAttribute("style", "background-image: url('" + data.Image + "');");
        var right = document.createElement("div");
        right.classList.add("admin_preview_right");
        var name = document.createElement("div");
        name.classList.add("admin_preview_name");
        name.textContent = data.Name;
        var description = document.createElement("description");
        description.classList.add("admin_preview_description");
        description.textContent = data.Description;
        ret.appendChild(image);
        right.appendChild(name);
        right.appendChild(description);
        ret.appendChild(right);
        return ret;
    };
    return RelationPicker;
}());
function bindRelationsOLD() {
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
            el.innerText = "-";
            el.classList.remove("admin_item_view_place");
            return;
        }
        var position = { lat: parseFloat(coords[0]), lng: parseFloat(coords[1]) };
        var zoom = 18;
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
        this.dirty = false;
        el.addEventListener("submit", function () {
            _this.dirty = false;
        });
        var els = el.querySelectorAll(".form_watcher");
        for (var i = 0; i < els.length; i++) {
            var input = els[i];
            input.addEventListener("input", function () {
                _this.dirty = true;
            });
            input.addEventListener("change", function () {
                _this.dirty = true;
            });
        }
        window.addEventListener("beforeunload", function (e) {
            if (_this.dirty) {
                var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
                e.returnValue = confirmationMessage;
                return confirmationMessage;
            }
        });
    }
    return Form;
}());
function bindFilter() {
    var els = document.querySelectorAll(".admin_filter_layout_date");
    for (var i = 0; i < els.length; i++) {
        new FilterDate(els[i]);
    }
}
var FilterDate = (function () {
    function FilterDate(el) {
        this.hidden = el.querySelector(".admin_table_filter_item");
        this.from = el.querySelector(".admin_filter_layout_date_from");
        this.to = el.querySelector(".admin_filter_layout_date_to");
        this.from.addEventListener("input", this.changed.bind(this));
        this.to.addEventListener("input", this.changed.bind(this));
    }
    FilterDate.prototype.changed = function () {
        var val = "";
        if (this.from.value && this.to.value) {
            val = this.from.value + " - " + this.to.value;
        }
        this.hidden.value = val;
        var event = new Event('change');
        this.hidden.dispatchEvent(event);
    };
    return FilterDate;
}());
document.addEventListener("DOMContentLoaded", function () {
    bindMarkdowns();
    bindTimestamps();
    bindRelations();
    bindImagePickers();
    bindLists();
    bindForm();
    bindImageViews();
    bindFlashMessages();
    bindFilter();
});
function bindFlashMessages() {
    var messages = document.querySelectorAll(".flash_message");
    for (var i = 0; i < messages.length; i++) {
        var message = messages[i];
        message.addEventListener("click", function (e) {
            var target = e.target;
            if (target.classList.contains("flash_message_close")) {
                var current = e.currentTarget;
                current.classList.add("hidden");
            }
        });
    }
}
`


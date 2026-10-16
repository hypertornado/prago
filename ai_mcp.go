package prago

import (
	"context"
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"

	"github.com/hypertornado/prago/pragoai"
)

const mcpAPIURL = "mcp"

const (
	mcpDefaultLimit int64 = 20
	mcpMaxLimit     int64 = 100
)

type mcpRequestContextKey struct{}

// MCPTool je nástroj, který aplikace vystavuje přes MCP endpoint
// /admin/api/mcp. Uvidí ho jen uživatelé s oprávněním nastaveným přes
// Permission.
type MCPTool struct {
	app        *App
	tool       *pragoai.MCPTool
	permission Permission
}

// NewMCPTool přidá nástroj do MCP serveru aplikace. Vstup T musí být
// struktura, její JSON schéma se odvodí z tagů json a ai (případně
// prago-name). Handler běží v kontextu přihlášeného uživatele.
func (app *App) NewMCPTool[T any](name, description string, handler func(request *Request, input T) (any, error)) *MCPTool {
	for _, tool := range app.mcpTools {
		if tool.tool.Name() == name {
			panic(fmt.Sprintf("MCP tool %s already exists", name))
		}
	}

	tool, err := pragoai.NewMCPTool(name, description, func(ctx context.Context, input T) (any, error) {
		return handler(ctx.Value(mcpRequestContextKey{}).(*Request), input)
	})
	Must(err)

	ret := &MCPTool{
		app:  app,
		tool: tool,
	}
	app.mcpTools = append(app.mcpTools, ret)
	return ret
}

func (tool *MCPTool) Permission(permission Permission) *MCPTool {
	Must(tool.app.validatePermission(permission))
	tool.permission = permission
	return tool
}

// ReadOnly označí nástroj, že nic nemění.
func (tool *MCPTool) ReadOnly() *MCPTool {
	tool.tool.ReadOnly()
	return tool
}

func (app *App) initAIMCP() {
	handler := func(request *Request) {
		_, isAPI := request.getLoginSessionID()
		if !isAPI || request.UserID() == 0 {
			request.Response().Header().Set("WWW-Authenticate", `Bearer realm="prago"`)
			renderAPIMessage(request, 401, "Unauthorized")
			return
		}

		server := pragoai.NewMCPServer(app.codeName, app.version)
		server.Title = app.name(request.Locale())
		server.Instructions = fmt.Sprintf("Nástroje pro práci s daty administrace aplikace %s. Data jsou uložená v resources: list_resources vrátí jejich seznam a pole, list_items a get_item čtou položky, search hledá fulltextem napříč administrací.", server.Title)
		for _, tool := range app.mcpTools {
			if request.Authorize(tool.permission) {
				server.AddTool(tool.tool)
			}
		}

		ctx := context.WithValue(request.Request().Context(), mcpRequestContextKey{}, request)
		server.ServeHTTP(request.Response(), request.Request().WithContext(ctx))
		request.Written = true
	}

	app.NewAPI(mcpAPIURL).Method("POST").Permission(everybodyPermission).Handler(handler)
	app.NewAPI(mcpAPIURL).Method("GET").Permission(everybodyPermission).Handler(handler)

	app.initAIMCPTools()

	app.ActionUI("_aimcp", func(request *Request) template.HTML {
		endpoint := app.BaseURL() + app.getAdminURL("api/"+mcpAPIURL)

		table := app.NewTable()
		table.Row(table.Cell("MCP server").Header().Colspan(2))
		table.Row(table.Cell("Endpoint"), table.Cell(endpoint))
		table.Row(
			table.Cell("Autentizace"),
			table.Cell(fmt.Sprintf("API klíč v HTTP hlavičce %s, nebo Authorization: Bearer", apiHTTPHeader)).URL(app.getAdminURL("_api")),
		)
		table.Row(
			table.Cell("Claude Code"),
			table.Cell(fmt.Sprintf(`claude mcp add --transport http %s %s --header "%s: PAK_..."`, app.codeName, endpoint, apiHTTPHeader)).Pre(),
		)

		table.Row(table.Cell("Nástroje").Header().Colspan(2))
		for _, tool := range app.mcpTools {
			if request.Authorize(tool.permission) {
				table.Row(table.Cell(tool.tool.Name()), table.Cell(tool.tool.Description()))
			}
		}

		return table.ExecuteHTML()
	}).Permission(loggedPermission).Name(Unlocalized("MCP server")).Board(aiBoard).Icon("glyphicons-basic-849-computer-network.svg")
}

func (app *App) afterInitMCPTools() {
	for _, tool := range app.mcpTools {
		if tool.permission == "" {
			panic(fmt.Sprintf("permission for MCP tool %s should not be empty", tool.tool.Name()))
		}
	}
}

type mcpSearchInput struct {
	Query  string `json:"query" ai:"Hledaný text"`
	Offset int64  `json:"offset,omitempty" ai:"Kolik výsledků přeskočit (stránkování)"`
	Limit  int64  `json:"limit,omitempty" ai:"Maximální počet výsledků, výchozí 20, nejvýše 100"`
}

type mcpSearchItem struct {
	Name    string `json:"name"`
	Prename string `json:"prename,omitempty"`
	URL     string `json:"url"`
}

type mcpResourceInfo struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Fields []*mcpFieldInfo `json:"fields"`
}

type mcpFieldInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	Description     string `json:"description,omitempty"`
	RelatedResource string `json:"related_resource,omitempty"`
}

type mcpListItemsInput struct {
	Resource  string       `json:"resource" ai:"ID resource z nástroje list_resources"`
	Offset    int64        `json:"offset,omitempty" ai:"Kolik položek přeskočit (stránkování)"`
	Limit     int64        `json:"limit,omitempty" ai:"Maximální počet položek, výchozí 20, nejvýše 100"`
	Order     string       `json:"order,omitempty" ai:"ID pole, podle kterého se řadí; bez vyplnění se použije výchozí řazení resource"`
	OrderDesc bool         `json:"order_desc,omitempty" ai:"Řadit sestupně"`
	Filters   []*mcpFilter `json:"filters,omitempty" ai:"Filtry na přesnou shodu hodnoty pole, musí platit všechny"`
}

type mcpFilter struct {
	Field string `json:"field" ai:"ID pole"`
	Value string `json:"value" ai:"Hodnota, které se pole musí rovnat"`
}

type mcpGetItemInput struct {
	Resource string `json:"resource" ai:"ID resource z nástroje list_resources"`
	ID       int64  `json:"id" ai:"ID položky"`
}

func (app *App) initAIMCPTools() {

	app.NewMCPTool("search", "Fulltextové vyhledávání napříč administrací (menu, resources a vlastní vyhledávání aplikace).", func(request *Request, input mcpSearchInput) (any, error) {
		items, hits, err := app.search(input.Query, max(input.Offset, 0), mcpLimit(input.Limit), request)
		if err != nil {
			return nil, err
		}
		ret := []*mcpSearchItem{}
		for _, item := range items {
			ret = append(ret, &mcpSearchItem{
				Name:    item.Name,
				Prename: item.Prename,
				URL:     app.mcpAbsoluteURL(item.URL),
			})
		}
		return map[string]any{
			"hits":  hits,
			"items": ret,
		}, nil
	}).Permission(loggedPermission).ReadOnly()

	app.NewMCPTool("list_resources", "Seznam resources (datových tabulek), ke kterým má uživatel přístup, včetně jejich polí.", func(request *Request, input struct{}) (any, error) {
		ret := []*mcpResourceInfo{}
		for _, resource := range app.resources {
			if !app.mcpCanViewResource(request, resource) {
				continue
			}
			info := &mcpResourceInfo{
				ID:     resource.id,
				Name:   resource.pluralName(request.Locale()),
				Fields: []*mcpFieldInfo{},
			}
			for _, field := range resource.fields {
				if !field.authorizeView(request) {
					continue
				}
				fieldInfo := &mcpFieldInfo{
					ID:   field.id,
					Name: field.name(request.Locale()),
					Type: field.typeID(),
				}
				if field.description != nil {
					fieldInfo.Description = field.description(request.Locale())
				}
				if field.relatedResource != nil {
					fieldInfo.RelatedResource = field.relatedResource.id
				}
				info.Fields = append(info.Fields, fieldInfo)
			}
			ret = append(ret, info)
		}
		return map[string]any{"resources": ret}, nil
	}).Permission(loggedPermission).ReadOnly()

	app.NewMCPTool("list_items", "Seznam položek resource se stránkováním, řazením a filtry. Vrací celkový počet položek odpovídajících filtrům.", func(request *Request, input mcpListItemsInput) (any, error) {
		resource, err := app.mcpGetResource(request, input.Resource)
		if err != nil {
			return nil, err
		}

		ctx := request.Request().Context()
		newQuery := func() (*listQuery, error) {
			query := resource.query(ctx)
			for _, filter := range input.Filters {
				field, err := resource.mcpGetField(request, filter.Field)
				if err != nil {
					return nil, err
				}
				value, err := mcpFilterValue(field, filter.Value)
				if err != nil {
					return nil, err
				}
				query.Is(field.id, value)
			}
			return query, nil
		}

		countQuery, err := newQuery()
		if err != nil {
			return nil, err
		}
		total, err := countQuery.count()
		if err != nil {
			return nil, err
		}

		query, err := newQuery()
		if err != nil {
			return nil, err
		}
		if input.Order != "" {
			field, err := resource.mcpGetField(request, input.Order)
			if err != nil {
				return nil, err
			}
			query.addOrder(field.id, input.OrderDesc)
		} else if resource.orderByColumn != "" {
			query.addOrder(resource.orderByColumn, resource.orderDesc)
		}
		query.Offset(max(input.Offset, 0)).Limit(mcpLimit(input.Limit))

		items, err := query.list()
		if err != nil {
			return nil, err
		}

		ret := []map[string]any{}
		itemsVal := reflect.ValueOf(items)
		for i := 0; i < itemsVal.Len(); i++ {
			ret = append(ret, resource.mcpItemData(itemsVal.Index(i).Interface(), request))
		}

		return map[string]any{
			"total": total,
			"items": ret,
		}, nil
	}).Permission(loggedPermission).ReadOnly()

	app.NewMCPTool("get_item", "Detail jedné položky resource podle ID.", func(request *Request, input mcpGetItemInput) (any, error) {
		resource, err := app.mcpGetResource(request, input.Resource)
		if err != nil {
			return nil, err
		}
		item := resource.query(request.Request().Context()).ID(input.ID)
		if item == nil {
			return nil, fmt.Errorf("item %d not found in resource %s", input.ID, resource.id)
		}
		return resource.mcpItemData(item, request), nil
	}).Permission(loggedPermission).ReadOnly()
}

func mcpLimit(limit int64) int64 {
	if limit <= 0 {
		return mcpDefaultLimit
	}
	return min(limit, mcpMaxLimit)
}

func (app *App) mcpAbsoluteURL(url string) string {
	if strings.HasPrefix(url, "/") {
		return app.BaseURL() + url
	}
	return url
}

// mcpCanViewResource skrývá sessions, protože obsahují přihlašovací a API
// klíče uživatelů, které nemají jak skončit v kontextu modelu.
func (app *App) mcpCanViewResource(request *Request, resource *Resource) bool {
	if resource == app.sessionsResource {
		return false
	}
	return request.Authorize(resource.canView)
}

func (app *App) mcpGetResource(request *Request, id string) (*Resource, error) {
	resource := app.getResourceByID(id)
	if resource == nil || !app.mcpCanViewResource(request, resource) {
		return nil, fmt.Errorf("unknown resource %q, use list_resources", id)
	}
	return resource, nil
}

func (resource *Resource) mcpGetField(request *Request, id string) (*Field, error) {
	field := resource.Field(id)
	if field == nil || !field.authorizeView(request) {
		return nil, fmt.Errorf("unknown field %q in resource %s, use list_resources", id, resource.id)
	}
	return field, nil
}

func mcpFilterValue(field *Field, value string) (any, error) {
	switch field.typ.Kind() {
	case reflect.Bool:
		return strconv.ParseBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(value, 10, 64)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(value, 64)
	default:
		return value, nil
	}
}

func (resource *Resource) mcpItemData(item any, request *Request) map[string]any {
	previewer := resource.previewer(request, item)
	itemVal := reflect.ValueOf(item).Elem()

	fields := map[string]any{}
	for _, field := range resource.fields {
		if !field.authorizeView(request) {
			continue
		}
		fields[field.id] = itemVal.FieldByName(field.fieldClassName).Interface()
	}

	return map[string]any{
		"id":     previewer.ID(),
		"name":   previewer.Name(),
		"url":    resource.app.mcpAbsoluteURL(previewer.URL("")),
		"fields": fields,
	}
}

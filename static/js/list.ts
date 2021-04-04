function bindLists() {
  var els = document.getElementsByClassName("admin_list");
  for (var i = 0; i < els.length; i++) {
    new List(<HTMLDivElement>els[i], <HTMLButtonElement>document.querySelector(".admin_tablesettings_buttons"));
  }
}

class List {
  adminPrefix: string;
  typeName: string;

  tbody: HTMLElement;
  el: HTMLDivElement;
  exportButton: HTMLAnchorElement;
  changed: boolean;
  changedTimestamp: number;
  
  defaultOrderColumn: string;
  orderColumn: string;
  defaultOrderDesc: boolean;
  orderDesc: boolean;
  page: number;

  defaultVisibleColumnsStr: string;

  progress: HTMLProgressElement;

  settingsRow: HTMLTableRowElement;
  settingsRowColumn: HTMLTableElement;
  settingsEl: HTMLDivElement;
  settingsCheckbox: HTMLInputElement;

  itemsPerPage: number;
  paginationSelect: HTMLSelectElement;

  statsCheckbox: HTMLInputElement;
  statsCheckboxSelectCount: HTMLSelectElement;
  statsContainer: HTMLDivElement;

  settingsButton: HTMLButtonElement;
  settingsPopup: ContentPopup;

  constructor(el: HTMLDivElement, openbutton: HTMLButtonElement) {
    this.el = el;

    this.settingsRow = document.querySelector(".admin_list_settingsrow");
    this.settingsRowColumn = document.querySelector(".admin_list_settingsrow_column");
    this.settingsEl = document.querySelector(".admin_tablesettings");

    //this.settingsCheckbox = this.el.querySelector(".admin_list_showmore");
    //this.settingsCheckbox.addEventListener("change", this.settingsCheckboxChange.bind(this));
    //this.settingsCheckboxChange();

    this.settingsPopup = new ContentPopup("Možnosti", this.settingsEl);
    this.settingsButton = this.el.querySelector(".admin_list_settings");
    this.settingsButton.addEventListener("click", () => {
      this.settingsPopup.show();
    })

    this.exportButton = document.querySelector(".admin_exportbutton");

    let urlParams = new URLSearchParams(window.location.search);

    this.page = parseInt(urlParams.get("_page"));
    if (!this.page) {
      this.page = 1;
    }

    this.typeName = el.getAttribute("data-type");
    if (!this.typeName) {
      return;
    }

    this.progress = <HTMLProgressElement>el.querySelector(".admin_table_progress");

    this.tbody = <HTMLElement>el.querySelector("tbody");
    this.tbody.textContent = "";

    this.bindFilter(urlParams);

    this.adminPrefix = document.body.getAttribute("data-admin-prefix");

    this.defaultOrderColumn = el.getAttribute("data-order-column");
    if (el.getAttribute("data-order-desc") == "true") {
      this.defaultOrderDesc = true;
    } else {
      this.defaultOrderDesc = false;
    }
    this.orderColumn = this.defaultOrderColumn;
    this.orderDesc = this.defaultOrderDesc;

    if (urlParams.get("_order")) {
      this.orderColumn = urlParams.get("_order");
    }
    if (urlParams.get("_desc") == "true") {
      this.orderDesc = true
    }
    if (urlParams.get("_desc") == "false") {
      this.orderDesc = false
    }

    this.defaultVisibleColumnsStr = el.getAttribute("data-visible-columns");
    var visibleColumnsStr = this.defaultVisibleColumnsStr;
    if (urlParams.get("_columns")) {
      visibleColumnsStr = urlParams.get("_columns");
    }

    let visibleColumnsArr = visibleColumnsStr.split(",");
    let visibleColumnsMap: any = {};
    for (var i = 0; i < visibleColumnsArr.length; i++) {
      visibleColumnsMap[visibleColumnsArr[i]] = true;
    }

    this.itemsPerPage = parseInt(el.getAttribute("data-items-per-page"));
    this.paginationSelect = <HTMLSelectElement>document.querySelector(".admin_tablesettings_pages");
    this.paginationSelect.addEventListener("change", this.load.bind(this));

    this.statsCheckbox = document.querySelector(".admin_tablesettings_stats");
    this.statsCheckbox.addEventListener("change", () => {
      this.filterChanged();
    });

    this.statsCheckboxSelectCount = document.querySelector(".admin_tablesettings_stats_limit");
    this.statsCheckboxSelectCount.addEventListener("change", () => {
      this.filterChanged();
    });

    this.statsContainer = document.querySelector(".admin_tablesettings_stats_container");

    if (this.hasMultipleActions()) {
      this.bindMultipleActions();
    }


    this.bindOptions(visibleColumnsMap);
    this.bindOrder();
  }

  checkboxesAr: NodeListOf<HTMLInputElement>;

  hasMultipleActions(): Boolean {
    if (this.el.classList.contains("admin_list-hasmultipleactions")) {
      return true;
    }
    return false;
  }

  bindMultipleActions() {
    var actions = this.el.querySelectorAll(".admin_list_multiple_action");
    for (var i = 0; i < actions.length; i++) {
      actions[i].addEventListener("click", this.multipleActionSelected.bind(this));
    }
  }

  multipleActionSelected(e: any) {
    var target: HTMLDivElement = e.target;
    var actionName = target.getAttribute("name");

    switch (actionName) {
      case "cancel":
        this.multipleUncheckAll();
        break;
      case "delete":
        var ids = this.multipleGetIDs();
        new Confirm(`Opravdu chcete smazat ${ids.length} položek?`, () => {
          var loader = new LoadingPopup();
          var params: any = {};
          params["action"] = "delete";
          params["ids"] = ids.join(",");
          var url = this.adminPrefix + "/" + this.typeName + "/api/multipleaction" + encodeParams(params);
          fetch(url, {
            method: "POST",
          }).then((e) => {
            loader.done();
            if (e.status != 200) {
              new Alert("Error while doing multipleaction delete");
              return;
            }
            this.load();
          });
        }, Function(), ButtonStyle.Delete);
        break;
      default:
        console.log("other");
    }

  }

  bindMultipleActionCheckboxes() {
    this.checkboxesAr = document.querySelectorAll(".admin_table_cell-multiple_checkbox");
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      checkbox.addEventListener("change", this.multipleCheckboxChanged.bind(this));
    }
    this.multipleCheckboxChanged();
  }

  multipleGetIDs(): Array<String> {
    var ret: Array<String> = [];
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      if (checkbox.checked) {
        ret.push(checkbox.getAttribute("data-id"));
      }
    }
    return ret;
  }

  multipleCheckboxChanged() {
    var checkedCount = 0;
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      if (checkbox.checked) {
        checkedCount++;
      }
    }

    var multipleActionsPanel: HTMLDivElement = this.el.querySelector(".admin_list_multiple_actions");
    if (checkedCount > 0) {
      multipleActionsPanel.classList.add("admin_list_multiple_actions-visible");
    } else {
      multipleActionsPanel.classList.remove("admin_list_multiple_actions-visible");
    }
    this.el.querySelector(".admin_list_multiple_actions_description").textContent = `Vybráno ${checkedCount} položek`
  }

  multipleUncheckAll() {
    for (var i = 0; i < this.checkboxesAr.length; i++) {
      var checkbox = <HTMLInputElement>this.checkboxesAr[i];
      checkbox.checked = false;
    }
    this.multipleCheckboxChanged()
  }

  settingsCheckboxChange() {
    if (this.settingsCheckbox.checked) {
      this.settingsRow.classList.add("admin_list_settingsrow-visible");
    } else {
      this.settingsRow.classList.remove("admin_list_settingsrow-visible");
    }
  }

  load() {
    this.progress.classList.remove("admin_table_progress-inactive");
    var request = new XMLHttpRequest();
    var params: any = {};
    if (this.page > 1) {
      params["_page"] = this.page;
    }
    if (this.orderColumn != this.defaultOrderColumn) {
      params["_order"] = this.orderColumn;
    }
    if (this.orderDesc != this.defaultOrderDesc) {

      params["_desc"] = this.orderDesc + "";
    }
    var columns = this.getSelectedColumnsStr();
    if (columns != this.defaultVisibleColumnsStr) {
      params["_columns"] = columns;
    }

    let filterData = this.getFilterData();
    for (var k in filterData) {
      params[k] = filterData[k];
    }
    this.colorActiveFilterItems();

    let selectedPages = parseInt(this.paginationSelect.value);
    if (selectedPages != this.itemsPerPage) {
      params["_pagesize"] = selectedPages;
    }

    var encoded = encodeParams(params);
    window.history.replaceState(null, null, document.location.pathname + encoded);

    if (this.statsCheckbox.checked) {
      params["_stats"] = "true";
      params["_statslimit"] = this.statsCheckboxSelectCount.value;
    }

    params["_format"] = "xlsx";
    if (this.exportButton) {
      this.exportButton.setAttribute("href", this.adminPrefix + "/" + this.typeName + "/api/list" + encodeParams(params));
    }

    params["_format"] = "json";
    encoded = encodeParams(params);

    request.open("GET", this.adminPrefix + "/" + this.typeName + "/api/list" + encoded, true);
    request.addEventListener("load", () => {
      this.tbody.innerHTML = "";
      if (request.status == 200) {
        var response = JSON.parse(request.response);

        this.tbody.innerHTML = response.Content;
        var countStr = response.CountStr;

        this.el.querySelector(".admin_table_count").textContent = countStr;
        this.statsContainer.innerHTML = response.StatsStr;
        bindOrder();
        this.bindPagination();
        this.bindClick();
        if (this.hasMultipleActions()) {
          this.bindMultipleActionCheckboxes();
        }
        this.tbody.classList.remove("admin_table_loading");
      } else {
        console.error("error while loading list");
      }
      this.progress.classList.add("admin_table_progress-inactive");
    });
    request.send(JSON.stringify({}));
  }

  bindOptions(visibleColumnsMap: any) {
    var columns: NodeListOf<HTMLInputElement> = document.querySelectorAll(".admin_tablesettings_column");
    for (var i = 0; i < columns.length; i++) {
      let columnName = columns[i].getAttribute("data-column-name");
      if (visibleColumnsMap[columnName]) {
        columns[i].checked = true;
      }
      columns[i].addEventListener("change", () => {
        this.changedOptions();
      });
    }
    this.changedOptions();
  }

  changedOptions() {
    var columns: any = this.getSelectedColumnsMap();

    var headers: NodeListOf<HTMLDivElement> = document.querySelectorAll(".admin_list_orderitem");
    for (var i = 0; i < headers.length; i++) {
      var name = headers[i].getAttribute("data-name");
      if (columns[name]) {
        headers[i].classList.remove("hidden");
      } else {
        headers[i].classList.add("hidden");
      }
    }

    var filters: NodeListOf<HTMLDivElement> = document.querySelectorAll(".admin_list_filteritem");
    for (var i = 0; i < filters.length; i++) {
      var name = filters[i].getAttribute("data-name");
      if (columns[name]) {
        filters[i].classList.remove("hidden");
      } else {
        filters[i].classList.add("hidden");
      }
    }

    this.settingsRowColumn.setAttribute("colspan", Object.keys(columns).length + "");

    this.load();
  }

  colorActiveFilterItems() {
    let itemsToColor = this.getFilterData();
    var filterItems: NodeListOf<HTMLDivElement> = this.el.querySelectorAll(".admin_list_filteritem");
    for (var i = 0; i < filterItems.length; i++) {
      var item = filterItems[i];
      let name = item.getAttribute("data-name");
      if (itemsToColor[name]) {
        item.classList.add("admin_list_filteritem-colored");
      } else {
        item.classList.remove("admin_list_filteritem-colored");
      }
    }
  }

  paginationChange(e:any) {
    var el = <HTMLAnchorElement>e.target;
    var page = parseInt(el.getAttribute("data-page"));
    this.page = page;
    this.load();
    e.preventDefault();
    return false;
  }

  bindPagination() {
    var paginationEl = this.el.querySelector(".pagination");
    var totalPages = parseInt(paginationEl.getAttribute("data-total"));
    var selectedPage = parseInt(paginationEl.getAttribute("data-selected"));
    for (var i = 1; i <= totalPages; i++) {
      var pEl = document.createElement("a");
      pEl.setAttribute("href", "#");
      pEl.textContent = i+"";
      if (i == selectedPage) {
        pEl.classList.add("pagination_page_current");
      } else {
        pEl.classList.add("pagination_page");
        pEl.setAttribute("data-page", i+"");
        pEl.addEventListener("click", this.paginationChange.bind(this));
      }
      paginationEl.appendChild(pEl);
    }

  }

  bindClick() {
    var rows = this.el.querySelectorAll(".admin_table_row");
    for (var i = 0; i < rows.length; i++) {
      var row = <HTMLTableRowElement>rows[i];
      var id = row.getAttribute("data-id");
      row.addEventListener("click", (e) => {
        if ((<HTMLDivElement>e.target).classList.contains("admin_table_cell-multiple_checkbox")) {
          return false;
        }
        var target = <HTMLElement>e.target;
        if (target.classList.contains("preventredirect")) {
          return;
        }
        var el = <HTMLDivElement>e.currentTarget;
        var url = el.getAttribute("data-url");

        if (e.shiftKey || e.metaKey || e.ctrlKey) {
          var openedWindow = window.open(url, "newwindow" + (new Date()));
          openedWindow.focus();
          return;
        }
        window.location.href = url;
      });

      var buttons = row.querySelector(".admin_list_buttons");
      buttons.addEventListener("click", (e) => {
        var url = (<HTMLDivElement>e.target).getAttribute("href");
        if (url != "") {
          window.location.href = url;
          e.preventDefault();
          e.stopPropagation();
          return false;
        }
      })
    }
  }

  bindOrder() {
    this.renderOrder();
    var headers = this.el.querySelectorAll(".admin_list_orderitem-canorder");
    for (var i = 0; i < headers.length; i++) {
      var header = <HTMLAnchorElement>headers[i];
      header.addEventListener("click", (e) => {
        var el = <HTMLAnchorElement>e.target;
        var name = el.getAttribute("data-name");
        if (name == this.orderColumn) {
          if (this.orderDesc) {
            this.orderDesc = false;
          } else {
            this.orderDesc = true;
          }
        } else {
          this.orderColumn = name;
          this.orderDesc = false;
        }
        this.renderOrder();
        this.load();
        e.preventDefault();
        return false;
      });
    }
  }

  renderOrder() {
    var headers = this.el.querySelectorAll(".admin_list_orderitem-canorder");
    for (var i = 0; i < headers.length; i++) {
      var header = <HTMLAnchorElement>headers[i];
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
  }

  getSelectedColumnsStr(): string {
    var ret = [];
    var checked: NodeListOf<HTMLInputElement> = document.querySelectorAll(".admin_tablesettings_column:checked");
    for (var i = 0; i < checked.length; i++) {
      ret.push(checked[i].getAttribute("data-column-name"));
    }
    return ret.join(",");
  }

  getSelectedColumnsMap(): any {
    var columns: any = {};
    var checked: NodeListOf<HTMLInputElement> = document.querySelectorAll(".admin_tablesettings_column:checked");
    for (var i = 0; i < checked.length; i++) {
      columns[checked[i].getAttribute("data-column-name")] = true;
    }
    return columns;
  }

  getFilterData(): any {
    var ret: any = {};
    var items = this.el.querySelectorAll(".admin_table_filter_item");
    for (var i = 0; i < items.length; i++) {
      var item = <HTMLInputElement>items[i];
      var typ = item.getAttribute("data-typ");
      var layout = item.getAttribute("data-filter-layout");
      if (item.classList.contains("admin_table_filter_item-relations")) {
        ret[typ] = item.querySelector("input").value;
      } else {
        var val = item.value.trim();
        if (val) {
          ret[typ] = val;
        }
      }
    }
    return ret;
  }

  bindFilter(params: any) {
    var filterFields = this.el.querySelectorAll(".admin_list_filteritem");
    for (var i = 0; i < filterFields.length; i++) {
      var field: HTMLDivElement = <HTMLDivElement>filterFields[i];
      var fieldName = field.getAttribute("data-name");
      var fieldLayout = field.getAttribute("data-filter-layout");
      var fieldInput = field.querySelector("input");
      var fieldSelect = field.querySelector("select");
      var fieldValue = params.get(fieldName);

      if (fieldValue) {
        if (fieldInput) {
          fieldInput.value = fieldValue;
        }
        if (fieldSelect) {
          fieldSelect.value = fieldValue;
        }
      }

      if (fieldInput) {
        fieldInput.addEventListener("input", this.inputListener.bind(this));
        fieldInput.addEventListener("change", this.inputListener.bind(this));
      }

      if (fieldSelect) {
        fieldSelect.addEventListener("input", this.inputListener.bind(this));
        fieldSelect.addEventListener("change", this.inputListener.bind(this));
      }

      if (fieldLayout == "filter_layout_relation") {
        this.bindFilterRelation(field, fieldValue);
      }

      if (fieldLayout == "filter_layout_date") {
        this.bindFilterDate(field, fieldValue);
      }

    }
    this.inputPeriodicListener();
  }

  inputListener(e: any) {
    if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
      return;
    }
    this.filterChanged();
  }

  filterChanged() {
    this.colorActiveFilterItems();
    this.tbody.classList.add("admin_table_loading");
    this.page = 1;
    this.changed = true;
    this.changedTimestamp = Date.now();
    this.progress.classList.remove("admin_table_progress-inactive");
  }

  bindFilterRelation(el: HTMLDivElement, value: any) {
    new ListFilterRelations(el, value, this);
  }

  bindFilterDate(el: HTMLDivElement, value: any) {
    new ListFilterDate(el, value);
  }


  inputPeriodicListener() {
    setInterval(() =>{
      if (this.changed == true && Date.now() - this.changedTimestamp > 500) {
        this.changed = false;
        this.load();
      }
    }, 200);
  }
}
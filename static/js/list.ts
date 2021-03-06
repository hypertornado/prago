class List {
  settings: ListSettings;

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

  itemsPerPage: number;
  paginationSelect: HTMLSelectElement;

  statsCheckbox: HTMLInputElement;
  statsCheckboxSelectCount: HTMLSelectElement;
  statsContainer: HTMLDivElement;

  multiple: ListMultiple;

  constructor(el: HTMLDivElement) {
    this.el = el;

    var dateFilterInputs = el.querySelectorAll<HTMLInputElement>(
      ".admin_filter_date_input"
    );
    dateFilterInputs.forEach((el) => {
      new DatePicker(el);
    });

    this.settings = new ListSettings(this);
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

    this.progress = <HTMLProgressElement>(
      el.querySelector(".admin_table_progress")
    );

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
      this.orderDesc = true;
    }
    if (urlParams.get("_desc") == "false") {
      this.orderDesc = false;
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
    this.paginationSelect = <HTMLSelectElement>(
      document.querySelector(".admin_tablesettings_pages")
    );
    this.paginationSelect.addEventListener("change", this.load.bind(this));

    this.statsCheckbox = document.querySelector(".admin_tablesettings_stats");
    this.statsCheckbox.addEventListener("change", () => {
      this.filterChanged();
    });

    this.statsCheckboxSelectCount = document.querySelector(
      ".admin_tablesettings_stats_limit"
    );
    this.statsCheckboxSelectCount.addEventListener("change", () => {
      this.filterChanged();
    });

    this.statsContainer = document.querySelector(
      ".admin_tablesettings_stats_container"
    );

    this.multiple = new ListMultiple(this);

    this.settings.bindOptions(visibleColumnsMap);
    this.bindOrder();
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
    var columns = this.settings.getSelectedColumnsStr();
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
    window.history.replaceState(
      null,
      null,
      document.location.pathname + encoded
    );

    if (this.statsCheckbox.checked) {
      params["_stats"] = "true";
      params["_statslimit"] = this.statsCheckboxSelectCount.value;
    }

    params["_format"] = "xlsx";
    if (this.exportButton) {
      this.exportButton.setAttribute(
        "href",
        this.adminPrefix +
          "/" +
          this.typeName +
          "/api/list" +
          encodeParams(params)
      );
    }

    params["_format"] = "json";
    encoded = encodeParams(params);

    request.open(
      "GET",
      this.adminPrefix + "/" + this.typeName + "/api/list" + encoded,
      true
    );
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
        if (this.multiple.hasMultipleActions()) {
          this.multiple.bindMultipleActionCheckboxes();
        }
        this.tbody.classList.remove("admin_table_loading");
      } else {
        console.error("error while loading list");
      }
      this.progress.classList.add("admin_table_progress-inactive");
    });
    request.send(JSON.stringify({}));
  }

  colorActiveFilterItems() {
    let itemsToColor = this.getFilterData();
    var filterItems: NodeListOf<HTMLDivElement> = this.el.querySelectorAll(
      ".admin_list_filteritem"
    );
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

  paginationChange(e: any) {
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
      pEl.textContent = i + "";
      if (i == selectedPage) {
        pEl.classList.add("pagination_page_current");
      } else {
        pEl.classList.add("pagination_page");
        pEl.setAttribute("data-page", i + "");
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
        if (
          (<HTMLDivElement>e.target).classList.contains(
            "admin_table_cell-multiple_checkbox"
          )
        ) {
          return false;
        }
        var target = <HTMLElement>e.target;
        if (target.classList.contains("preventredirect")) {
          return;
        }
        var el = <HTMLDivElement>e.currentTarget;
        var url = el.getAttribute("data-url");

        if (e.shiftKey || e.metaKey || e.ctrlKey) {
          var openedWindow = window.open(url, "newwindow" + new Date());
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
      });
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
    if (
      e.keyCode == 9 ||
      e.keyCode == 16 ||
      e.keyCode == 17 ||
      e.keyCode == 18
    ) {
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
    setInterval(() => {
      if (this.changed == true && Date.now() - this.changedTimestamp > 500) {
        this.changed = false;
        this.load();
      }
    }, 200);
  }
}

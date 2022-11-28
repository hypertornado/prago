class List {
  minCellWidth = 50;
  normalCellWidth = 100;
  maxCellWidth = 500;

  settings: ListSettings;

  adminPrefix: string;
  typeName: string;

  rootContent: HTMLDivElement;

  list: HTMLDivElement;

  listHeaderContainer: HTMLDivElement;
  listHeader: HTMLDivElement;
  listTable: HTMLDivElement;
  listFooter: HTMLDivElement;

  tableContent: HTMLElement;
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

  loadStats: Boolean;

  currentRequest: XMLHttpRequest;

  constructor(list: HTMLDivElement) {
    this.list = list;

    this.rootContent = document.querySelector(".root_content");
    this.listHeaderContainer = this.rootContent.querySelector(
      ".list_header_container"
    );
    this.listTable = this.list.querySelector(".list_table");
    this.listHeader = this.list.querySelector(".list_header");
    this.listFooter = this.list.querySelector(".list_footer");

    var dateFilterInputs = list.querySelectorAll<HTMLInputElement>(
      ".list_filter_date_input"
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

    this.typeName = list.getAttribute("data-type");
    if (!this.typeName) {
      return;
    }

    this.progress = <HTMLProgressElement>list.querySelector(".list_progress");

    this.tableContent = <HTMLElement>list.querySelector(".list_table_content");
    //this.tableContent.textContent = "";

    this.bindFilter(urlParams);

    this.adminPrefix = document.body.getAttribute("data-admin-prefix");

    this.defaultOrderColumn = list.getAttribute("data-order-column");
    if (list.getAttribute("data-order-desc") == "true") {
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

    this.defaultVisibleColumnsStr = list.getAttribute("data-visible-columns");
    var visibleColumnsStr = this.defaultVisibleColumnsStr;
    if (urlParams.get("_columns")) {
      visibleColumnsStr = urlParams.get("_columns");
    }

    let visibleColumnsArr = visibleColumnsStr.split(",");
    let visibleColumnsMap: any = {};
    for (var i = 0; i < visibleColumnsArr.length; i++) {
      visibleColumnsMap[visibleColumnsArr[i]] = true;
    }

    this.itemsPerPage = parseInt(list.getAttribute("data-items-per-page"));
    this.paginationSelect = <HTMLSelectElement>(
      document.querySelector(".list_settings_pages")
    );
    this.paginationSelect.addEventListener("change", this.load.bind(this));

    this.statsCheckboxSelectCount = document.querySelector(".list_stats_limit");
    this.statsCheckboxSelectCount.addEventListener("change", () => {
      this.filterChanged();
    });

    this.statsContainer = document.querySelector(".list_stats_container");

    this.multiple = new ListMultiple(this);

    this.settings.bindOptions(visibleColumnsMap);
    this.bindOrder();
    this.bindInitialHeaderWidths();
    this.bindResizer();
    this.bindHeaderPositionCalculator();
  }

  copyColumnWidths() {
    let totalWidth = this.listHeader.getBoundingClientRect().width;
    this.tableContent.setAttribute("style", "width: " + totalWidth + "px;");

    let headerItems = this.list.querySelectorAll(
      ".list_header > :not(.hidden)"
    );

    let tableRows = this.list.querySelectorAll(".list_row");

    for (let i = 0; i < tableRows.length; i++) {
      let rowItems = tableRows[i].children;

      for (let j = 0; j < headerItems.length; j++) {
        if (j >= rowItems.length) {
          break;
        }
        let headerEl = headerItems[j];
        let tableEl = rowItems[j];

        var clientRect = headerEl.getBoundingClientRect();
        var elWidth = clientRect.width;
        tableEl.setAttribute("style", "width: " + elWidth + "px;");
      }
    }
  }

  load() {
    if (this.currentRequest) {
      this.currentRequest.abort();
    }

    this.list.classList.add("list-loading");
    //this.progress.classList.remove("list_progress-inactive");
    var request = new XMLHttpRequest();

    this.currentRequest = request;

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

    if (this.loadStats) {
      this.statsContainer.innerHTML = '<div class="progress"></div>';
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
      this.currentRequest = null;
      this.tableContent.innerHTML = "";
      if (request.status == 200) {
        var response = JSON.parse(request.response);

        this.tableContent.innerHTML = response.Content;
        var countStr = response.CountStr;

        this.list.querySelector(".list_count").textContent = countStr;
        this.statsContainer.innerHTML = response.StatsStr;
        this.listFooter.innerHTML = response.FooterStr;
        bindReOrder();
        this.bindPagination();
        this.bindClick();
        if (this.multiple.hasMultipleActions()) {
          this.multiple.bindMultipleActionCheckboxes();
        }
      } else {
        console.error("error while loading list");
      }
      this.copyColumnWidths();
      this.list.classList.remove("list-loading");
    });
    request.send(JSON.stringify({}));
  }

  colorActiveFilterItems() {
    let itemsToColor = this.getFilterData();
    var filterItems: NodeListOf<HTMLDivElement> = this.list.querySelectorAll(
      ".list_header_item_filter"
    );
    for (var i = 0; i < filterItems.length; i++) {
      var item = filterItems[i];
      let name = item.getAttribute("data-name");
      if (itemsToColor[name]) {
        item.classList.add("list_filteritem-colored");
      } else {
        item.classList.remove("list_filteritem-colored");
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
    var paginationEl = this.list.querySelector(".pagination");
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
    var rows = this.list.querySelectorAll(".list_row");
    for (var i = 0; i < rows.length; i++) {
      var row = <HTMLTableRowElement>rows[i];
      var id = row.getAttribute("data-id");
      row.addEventListener("click", (e) => {
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
    }
  }

  bindOrder() {
    this.renderOrder();
    var headers = this.list.querySelectorAll(".list_header_item_name-canorder");
    for (var i = 0; i < headers.length; i++) {
      var header = <HTMLAnchorElement>headers[i];
      header.addEventListener("click", (e) => {
        var el = <HTMLAnchorElement>e.currentTarget;
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

  bindResizer() {
    var resizers = this.list.querySelectorAll(".list_header_item_resizer");

    for (var i = 0; i < resizers.length; i++) {
      var resizer = <HTMLAnchorElement>resizers[i];
      let parentEl: HTMLDivElement = <HTMLDivElement>resizer.parentElement;

      let naturalWidth = parseInt(parentEl.getAttribute("data-natural-width"));

      resizer.addEventListener("drag", (e) => {
        var clientRect = parentEl.getBoundingClientRect();
        var clientX = clientRect.left;

        if (e.clientX == 0) {
          return false;
        }
        let width = e.clientX - clientX;
        this.setCellWidth(parentEl, width);
      });

      resizer.addEventListener("dblclick", (e) => {
        let width = this.getCellWidth(parentEl);

        if (width == this.minCellWidth) {
          this.setCellWidth(parentEl, naturalWidth);
        } else {
          if (width == naturalWidth) {
            this.setCellWidth(parentEl, this.maxCellWidth);
          } else {
            if (width == this.maxCellWidth) {
              this.setCellWidth(parentEl, this.minCellWidth);
            } else {
              this.setCellWidth(parentEl, naturalWidth);
            }
          }
        }
        this.copyColumnWidths();
      });

      resizer.addEventListener("dragend", (e) => {
        this.copyColumnWidths();
      });
    }
  }

  getCellWidth(cell: HTMLDivElement): number {
    return cell.getBoundingClientRect().width;
  }

  setCellWidth(cell: HTMLDivElement, width: number) {
    if (width < this.minCellWidth) {
      width = this.minCellWidth;
    }
    if (width > this.maxCellWidth) {
      width = this.maxCellWidth;
    }

    let cellName = cell.getAttribute("data-name");

    if (width + "" == cell.getAttribute("data-natural-width")) {
      this.webStorageDeleteWidth(cellName);
    } else {
      this.webStorageSetWidth(cellName, width);
    }

    cell.setAttribute("style", "width: " + width + "px;");
  }

  bindInitialHeaderWidths() {
    let headerItems = this.list.querySelectorAll(".list_header_item");
    for (var i = 0; i < headerItems.length; i++) {
      var itemEl = <HTMLDivElement>headerItems[i];

      let width = parseInt(itemEl.getAttribute("data-natural-width"));

      let cellName = itemEl.getAttribute("data-name");

      let savedWidth = this.webStorageLoadWidth(cellName);
      if (savedWidth > 0) {
        width = savedWidth;
      }
      this.setCellWidth(itemEl, width);
    }
  }

  webStorageWidthName(cell: string): string {
    let tableName = this.typeName;
    return "prago_cellwidth_" + tableName + "_" + cell;
  }

  webStorageLoadWidth(cell: string): number {
    let val = window.localStorage[this.webStorageWidthName(cell)];
    if (val) {
      return parseInt(val);
    }
    return 0;
  }

  webStorageSetWidth(cell: string, width: number) {
    window.localStorage[this.webStorageWidthName(cell)] = width;
  }

  webStorageDeleteWidth(cell: string) {
    window.localStorage.removeItem(this.webStorageWidthName(cell));
  }

  renderOrder() {
    var headers = this.list.querySelectorAll(".list_header_item_name-canorder");
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
    var items = this.list.querySelectorAll(".list_filter_item");
    for (var i = 0; i < items.length; i++) {
      var item = <HTMLInputElement>items[i];
      var typ = item.getAttribute("data-typ");
      var layout = item.getAttribute("data-filter-layout");
      if (item.classList.contains("list_filter_item-relations")) {
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
    var filterFields = this.list.querySelectorAll(".list_header_item_filter");
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
    this.page = 1;
    this.changed = true;
    this.changedTimestamp = Date.now();
    this.list.classList.add("list-loading");
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

  bindHeaderPositionCalculator() {
    this.listHeaderPositionChanged();

    window.addEventListener(
      "resize",
      this.listHeaderPositionChanged.bind(this)
    );

    this.list.addEventListener(
      "scroll",
      this.listHeaderPositionChanged.bind(this)
    );

    this.rootContent.addEventListener(
      "scroll",
      this.listHeaderPositionChanged.bind(this)
    );

    this.listTable.addEventListener(
      "scroll",
      this.listHeaderPositionChanged.bind(this)
    );
  }

  listHeaderPositionChanged() {
    let rect = this.rootContent.getBoundingClientRect();

    var leftScroll = -this.listTable.scrollLeft;
    this.listHeader.setAttribute("style", "margin-left: " + leftScroll + "px;");

    this.listHeaderContainer.setAttribute(
      "style",
      "top: " + rect.top + "px; left: " + rect.left + "px;"
    );

    return true;
  }
}

class List {
  minCellWidth = 50;
  normalCellWidth = 100;
  maxCellWidth = 500;

  settings: ListSettings;

  typeName: string;

  rootContent: HTMLDivElement;

  listEl: HTMLDivElement;

  listHeaderContainer: HTMLDivElement;
  listHeader: HTMLDivElement;
  listTable: HTMLDivElement;
  listFooter: HTMLDivElement;

  tableContent: HTMLElement;
  changed: boolean;
  changedTimestamp: number;

  defaultOrderColumn: string;
  orderColumn: string;
  defaultOrderDesc: boolean;
  orderDesc: boolean;
  page: number;

  defaultVisibleColumnsStr: string;
  visibleColumnsStr: string;

  progress: HTMLProgressElement;

  itemsPerPage: number;

  multiple: ListMultiple;

  currentRequest: XMLHttpRequest;

  filter: ListFilter;

  constructor(listEl: HTMLDivElement) {
    this.listEl = listEl;

    this.rootContent = document.querySelector(".root_content");
    this.listHeaderContainer = this.rootContent.querySelector(
      ".list_header_container"
    );
    this.listTable = this.listEl.querySelector(".list_table");
    this.listHeader = this.listEl.querySelector(".list_header");
    this.listFooter = this.listEl.querySelector(".list_footer");

    this.defaultVisibleColumnsStr = listEl.getAttribute("data-visible-columns");
    this.visibleColumnsStr = this.defaultVisibleColumnsStr;

    this.settings = new ListSettings(this);

    let urlParams = new URLSearchParams(window.location.search);

    this.page = parseInt(urlParams.get("_page"));
    if (!this.page) {
      this.page = 1;
    }

    this.typeName = listEl.getAttribute("data-type");
    if (!this.typeName) {
      return;
    }

    this.progress = <HTMLProgressElement>listEl.querySelector(".list_progress");

    this.tableContent = <HTMLElement>listEl.querySelector(".list_table_content");

    this.filter = new ListFilter(this, urlParams);

    this.inputPeriodicListener();

    this.defaultOrderColumn = listEl.getAttribute("data-order-column");
    if (listEl.getAttribute("data-order-desc") == "true") {
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
    
    this.itemsPerPage = parseInt(listEl.getAttribute("data-items-per-page"));

    this.multiple = new ListMultiple(this);

    this.settings.setVisibleColumns();

    this.bindOrder();
    this.bindInitialHeaderWidths();
    this.bindResizer();
    this.bindHeaderPositionCalculator();
  }

  copyColumnWidths() {
    let totalWidth = this.listHeader.getBoundingClientRect().width;

    let headerItems = this.listEl.querySelectorAll(
      ".list_header > :not(.hidden)"
    );

    let widths = [];

    for (let j = 0; j < headerItems.length; j++) {
      let headerEl = headerItems[j];
      var clientRect = headerEl.getBoundingClientRect();
      var elWidth = clientRect.width;
      widths.push(elWidth);
    }

    let tableRows = this.listEl.querySelectorAll(".list_row");
    for (let i = 0; i < tableRows.length; i++) {
      let rowItems = tableRows[i].children;

      for (let j = 0; j < widths.length; j++) {
        if (j >= rowItems.length) {
          break;
        }
        let tableEl = <HTMLDivElement>rowItems[j];
        tableEl.style.width = widths[j] + "px";
      }
    }

    let placeholderItems = this.listEl.querySelectorAll(
      ".list_tableplaceholder_row"
    );
    if (placeholderItems.length > 0) {
      let placeholderWidth = totalWidth;
      for (let i = 0; i < placeholderItems.length; i++) {
        let item: HTMLDivElement = <HTMLDivElement>placeholderItems[i];
        item.style.width = placeholderWidth + "px";
      }
    }
  }

  load() {
    if (this.currentRequest) {
      this.currentRequest.abort();
    }
    this.listEl.classList.add("list-loading");

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
    

    let filterData = this.filter.getFilterData();
    for (var k in filterData) {
      params[k] = filterData[k];
    }
    this.filter.colorActiveFilterItems();


    var encoded = encodeParams(params);
    window.history.replaceState(
      null,
      null,
      document.location.pathname + encoded
    );

    var columns = this.visibleColumnsStr;
    if (columns != this.defaultVisibleColumnsStr) {
      params["_columns"] = columns;
    }
    params["_pagesize"] = this.itemsPerPage;

    encoded = encodeParams(params);

    request.open(
      "GET",
      "/admin/" + this.typeName + "/api/list" + encoded,
      true
    );

    request.addEventListener("load", () => {
      this.currentRequest = null;
      this.tableContent.innerHTML = "";
      if (request.status == 200) {
        var response = JSON.parse(request.response);
        this.tableContent.innerHTML = response.Content;
        this.listFooter.innerHTML = response.FooterStr;
        bindReOrder();
        this.bindSettingsButton();
        this.bindPagination();
        this.bindClick();
        this.bindFetchStats();
        if (this.multiple.hasMultipleActions()) {
          this.multiple.bindMultipleActionCheckboxes();
        }
      } else {
        new Alert("Chyba při načítání položek.");
        console.error("error while loading list");
      }
      this.copyColumnWidths();
      this.listEl.classList.remove("list-loading");
      this.listHeaderContainer.classList.add("list_header_container-visible");
    });
    request.send(JSON.stringify({}));
  }

  paginationChange(e: any) {
    var el = <HTMLAnchorElement>e.target;
    var page = parseInt(el.getAttribute("data-page"));
    this.page = page;
    this.load();
    e.preventDefault();
    return false;
  }

  bindSettingsButton() {
    let btn: HTMLButtonElement = this.listEl.querySelector(".list_settings_btn2");
    this.settings.bindSettingsBtn(btn);
  }

  bindPagination() {
    var paginationEl = this.listEl.querySelector(".pagination");
    var totalPages = parseInt(paginationEl.getAttribute("data-total"));
    var selectedPage = parseInt(paginationEl.getAttribute("data-selected"));
    if (totalPages < 2) {
      return;
    }

    let beforeItemWasShown = true;

    for (var i = 1; i <= totalPages; i++) {
      let shouldBeShown = false;
      let maxDistance = 2;
      if (Math.abs(1 - i) <= maxDistance) {
        shouldBeShown = true;
      }
      if (Math.abs(totalPages - i) <= maxDistance) {
        shouldBeShown = true;
      }
      if (Math.abs(selectedPage - i) <= maxDistance) {
        shouldBeShown = true;
      }

      if (shouldBeShown) {
        if (!beforeItemWasShown) {
          let delimiterEl = document.createElement("div");
          delimiterEl.classList.add("pagination_page_delimiter");
          delimiterEl.innerText = "…";
          paginationEl.appendChild(delimiterEl);
        }

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
        beforeItemWasShown = true;
      } else {
        beforeItemWasShown = false;
      }
    }
  }

  bindFetchStats() {
    var cells = this.listEl.querySelectorAll(".list_cell[data-fetch-url]");
    for (var i = 0; i < cells.length; i++) {
      let cell = <HTMLDivElement>cells[i];
      let url = cell.getAttribute("data-fetch-url");
      if (!url) {
        continue;
      }

      if (cell.classList.contains("list_cell-fetched")) {
        continue;
      }

      if (!document.contains(cell)) {
        //we reloaded list
        continue;
      }

      let cellContentSpan = <HTMLSpanElement>(
        cell.querySelector(".list_cell_name")
      );

      fetch(url)
        .then((data) => {
          return data.json();
        })
        .then((data) => {
          cellContentSpan.innerText = data.Value;
          cell.setAttribute("title", data.Value);
          cell.classList.add("list_cell-fetched");
          this.bindFetchStats();
        })
        .catch((error) => {
          cellContentSpan.innerText = "⚠️";
          console.error("cant fetch data:", error);
        });
      return;
    }
  }

  bindClick() {
    var rows = this.listEl.querySelectorAll(".list_row");
    for (var i = 0; i < rows.length; i++) {
      let row = <HTMLTableRowElement>rows[i];

      row.addEventListener("contextmenu", this.contextClick.bind(this));

      row.addEventListener("click", (e) => {
        var el = <HTMLDivElement>e.currentTarget;
        var url = el.getAttribute("data-url");

        if (e.altKey) {
          url += "/edit";

          let targetEl = <HTMLDivElement>e.target;
          targetEl = targetEl.closest(".list_cell");
          let focusID = targetEl.getAttribute("data-cell-id");
          if (focusID) {
            url += "?_focus=" + focusID;
          }
          new PopupForm(url, (data: any) => {
            this.load();
          });
          return
        }

        if (e.shiftKey || e.metaKey || e.ctrlKey) {
          var openedWindow = window.open(url, "newwindow" + new Date() + Math.random());
          openedWindow.focus();
          return;
        }
        window.location.href = url;
      });
    }
  }

  createCmenu(e: PointerEvent, rowEl: HTMLDivElement, alignByElement?: boolean) {
    rowEl.classList.add("list_row-context");

    let actions = JSON.parse(rowEl.getAttribute("data-actions"));

    var commands: CMenuCommand[] = [];

    let allowPopupForm = true;
    if (e.altKey || e.metaKey || e.shiftKey || e.ctrlKey) {
      allowPopupForm = false;
    }

    for (let action of actions.MenuButtons) {

      let actionURL = null;
      let handler = null;
      if (action.FormURL && allowPopupForm) {
        handler = () => {
          new PopupForm(action.FormURL, (data: any) => {
            this.load();
          })
        }
      } else {
        actionURL = action.URL;
      }

      commands.push({
        Icon: action.Icon,
        Name: action.Name,
        URL: actionURL,
        Style: action.Style,
        Handler: handler,
      });
    }

    let name = rowEl.getAttribute("data-name");
    let preName = rowEl.getAttribute("data-prename")
    if (name == preName) {
      preName = null;
    }

    cmenu({
      Event: e,
      AlignByElement: alignByElement,
      ImageURL: rowEl.getAttribute("data-image-url"),
      PreName: preName,
      Name: name,
      Description: rowEl.getAttribute("data-description"),
      Commands: commands,
      DismissHandler: () => {
        rowEl.classList.remove("list_row-context");
      },
    });
    e.preventDefault();
  }

  contextClick(e: PointerEvent) {
    let rowEl = <HTMLDivElement>e.currentTarget;
    this.createCmenu(e, rowEl, false);
  }

  bindOrder() {
    this.renderOrder();
    var headers = this.listEl.querySelectorAll(".list_header_item_name-canorder");
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
    var resizers = this.listEl.querySelectorAll(".list_header_item_resizer");

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
    this.copyColumnWidths();
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
    let headerItems = this.listEl.querySelectorAll(".list_header_item");
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
    var headers = this.listEl.querySelectorAll(".list_header_item_name-canorder");
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

    this.listEl.addEventListener(
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

    let scrolledClassName = "list_header_container-scrolled";
    if (this.rootContent.scrollTop > 50) {
      this.listHeaderContainer.classList.add(scrolledClassName);
    } else {
      this.listHeaderContainer.classList.remove(scrolledClassName);
    }

    return true;
  }
}

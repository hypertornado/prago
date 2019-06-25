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
  filterInputs: NodeListOf<Element>;
  changed: boolean;
  changedTimestamp: number;
  
  orderColumn: string;
  orderDesc: boolean;
  page: number;

  prefilterField: string;
  prefilterValue: string;

  progress: HTMLProgressElement;

  settingsEl: HTMLDivElement;
  //openbutton: HTMLButtonElement;
  //closebutton: HTMLButtonElement;

  constructor(el: HTMLDivElement, openbutton: HTMLButtonElement) {
    this.el = el;
    //this.openbutton = openbutton;
    //this.closebutton = this.el.querySelector(".admin_tablesettings_close");
    this.settingsEl = this.el.querySelector(".admin_tablesettings");

    this.page = 1;

    this.typeName = el.getAttribute("data-type");
    if (!this.typeName) {
      return;
    }

    this.progress = <HTMLProgressElement>el.querySelector(".admin_table_progress");

    this.tbody = <HTMLElement>el.querySelector("tbody");
    this.tbody.textContent = "";

    this.bindFilter();

    this.adminPrefix = document.body.getAttribute("data-admin-prefix");
    this.prefilterField = el.getAttribute("data-prefilter-field");
    this.prefilterValue = el.getAttribute("data-prefilter-value");

    this.orderColumn = el.getAttribute("data-order-column");
    if (el.getAttribute("data-order-desc") == "true") {
      this.orderDesc = true;
    } else {
      this.orderDesc = false;
    }

    //this.openbutton.addEventListener("click", this.toggleShowHide.bind(this));
    //this.closebutton.addEventListener("click", this.toggleShowHide.bind(this));
    this.bindOptions();
    this.bindOrder();
  }

  /*toggleShowHide() {
    this.settingsEl.classList.toggle("hidden");
    this.openbutton.classList.toggle("hidden");
  }*/

  load() {
    this.progress.classList.remove("hidden");
    var request = new XMLHttpRequest();
    request.open("POST", this.adminPrefix + "/_api/list/" + this.typeName + document.location.search, true);
    request.addEventListener("load", () => {
      this.tbody.innerHTML = "";
      if (request.status == 200) {
        this.tbody.innerHTML = request.response;
        var count = request.getResponseHeader("X-Count");
        var totalCount = request.getResponseHeader("X-Total-Count");
        var countStr: string = count + " / " + totalCount;
        this.el.querySelector(".admin_table_count").textContent = countStr;
        bindOrder();
        //bindDelete();
        this.bindPagination();
        this.bindClick();
        this.tbody.classList.remove("admin_table_loading");
      } else {
        console.error("error while loading list");
      }
      this.progress.classList.add("hidden");
    });
    var requestData = this.getListRequest();
    request.send(JSON.stringify(requestData));
  }

  bindOptions() {
    var columns: NodeListOf<HTMLInputElement> = this.el.querySelectorAll(".admin_tablesettings_column");
    for (var i = 0; i < columns.length; i++) {
      columns[i].addEventListener("change", () => {
        this.changedOptions();
      });
    }
    this.changedOptions();
  }

  changedOptions() {
    var columns: any = this.getSelectedColumns();

    var headers: NodeListOf<HTMLDivElement> = this.el.querySelectorAll(".admin_list_orderitem");
    for (var i = 0; i < headers.length; i++) {
      var name = headers[i].getAttribute("data-name");
      if (columns[name]) {
        headers[i].classList.remove("hidden");
      } else {
        headers[i].classList.add("hidden");
      }
    }

    var filters: NodeListOf<HTMLDivElement> = this.el.querySelectorAll(".admin_list_filteritem");
    for (var i = 0; i < filters.length; i++) {
      var name = filters[i].getAttribute("data-name");
      if (columns[name]) {
        filters[i].classList.remove("hidden");
      } else {
        filters[i].classList.add("hidden");
      }
    }

    this.load();
  }

  bindPagination() {
    var pages = this.el.querySelectorAll(".pagination_page");
    for (var i = 0; i < pages.length; i++) {
      var pageEl = <HTMLAnchorElement>pages[i];
      pageEl.addEventListener("click", (e) => {
        var el = <HTMLAnchorElement>e.target;
        var page = parseInt(el.getAttribute("data-page"));
        this.page = page;
        this.load();
        e.preventDefault();
        return false;
      })
    }
  }

  bindClick() {
    var rows = this.el.querySelectorAll(".admin_table_row");
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
          var openedWindow = window.open(url, "newwindow");
          console.log(openedWindow);
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

  getSelectedColumns(): any {
    var columns: any = {};
    var checked: NodeListOf<HTMLInputElement> = this.el.querySelectorAll(".admin_tablesettings_column:checked");
    for (var i = 0; i < checked.length; i++) {
      columns[checked[i].getAttribute("data-column-name")] = true;
    }
    return columns;
  }

  getListRequest(): any {

    var ret: any = {};
    ret.Page = this.page;
    ret.OrderBy = this.orderColumn;
    ret.OrderDesc = this.orderDesc;
    ret.Filter = this.getFilterData();
    ret.PrefilterField = this.prefilterField;
    ret.PrefilterValue = this.prefilterValue;
    ret.Columns = this.getSelectedColumns();
    return ret;
  }

  getFilterData(): any {
    var ret: any = {};
    var items = this.el.querySelectorAll(".admin_table_filter_item");
    for (var i = 0; i < items.length; i++) {
      var item = <HTMLInputElement>items[i];
      var typ = item.getAttribute("data-typ");
      var val = item.value.trim();
      if (val) {
        ret[typ] = val;
      }
    }
    return ret;
  }

  bindFilter() {
    this.bindFilterRelations();
    this.filterInputs = this.el.querySelectorAll(".admin_table_filter_item");
    for (var i = 0; i < this.filterInputs.length; i++) {
      var input: HTMLInputElement = <HTMLInputElement>this.filterInputs[i];
      input.addEventListener("input", this.inputListener.bind(this));
    }
    this.inputPeriodicListener();
  }

  inputListener(e: any) {
    if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
      return;
    }
    this.tbody.classList.add("admin_table_loading");
    this.page = 1;
    this.changed = true;
    this.changedTimestamp = Date.now();
    this.progress.classList.remove("hidden");
  }

  bindFilterRelations() {
    var els = this.el.querySelectorAll(".admin_table_filter_item-relations");
    for (var i = 0; i < els.length; i++) {
      this.bindFilterRelation(<HTMLSelectElement>els[i]);
    }
  }

  bindFilterRelation(select: HTMLSelectElement) {
    var typ = select.getAttribute("data-typ");

    var adminPrefix = document.body.getAttribute("data-admin-prefix");
    var request = new XMLHttpRequest();
    request.open("GET", adminPrefix + "/_api/resource/" + typ, true);

    request.addEventListener("load", () => {
      if (request.status == 200) {
        var resp = JSON.parse(request.response);
        for (var item of resp) {
          var option = document.createElement("option");
          option.setAttribute("value", item.id);
          option.innerText = item.name;
          select.appendChild(option);
        }
      } else {
        console.error("Error wile loading relation " + typ + ".");
      }
    });
    request.send();
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
function bindLists() {
  var els = document.getElementsByClassName("admin_table-list");
  for (var i = 0; i < els.length; i++) {
    new List(<HTMLTableElement>els[i]);
  }
}

class List {
  adminPrefix: string;
  typeName: string;

  tbody: HTMLElement;
  el: HTMLTableElement;
  filterInputs: NodeListOf<Element>;
  changed: boolean;
  changedTimestamp: number;
  
  orderColumn: string;
  orderDesc: boolean;
  page: number;

  progress: HTMLProgressElement;

  constructor(el: HTMLTableElement) {
    this.el = el;

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

    this.orderColumn = el.getAttribute("data-order-column");
    if (el.getAttribute("data-order-desc") == "true") {
      this.orderDesc = true;
    } else {
      this.orderDesc = false;
    }

    this.bindOrder();

    this.load();
  }

  load() {
    this.progress.classList.remove("hidden");
    var request = new XMLHttpRequest();
    request.open("POST", this.adminPrefix + "/_api/list/" + this.typeName + document.location.search, true);
    request.addEventListener("load", () => {
      this.tbody.innerHTML = "";
      if (request.status == 200) {
        this.tbody.innerHTML = request.response;
        var count = request.getResponseHeader("X-Total-Count");
        this.el.querySelector(".admin_table_count").textContent = count;
        bindOrder();
        bindDelete();
        this.bindPage();
      } else {
        console.error("error while loading list");
      }
      this.progress.classList.add("hidden");
    });
    var requestData = this.getListRequest();
    request.send(JSON.stringify(requestData));
  }

  bindPage() {
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

  bindOrder() {
    this.renderOrder();
    var headers = this.el.querySelectorAll(".admin_table_orderheader");
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
    var headers = this.el.querySelectorAll(".admin_table_orderheader");
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

  getListRequest(): any {
    var ret: any = {};
    ret.Page = this.page;
    ret.OrderBy = this.orderColumn;
    ret.OrderDesc = this.orderDesc;
    ret.Filter = this.getFilterData();
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
      input.addEventListener("change", this.inputListener.bind(this));
      input.addEventListener("keyup", this.inputListener.bind(this));
    }
    this.inputPeriodicListener();
  }

  inputListener(e: any) {
    if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
      return;
    }
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

const getParams = (query: string) => {
  if (!query) {
    return { };
  }

  return (/^[?#]/.test(query) ? query.slice(1) : query)
    .split('&')
    .reduce((params: any, param: any) => {
      let [ key, value ] = param.split('=');
      params[key] = value ? decodeURIComponent(value.replace(/\+/g, ' ')) : '';
      return params;
    }, { });
};
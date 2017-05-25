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

  constructor(el: HTMLTableElement) {
    this.el = el;

    this.page = 1;

    this.typeName = el.getAttribute("data-type");
    if (!this.typeName) {
      return;
    }


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
    var request = new XMLHttpRequest();
    request.open("POST", this.adminPrefix + "/_api/list/" + this.typeName + document.location.search, true);
    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.tbody.innerHTML = request.response;
        bindOrder();
        bindDelete();
        this.bindPage();
      } else {
        alert("error");
      }
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
    ret.Filter = {};
    return ret;
  }

  bindFilter() {
    this.filterInputs = this.el.querySelectorAll(".admin_table_filter_item");
    for (var i = 0; i < this.filterInputs.length; i++) {
      var input: HTMLInputElement = <HTMLInputElement>this.filterInputs[i];
      input.addEventListener("change", this.inputListener.bind(this));
      input.addEventListener("keyup", this.inputListener.bind(this));
    }
    this.inputPeriodicListener();
  }

  inputListener() {
    this.changed = true;
    this.changedTimestamp = Date.now();
  }

  inputPeriodicListener() {
    setInterval(() =>{
      if (this.changed == true && Date.now() - this.changedTimestamp > 500) {
        this.changed = false;
        console.log("X");
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
    .reduce((params, param) => {
      let [ key, value ] = param.split('=');
      params[key] = value ? decodeURIComponent(value.replace(/\+/g, ' ')) : '';
      return params;
    }, { });
};
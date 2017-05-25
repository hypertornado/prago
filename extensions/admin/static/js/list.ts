function bindLists() {
  var els = document.getElementsByClassName("admin_table-list");
  for (var i = 0; i < els.length; i++) {
    new List(<HTMLTableElement>els[i]);
  }
}

class List {
  tbody: HTMLElement;
  el: HTMLTableElement;
  filterInputs: NodeListOf<Element>;
  changed: boolean;
  changedTimestamp: number;
  orderColumn: string;
  orderDesc: boolean;

  constructor(el: HTMLTableElement) {
    this.el = el;
    this.parseSearch(document.location.search);

    var typeName = el.getAttribute("data-type");
    if (!typeName) {
      return;
    }


    this.tbody = <HTMLElement>el.querySelector("tbody");
    this.tbody.textContent = "";

    this.bindFilter();

    var adminPrefix = document.body.getAttribute("data-admin-prefix");

    this.orderColumn = el.getAttribute("data-order-column");
    if (el.getAttribute("data-order-desc") == "true") {
      this.orderDesc = true;
    } else {
      this.orderDesc = false;
    }

    console.log(this.orderColumn, this.orderDesc);

    var request = new XMLHttpRequest();
    request.open("POST", adminPrefix + "/_api/list/" + typeName + document.location.search, true);
    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.tbody.innerHTML = request.response;
        bindOrder();
        bindDelete();
      } else {
        alert("error");
      }
    });

    var requestData = this.getListRequest();
    console.log(requestData);

    request.send(JSON.stringify(requestData));
  }

  getListRequest(): any {
    var ret: any = {};
    ret.Page = 1;
    ret.OrderBy = this.orderColumn;
    ret.OrderDesc = this.orderDesc;
    ret.Filter = {};
    return ret;
  }

  parseSearch(url: string) {
    //url = url.substr(1);
    //console.log(getParams(url))
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
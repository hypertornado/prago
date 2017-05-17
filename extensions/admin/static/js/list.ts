function bindLists() {
  var els = document.getElementsByClassName("admin_table-list");
  for (var i = 0; i < els.length; i++) {
    new List(<HTMLTableElement>els[i]);
  }
}

class List {
  tbody: HTMLElement;

  constructor(el: HTMLTableElement) {
    var typeName = el.getAttribute("data-type");
    if (!typeName) {
      return;
    }


    this.tbody = <HTMLElement>el.querySelector("tbody");
    this.tbody.textContent = "";

    var adminPrefix = document.body.getAttribute("data-admin-prefix");

    var request = new XMLHttpRequest();
    request.open("GET", adminPrefix + "/_api/list/" + typeName + document.location.search, true);
    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.tbody.innerHTML = request.response;
        bindOrder();
        bindDelete();
      } else {
        alert("error");
      }
    });
    request.send();
  }
}
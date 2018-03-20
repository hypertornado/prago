function bindRelationsView() {
  var els = document.querySelectorAll(".admin_item_view_relation");
  for (var i = 0; i < els.length; i++) {
    new RelationsView(<HTMLDivElement>els[i]);
  }
}

class RelationsView {
  
  constructor(el: HTMLDivElement) {

    var idStr = el.getAttribute("data-id");
    var typ = el.getAttribute("data-type");

    var adminPrefix = document.body.getAttribute("data-admin-prefix");

    var request = new XMLHttpRequest();
    request.open("GET", adminPrefix + "/_api/resource/" + typ + "/" + idStr, true);

    request.addEventListener("load", () => {
      el.innerHTML = "";
      if (request.status == 200) {
        var resp = JSON.parse(request.response);

        var link = document.createElement("a");
        link.setAttribute("href", adminPrefix + "/" + typ + "/" + idStr);
        var name = resp.name;
        if (name == "") {
          name += " ";
        }
        link.textContent = name;

        el.appendChild(link);
      } else {
        el.textContent = "Error while loading";
      }
    })
    request.send();
  }

}
function bindRelations() {
  function bindRelation(el: HTMLElement) {
    var input: HTMLInputElement = <HTMLInputElement>el.getElementsByTagName("input")[0];

    var relationName = input.getAttribute("data-relation");
    var originalValue = input.value;

    var select = document.createElement("select");
    select.classList.add("input");
    select.classList.add("form_input");

    select.addEventListener("change", function() {
      input.value = select.value;
    });

    var adminPrefix = document.body.getAttribute("data-admin-prefix");

    var request = new XMLHttpRequest();
    request.open("GET", adminPrefix + "/_api/resource/" + relationName, true);

    var progress = el.getElementsByTagName("progress")[0];

    request.addEventListener("load", () => {
      if (request.status >= 200 && request.status < 400) {
        var resp = JSON.parse(request.response);
        addOption(select, "0", "", false);

        Array.prototype.forEach.call(resp, function (item: any, i: number){
          var selected = false;
          if (originalValue == item["id"]) {
            selected = true;
          }
          addOption(select, item["id"], item["name"], selected);
        });
        el.appendChild(select);
      } else {
        console.error("Error wile loading relation " + relationName + ".");
      }
      progress.style.display = 'none';
    });

    request.onerror = function() {
      console.error("Error wile loading relation " + relationName + ".");
      progress.style.display = 'none';
    };
    request.send();
  }

  function addOption(select: HTMLSelectElement, value: string, description: string, selected: boolean) {
    var option = document.createElement("option");
    if (selected) {
      option.setAttribute("selected", "selected");
    }
    option.setAttribute("value", value);
    option.innerText = description;
    select.appendChild(option);

  }

  var elements = document.querySelectorAll(".admin_item_relation");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    bindRelation(el);
  });
}
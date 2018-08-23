function bindOrder() {
  function orderTable(el: HTMLElement) {
    var rows = el.getElementsByClassName("admin_table_row");
    Array.prototype.forEach.call(rows, function(item: HTMLElement, i: number){
      bindDraggable(<HTMLTableRowElement>item);
    })

    var draggedElement: HTMLTableRowElement;
    function bindDraggable(row: HTMLTableRowElement) {
      row.setAttribute("draggable", "true");

      row.addEventListener("dragstart", function(ev){
        //el.classList.add("admin_table-dragging");

        row.classList.add("admin_table_row-selected");

        draggedElement = this;
        (ev as DragEvent).dataTransfer.setData('text/plain', '');
        (ev as DragEvent).dataTransfer.effectAllowed = "move";

        var d = document.createElement("div");
        d.style.display = "none";
        (ev as DragEvent).dataTransfer.setDragImage(d, 0, 0);
      });

      row.addEventListener("dragenter", function(ev) {
        var targetEl: HTMLElement = this;
        if (this != draggedElement) {
          var draggedIndex: number = -1;
          var thisIndex: number = -1;
          Array.prototype.forEach.call(el.getElementsByClassName("admin_table_row"), function(item: HTMLElement, i: number) {
            if (item == draggedElement) {
              draggedIndex = i;
            }
            if (item == targetEl) {
              thisIndex = i;
            }
          });
          if (draggedIndex < thisIndex) {
            thisIndex += 1;
          }
          DOMinsertChildAtIndex(targetEl.parentElement, draggedElement, thisIndex);
          //saveOrder();
        }
        return false;
      });

      row.addEventListener("drop", function(ev) {
        saveOrder();
        row.classList.remove("admin_table_row-selected");
        return false;
      });

      row.addEventListener("dragover", function(ev) {
        ev.preventDefault();
      });
    }

    function saveOrder() {
      var adminPrefix = document.body.getAttribute("data-admin-prefix");
      var typ = document.querySelector(".admin_table-order").getAttribute("data-type");

      var ajaxPath = adminPrefix + "/_api/order/" + typ;
      var order: any[] = [];
      var rows = el.getElementsByClassName("admin_table_row");

      Array.prototype.forEach.call(rows, function(item: HTMLElement, i: number) {
        order.push(parseInt(item.getAttribute("data-id")))
      });

      var request = new XMLHttpRequest();
      request.open("POST", ajaxPath, true);

      request.addEventListener("load", () => {
        if (request.status != 200) {
          console.error("Error while saving order.")
        }
      });
      request.send(JSON.stringify({"order": order}))
    }
  }

  var elements = document.querySelectorAll(".admin_table-order");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    orderTable(el);
  });
}
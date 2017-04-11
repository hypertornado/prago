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
        draggedElement = this;
        (ev as DragEvent).dataTransfer.setData('text/plain', '');
      });

      row.addEventListener("drop", function(ev) {
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
          })

          if (draggedIndex <= thisIndex) {
            thisIndex += 1
          }

          DOMinsertChildAtIndex(targetEl.parentElement, draggedElement, thisIndex + 2);

          saveOrder();
        }
        return false;
      })

      row.addEventListener("dragover", function(ev) {
        ev.preventDefault();
      });
    }

    function saveOrder() {
      var ajaxPath = document.location.pathname + "/order"
      var order: any[] = [];
      var rows = el.getElementsByClassName("admin_table_row");

      Array.prototype.forEach.call(rows, function(item: HTMLElement, i: number) {
        order.push(parseInt(item.getAttribute("data-id")))
      });

      var request = new XMLHttpRequest();
      request.open("POST", ajaxPath, true);

      request.onload = function() {
        if (this.status != 200) {
          console.error("Error while saving order.")
        }
      }
      request.send(JSON.stringify({"order": order}))
    }
  }

  var elements = document.querySelectorAll(".admin_table-order");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    orderTable(el);
  });
}
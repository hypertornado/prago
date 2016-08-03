function bindOrder() {

  function orderTable(el) {
    var el = $(el);
    var rows = el.find(".admin_table_row");
    rows.each(function (k, v) {
      bindDraggable(v);
    });

    var draggedElement;
    function bindDraggable(row) {
      $(row).attr("draggable", true);

      $(row).on("dragstart", function(e) {
        draggedElement = this;
        e.originalEvent.dataTransfer.setData('text/plain', '');
      });

      $(row).on("drop", function(e) {
        if (this != draggedElement) {
          var draggedIndex = el.find(".admin_table_row").index(draggedElement);
          var thisIndex = el.find(".admin_table_row").index($(this));

          var append = true;
          if (draggedIndex > thisIndex) {
            append = false;
          }

          draggedElement.remove();
          
          if (append) {
            $(this).after(draggedElement);  
          } else {
            $(this).before(draggedElement);
          }

          saveOrder();
        }
        return false;
      });

      $(row).on("dragover", function(e){
        e.preventDefault();
      });

    }

    function saveOrder() {
      var ajaxPath = document.location.pathname + "/order"
      var order = [];
      var rows = el.find(".admin_table_row");
      rows.each(function (k, v) {
        order.push($(v).data("id"));
      });

      $.ajax({
        "url": ajaxPath,
        "method": "POST",
        "data": JSON.stringify({"order": order}),
        "dataType": "json",
        "contentType": "application/json",
        "error": function() {
          alert("Error while saving order.");
        }
      });
    }
  }


  $(".admin_table-order").each(function (k, v) {
    orderTable(v);
  })
}
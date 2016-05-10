window.onload = function() {
  imagePicker();
}

function imagePicker() {
  var popup = $("#admin_images_popup");

  popup.click(function(e){
    if (e.target == popup[0]) {
      hidePopup();
    }
  });

  var loadedContainer = $("#admin_images_popup .admin_images_popup_box_new_list");
  var selectedContainer = $("#admin_images_popup .admin_images_popup_box_content");
  var doFilter = function() {
    loadedContainer.text("Loading...");
    loadImages("", popup.find(".admin_images_popup_filter").val(), function(items) {
      loadedContainer.text("");
      items.forEach(function (item){
        var img = itemToImg(item);
        img.click(function(e) {
          var cloned = $(e.currentTarget).clone();
          cloned.unbind();
          cloned.click(function (event) {
            $(event.currentTarget).remove();
          });
          bindDraggableEvents(cloned);
          selectedContainer.append(cloned);
        });
        loadedContainer.append(img);
      });
    });
  }
  doFilter();

  var draggedElement;

  function bindDraggableEvents(el) {
    el.on("dragstart", function(e) {
      draggedElement = this;
    });

    el.on("drop", function(e) {
      if (this != draggedElement) {
        var uid = $(this).data("uid");
        var src = $(this).attr("src");

        $(this).data("uid", $(draggedElement).data("uid"));
        $(this).attr("src", $(draggedElement).attr("src"));

        $(draggedElement).data("uid", uid);
        $(draggedElement).attr("src", src);
      }
    });

    el.on("dragover", function(e){
      //console.log("dragover");
      e.preventDefault();
    });
  }

  function itemToImg(item) {
    var img = $("<img>").attr("src", item.Thumb).addClass("admin_images_img");
    img.attr("data-uid", item.UID);
    return img;
  }

  function loadImageToPopup(value) {
    selectedContainer.text("Loading...");
    if (value.length > 0) {
      loadImages(value, "", function(items) {
        selectedContainer.text("");
        items.forEach(function (item){
          var img = itemToImg(item);
          img.attr("draggable", "true");
          bindDraggableEvents(img);
          img.click(function (event) {
            $(event.currentTarget).remove();
          });
          selectedContainer.append(img);
        });
      });
    }
  }

  var connectedItem;

  popup.find(".admin_images_popup_save").click(function() {
    hidePopup();
    var items = [];
    var children = selectedContainer.children();

    for (var i = 0; i < children.length; i++) {
      items.push($(children[i]).data("uid"));
    }

    var str = items.join(",");
    $(connectedItem).find("input").val(str);
    showPreview(connectedItem)
  });

  popup.find(".admin_images_popup_cancel").click(hidePopup);
  popup.find(".admin_images_popup_filter_button").click(doFilter);

  function showPopup(el) {
    connectedItem = el;
    loadImageToPopup($(el).find("input").val());
    popup.show();
  }

  function hidePopup() {
    popup.hide();
  }

  function loadImages(ids, q, handler) {
    $.ajax({
      "url": "/admin/_api/image/list",
      "data": {"ids":ids, "q": q},
      "complete": function (responseData) {
        handler(responseData.responseJSON);
      }
    });
  }

  function bindImage(el) {
    showPreview(el);
    $(el).find(".admin_images_edit").click(function(){
      showPopup(el);
      return false;
    });
    $(el).find(".admin_images_edit").click();
  }

  function showPreview(el) {
    var value = $(el).find("input").val();
    var list = $(el).find(".admin_images_list");
    list.text("Loading...");
    var jsonResponse;
    console.log(value);
    if (value.length > 0) {
      loadImages(value, "", function(items) {
        list.text("");
        items.forEach(function (item){
          var link = $("<a href='/admin/file/"+item.ID+"'></a>");
          link.append($("<img>").attr("src", item.Thumb).addClass("admin_images_img"));
          list.append(link);
        });
      });
    }
  }

  $(".admin_images").each(
    function(i) {
      bindImage(this);
    }
  );
}
function bindImagePicker() {
  var popup = $("#admin_images_popup");
  var adminPrefix = $("body").data("admin-prefix");

  popup.click(function(e){
    if (e.target == popup[0]) {
      hidePopup();
    }
  });

  var loadedContainer = popup.find(".admin_images_popup_box_new_list");
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
      e.preventDefault();
    });
  }

  function itemToImg(item) {
    var img = $("<img>").attr("src", item.Thumb).addClass("admin_images_img");
    img.attr("data-uid", item.UID);
    return img;
  }

  function createDraggableImg(item) {
    var img = itemToImg(item);
    img.attr("draggable", "true");
    bindDraggableEvents(img);
    img.click(function (event) {
      $(event.currentTarget).remove();
    });
    return img;
  }

  function loadImageToPopup(value) {
    if (value.length > 0) {
      selectedContainer.text("Loading...");
      loadImages(value, "", function(items) {
        selectedContainer.text("");
        items.forEach(function (item){
          var img = createDraggableImg(item);
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
    $(el).find("admin_images_popup_box").focus();
    popup.show();
    doFilter();
  }

  function hidePopup() {
    popup.hide();
  }

  function loadImages(ids, q, handler) {
    $.ajax({
      "url": adminPrefix + "/_api/image/list",
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
  }

  function showPreview(el) {
    var value = $(el).find("input").val();
    var list = $(el).find(".admin_images_list");
    list.text("");
    $(el).find(".admin_images_edit").hide();
    $(el).find("progress").show();
    var jsonResponse;
    if (value.length > 0) {
      loadImages(value, "", function(items) {
        doneLoading(el);
        items.forEach(function (item){
          var link = $("<a href='" + adminPrefix + "/file/"+item.ID+"' target='_blank'></a>");
          link.append($("<img>").attr("src", item.Thumb).addClass("admin_images_img"));
          list.append(link);
        });
      });
    } else {
      doneLoading(el);
    }
  }

  function doneLoading(el) {
    var list = $(el).find(".admin_images_list");
    list.text("");
    $(el).find(".admin_images_edit").show();
    $(el).find("progress").hide();
  }

  $(".admin_images_popup_box_upload_btn").click(function (e) {
    var files = $(".admin_images_popup_box_upload input")[0].files
    var data = new FormData();
    $.each(files, function(key, value) {
        data.append("file", value);
    });

    data.append("description", $(".admin_popup_file_description").val());
    $(".admin_popup_file_description").val("");

    $(".admin_images_popup_box_upload_message").text("Uploading...");
    $(".admin_images_popup_box_upload_btn").hide();
    $(".admin_images_popup_box_upload input").hide();

    $.ajax({
        url: adminPrefix + "/_api/image/upload",
        type: 'POST',
        data: data,
        cache: false,
        dataType: 'json',
        processData: false, // Don't process the files
        contentType: false, // Set content type to false as jQuery will tell the server its a query string request
        success: function(items) {
          items.forEach(function (item){
            var img = createDraggableImg(item);
            selectedContainer.append(img);
          });

          $(".admin_images_popup_box_upload_message").text("Uploaded successfully.");
          $(".admin_images_popup_box_upload_btn").show();
          $(".admin_images_popup_box_upload input").show();
        },
        error: function() {
            $(".admin_images_popup_box_upload_message").text("Error while uploading files.");
            $(".admin_images_popup_box_upload_btn").show();
            $(".admin_images_popup_box_upload input").show();
        }
    });
  });

  $(".admin_images").each(
    function() {
      bindImage(this);
    }
  );
}
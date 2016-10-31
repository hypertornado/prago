function bindImagePicker() {
  var popup = document.getElementById("admin_images_popup");
  var adminPrefix = document.body.getAttribute("data-admin-prefix");

  popup.addEventListener("click", function(e: Event) {
    if (e.target == popup) {
      hidePopup();
    }
  });

  var loadedContainer = popup.getElementsByClassName("admin_images_popup_box_new_list")[0];
  var selectedContainer = popup.getElementsByClassName("admin_images_popup_box_content")[0];
  var doFilter = function() {
    loadedContainer.textContent = "Loading...";
    var popupFilter: HTMLInputElement = <HTMLInputElement>popup.getElementsByClassName("admin_images_popup_filter")[0];
    loadImages("", popupFilter.value, function(items: any[]) {
      loadedContainer.textContent = "";
      items.forEach(function (item){
        var img: HTMLElement = itemToImg(item);
        img.addEventListener("click", function(e) {
          var currentTarget = <HTMLElement>e.currentTarget
          var cloned = <HTMLElement>currentTarget.cloneNode(true);
          cloned.addEventListener("click", function (event) {
            var currentTarget = <HTMLElement>event.currentTarget
            currentTarget.remove();
          });
          bindDraggableEvents(cloned);
          selectedContainer.appendChild(cloned);
        });
        loadedContainer.appendChild(img);
      });
    });
  }

  var draggedElement: HTMLElement;

  function bindDraggableEvents(el: HTMLElement) {
    el.addEventListener("dragstart", function() {
      draggedElement = this;
    });

    el.addEventListener("drop", function(e) {
      if (this != draggedElement) {
        var uid = this.getAttribute("data-uid");
        var src = this.getAttribute("src");

        this.setAttribute("data-uid", draggedElement.getAttribute("data-uid"));
        this.setAttribute("src", draggedElement.getAttribute("src"));

        draggedElement.setAttribute("data-uid", uid);
        draggedElement.setAttribute("src", src);
      }
    });

    el.addEventListener("dragover", function(e) {
      e.preventDefault();
    });
  }

  function itemToImg(item: any) {
    var img = document.createElement("img");
    img.classList.add("admin_images_img");
    img.setAttribute("src", item.Thumb)
    img.setAttribute("data-uid", item.UID)
    return img;
  }

  function createDraggableImg(item: any) {
    var img = itemToImg(item);
    img.setAttribute("draggable", "true");
    bindDraggableEvents(img);
    img.addEventListener("click", function(event) {
      var element = <HTMLElement>event.currentTarget;
      element.remove()
    })
    return img;
  }

  function loadImageToPopup(value: string) {
    if (value.length > 0) {
      selectedContainer.textContent = "Loading...";
      loadImages(value, "", function(items: any[]) {
        selectedContainer.textContent = "";
        items.forEach(function (item){
          var img = createDraggableImg(item);
          selectedContainer.appendChild(img);
        });
      });
    }
  }

  var connectedItem: HTMLElement;

  popup.getElementsByClassName("admin_images_popup_save")[0].addEventListener("click", function() {
    hidePopup();
    var items: any[] = [];

    for (var i = 0; i < selectedContainer.children.length; i++) {
      items.push(selectedContainer.children[i].getAttribute("data-uid"));
    }

    var str = items.join(",");
    connectedItem.getElementsByTagName("input")[0].value = str
    showPreview(connectedItem);
  });

  popup.getElementsByClassName("admin_images_popup_cancel")[0].addEventListener("click", hidePopup);
  popup.getElementsByClassName("admin_images_popup_filter_button")[0].addEventListener("click", doFilter);

  function showPopup(el: HTMLElement) {
    connectedItem = el;
    var val: string = el.getElementsByTagName("input")[0].value
    loadImageToPopup(val);
    var focusable = <HTMLElement>document.getElementsByClassName("admin_images_popup_box")[0];
    focusable.focus();
    popup.style.display = "block";
    doFilter();
  }

  function hidePopup() {
    popup.style.display = "none";
  }

  function loadImages(ids: string, q: string, handler: (any)) {
    var url: string = adminPrefix + "/_api/image/list?";

    if (ids.length > 0) {
      url += "ids=" + encodeURIComponent(ids);
    } else {
      url += "q=" + encodeURIComponent(q);
    }

    var request = new XMLHttpRequest();
    request.open("GET", url, true);

    request.onload = function() {
        if (this.status == 200) {
          handler(JSON.parse(this.response))
        } else {
          console.error("Error while loading images.");
        }
    }
    request.send();
  }

  function bindImage(el: HTMLElement) {
    showPreview(el);
    el.getElementsByClassName("admin_images_edit")[0].addEventListener("click", function() {
      showPopup(el);
      return false;
    });
  }

  function showPreview(el: HTMLElement) {
    var value = el.getElementsByTagName("input")[0].value;
    var list = el.getElementsByClassName("admin_images_list")[0];
    list.textContent = "";
    (<HTMLElement>el.getElementsByClassName("admin_images_edit")[0]).style.display = "none";
    (<HTMLElement>el.getElementsByTagName("progress")[0]).style.display = "";
    if (value.length > 0) {
      loadImages(value, "", function(items: any[]) {
        doneLoading(el);
        items.forEach(function (item){
          var link = document.createElement("a");
          link.setAttribute("href", adminPrefix+"/file/"+item.ID);
          link.setAttribute("target", "_blank");

          var img = document.createElement("img");
          img.setAttribute("src", item.Thumb);
          img.classList.add("admin_images_img");

          link.appendChild(img);
          list.appendChild(link);
        });
      });
    } else {
      doneLoading(el);
    }
  }

  function doneLoading(el: HTMLElement) {
    el.getElementsByClassName("admin_images_list")[0].textContent = "";
    (<HTMLElement>el.getElementsByClassName("admin_images_edit")[0]).style.display = "";
    el.getElementsByTagName("progress")[0].style.display = "none";
  }

  function showLoadedResult(text: string) {
    document.getElementsByClassName("admin_images_popup_box_upload_message")[0].textContent = text;
    (<HTMLElement>document.querySelector("admin_images_popup_box_upload_btn")).style.display = "";
    (<HTMLElement>document.querySelector("admin_images_popup_box_upload input")).style.display = "";
  }

  document.getElementsByClassName("admin_images_popup_box_upload_btn")[0].addEventListener("click", function (e) {
    var filesInput = <HTMLInputElement>document.querySelector(".admin_images_popup_box_upload input");
    var files = filesInput.files
    var data = new FormData();
    Array.prototype.forEach.call(files, function(item: any, i: number){
      data.append("file", item);
    })

    data.append("description", (<HTMLInputElement>document.getElementsByClassName("admin_popup_file_description")[0]).value);
    (<HTMLInputElement>document.getElementsByClassName("admin_popup_file_description")[0]).value = "";


    document.getElementsByClassName("admin_images_popup_box_upload_message")[0].textContent = "Uploading...";
    (<HTMLElement>document.getElementsByClassName("admin_images_popup_box_upload_btn")[0]).style.display = "none";
    (<HTMLElement>document.querySelector("admin_images_popup_box_upload input")).style.display = "none";

    var request = new XMLHttpRequest();
    request.setRequestHeader('Content-Type', 'multipart/form-data');
    request.open("POST", adminPrefix + "/_api/image/upload");

    request.onload = function() {
      if (this.status == 200) {
        var items: any[] = JSON.parse(this.response);
        Array.prototype.forEach.call(items, function(item: any, i: number) {
          var img = createDraggableImg(item);
          selectedContainer.appendChild(img);
        })
        showLoadedResult("Uploaded successfully.");
      } else {
        showLoadedResult("Error while uploading files.");
      }
    }
    request.send(data);

    /*$.ajax({
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
    });*/
  });

  var elements = document.querySelectorAll(".admin_images");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    bindImage(el);
  });
}
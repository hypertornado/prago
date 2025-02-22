class ImagePicker {
  el: HTMLDivElement;
  preview: HTMLDivElement;
  hiddenInput: HTMLInputElement;
  fileInput: HTMLInputElement;
  progress: HTMLProgressElement;
  draggedElement: HTMLAnchorElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    this.hiddenInput = <HTMLInputElement>(
      el.querySelector(".admin_images_hidden")
    );
    this.preview = <HTMLDivElement>el.querySelector(".admin_images_preview");
    this.fileInput = <HTMLInputElement>(
      this.el.querySelector(".admin_images_fileinput input")
    );
    this.progress = <HTMLProgressElement>this.el.querySelector("progress");

    this.el.querySelector(".admin_images_loaded").classList.remove("hidden");
    this.hideProgress();

    let fileResponses = JSON.parse(
      this.hiddenInput.getAttribute("data-file-responses")
    );
    this.addFiles(fileResponses);

    this.el.addEventListener("click", (e) => {
      if (e.altKey) {
        var ids = window.prompt("IDs of images", this.hiddenInput.value);
        this.hiddenInput.value = ids;
        e.preventDefault();
        return false;
      }
    });

    this.fileInput.addEventListener("dragenter", (ev) => {
      this.fileInput.classList.add("admin_images_fileinput-droparea");
    });

    this.fileInput.addEventListener("dragleave", (ev) => {
      this.fileInput.classList.remove("admin_images_fileinput-droparea");
    });

    this.fileInput.addEventListener("dragover", (ev) => {
      ev.preventDefault();
    });

    this.fileInput.addEventListener("drop", (ev) => {
      var text = ev.dataTransfer.getData("Text");
      return;
    });

    /*for (var i = 0; i < ids.length; i++) {
      var id = ids[i];
      if (id) {
        this.addImage(id);
      }
    }*/

    this.fileInput.addEventListener("change", (e) => {
      var files = this.fileInput.files;
      var formData = new FormData();
      if (files.length == 0) {
        return;
      }

      for (var i = 0; i < files.length; i++) {
        formData.append("file", files[i]);
      }

      var request = new XMLHttpRequest();
      request.open("POST", "/admin/file/api/upload");

      request.addEventListener("load", (e) => {
        this.hideProgress();
        if (request.status == 200) {
          var data = JSON.parse(request.response);
          this.addFiles(data);
        } else {
          new Alert("Chyba při nahrávání souboru.");
          console.error("Error while loading item.");
        }
      });

      this.showProgress();
      request.send(formData);
    });
  }

  updateHiddenData() {
    var ids: any[] = [];
    for (var i = 0; i < this.preview.children.length; i++) {
      var item = <HTMLDivElement>this.preview.children[i];
      var uuid = item.getAttribute("data-uuid");
      ids.push(uuid);
    }
    this.hiddenInput.value = ids.join(",");
  }

  addFiles(files: any) {
    for (var i = 0; i < files.length; i++) {
      let file = files[i];
      this.addFile(file);
    }
  }

  addFile(file: any) {
    var container = document.createElement("a");
    container.classList.add("admin_images_image");
    container.setAttribute("data-uuid", file.UUID);
    container.setAttribute("draggable", "true");
    container.setAttribute("target", "_blank");
    container.setAttribute("href", file.FileURL);
    container.setAttribute(
      "style",
      "background-image: url('" + file.ThumbnailURL + "');"
    );

    var descriptionEl = document.createElement("div");
    descriptionEl.classList.add("admin_images_image_description");
    container.appendChild(descriptionEl);

    descriptionEl.innerText = file.Name;
    container.setAttribute("title", file.Name);

    container.addEventListener("dragstart", (e) => {
      this.draggedElement = <HTMLAnchorElement>e.target;
      //(e as DragEvent).dataTransfer.setData('text/plain', '');
    });

    container.addEventListener("drop", (e) => {
      //@ts-ignore
      var droppedElement: Element = e.toElement;

      //firefox dont have toElement, but have originalTarget
      if (!droppedElement) {
        droppedElement = <Element>(<any>e).originalTarget;
      }

      for (var i = 0; i < 3; i++) {
        if (droppedElement.nodeName == "A") {
          break;
        } else {
          droppedElement = (<HTMLElement>droppedElement).parentElement;
        }
      }

      var draggedIndex: number = -1;
      var droppedIndex: number = -1;
      var parent = this.draggedElement.parentElement;

      for (var i = 0; i < parent.children.length; i++) {
        var child = parent.children[i];
        if (child == this.draggedElement) {
          draggedIndex = i;
        }
        if (child == droppedElement) {
          droppedIndex = i;
        }
      }
      if (draggedIndex == -1 || droppedIndex == -1) {
        return;
      }

      if (draggedIndex <= droppedIndex) {
        droppedIndex += 1;
      }

      DOMinsertChildAtIndex(parent, this.draggedElement, droppedIndex);
      this.updateHiddenData();

      e.preventDefault();
      return false;
    });

    container.addEventListener("dragover", (e) => {
      e.preventDefault();
    });

    container.addEventListener("click", (e) => {
      var target = <HTMLDivElement>e.target;
      if (target.classList.contains("admin_images_image_delete")) {
        var parent = (<HTMLDivElement>e.currentTarget).parentNode;
        parent.removeChild(<HTMLDivElement>e.currentTarget);
        this.updateHiddenData();
        e.preventDefault();
        return false;
      }
    });

    var del = document.createElement("div");
    del.textContent = "×";
    del.classList.add("admin_images_image_delete");
    container.appendChild(del);

    this.preview.appendChild(container);
    this.updateHiddenData();
  }

  hideProgress() {
    this.progress.classList.add("hidden");
  }

  showProgress() {
    this.progress.classList.remove("hidden");
  }
}

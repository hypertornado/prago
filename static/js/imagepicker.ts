class ImagePicker {
  el: HTMLDivElement;
  preview: HTMLDivElement;
  preview2: HTMLDivElement;
  hiddenInput: HTMLInputElement;
  fileInput: HTMLInputElement;
  progress: HTMLProgressElement;
  draggedElement: HTMLAnchorElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    this.hiddenInput = <HTMLInputElement>(
      el.querySelector(".admin_images_hidden")
    );
    this.preview2 = <HTMLDivElement>el.querySelector(".imagepicker_preview");
    this.fileInput = <HTMLInputElement>(
      this.el.querySelector(".imagepicker_btn input")
    );
    this.progress = <HTMLProgressElement>this.el.querySelector("progress");

    this.el.querySelector(".imagepicker_content").classList.remove("hidden");
    this.hideProgress();

    this.el.addEventListener("click", (e) => {
      if (e.altKey) {
        var ids = window.prompt("IDs of images", this.hiddenInput.value);
        this.hiddenInput.value = ids;
        this.load();
        e.preventDefault();
        return false;
      }
    });

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
          console.log(data);
          for (var i = 0; i < data.length; i++) {
            console.log(data[i]);
            this.addUUID(data[i]);
          }
        } else {
          new Alert("Chyba při nahrávání souboru.");
          console.error("Error while loading item.");
        }
      });

      this.showProgress();
      request.send(formData);
    });

    this.load();
  }

  hideProgress() {
    this.progress.classList.add("hidden");
  }

  showProgress() {
    this.progress.classList.remove("hidden");
  }

  load() {
    this.showProgress();

    var request = new XMLHttpRequest();
    request.open("GET", "/admin/api/imagepicker"+encodeParams({
      "ids": this.hiddenInput.value,
    }));

    request.addEventListener("load", (e) => {
      this.hideProgress();
      if (request.status == 200) {
        var data = JSON.parse(request.response);
        this.addFiles2(data);
      } else {
        new Alert("Chyba při načítání dat obrázků.");
        console.error("Error while loading item.");
      }
    });
    request.send();
  }

  addFiles2(data: any) {
    this.preview2.innerHTML = "";

    for (let i = 0; i < data.Items.length; i++) {
      let item = data.Items[i];
      
      let itemEl = document.createElement("button");
      itemEl.setAttribute("data-uuid", item.UUID);
      itemEl.setAttribute("title", item.ImageName);
      itemEl.classList.add("imagepicker_preview_item");
      itemEl.setAttribute("style", "background-image: url('" + item.ThumbURL +"');")


      itemEl.addEventListener("click", (e) => {
        e.stopPropagation();
        e.preventDefault();

        var commands: CMenuCommand[] = [];
        commands.push({
          Name: "Zobrazit",
          Icon: "glyphicons-basic-588-book-open-text.svg",
          URL: item.ViewURL,
        });
        commands.push({
          Name: "Upravit popis",
          Icon: "glyphicons-basic-31-pencil.svg",
          //URL: item.EditURL,
          Handler: () => {
            new PopupForm(item.EditURL, () => {
              this.load();
            })
          }
        });
        commands.push({
          Name: "První",
          Icon: "glyphicons-basic-212-arrow-up.svg",
          Handler: () => {
            DOMinsertChildAtIndex(this.preview2, itemEl, 0);
            this.updateHiddenData2();
            this.load();
          },
        });
        commands.push({
          Name: "Nahoru",
          Icon: "glyphicons-basic-828-arrow-thin-up.svg",
          Handler: () => {
            DOMinsertChildAtIndex(this.preview2, itemEl, i-1);
            this.updateHiddenData2();
            this.load();
          },
        });
        commands.push({
          Name: "Dolů",
          Icon: "glyphicons-basic-827-arrow-thin-down.svg",
          Handler: () => {
            DOMinsertChildAtIndex(this.preview2, itemEl, i+2);
            this.updateHiddenData2();
            this.load();
          },
        });
        commands.push({
          Name: "Smazat",
          Handler: () => {
            itemEl.remove();
            this.updateHiddenData2();
          },
          Icon: "glyphicons-basic-17-bin.svg",
          Style: "destroy",
        });

        var rows: CMenuTableRow[] = [];
        for (var j = 0; j < item.Metadata.length; j++) {
          rows.push({
            Name: item.Metadata[j][0],
            Value: item.Metadata[j][1],
          })
        }

        cmenu({
          Event: e,
          AlignByElement: true,
          Name: item.ImageName,
          Description: item.ImageDescription,
          Commands: commands,
          Rows: rows,
        })

      })

      this.preview2.appendChild(itemEl);
    }

  }

  updateHiddenData2() {
    var ids: any[] = [];
    for (var i = 0; i < this.preview2.children.length; i++) {
      let item = <HTMLDivElement>this.preview2.children[i];
      var uuid = item.getAttribute("data-uuid");
      ids.push(uuid);
    }
    this.hiddenInput.value = ids.join(",");
  }
  

  addUUID(uuid: string) {
    if (!uuid) {
      return;
    }
    let val = this.hiddenInput.value;
    if (val) {
      val += ",";
    }
    val += uuid;
    this.hiddenInput.value = val;
    this.load();
  }


}

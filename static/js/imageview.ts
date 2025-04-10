class ImageView {
  el: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    var filesData = JSON.parse(el.getAttribute("data-images"));
    this.addFiles(filesData);
  }

  addFiles(filesData: any) {
    this.el.innerHTML = "";
    if (!filesData.Items) {
      return;
    }

    for (var i = 0; i < filesData.Items.length; i++) {
      let file = filesData.Items[i];
      this.addFile(file);
    }
  }

  addFile(file: any) {
    let container = document.createElement("button");
    container.classList.add("imageview_image");
    container.setAttribute("href", file.ViewURL);
    container.setAttribute("title", file.ImageDescription);

    let imgEl = document.createElement("img");
    imgEl.classList.add("imageview_image_img");
    imgEl.setAttribute("src", file.ThumbURL)
    container.appendChild(imgEl);

    container.addEventListener("click", (e: PointerEvent) => {
      e.preventDefault();
      e.stopPropagation();

      let commands = [];
      commands.push({
        Name: "Zobrazit",
        URL: file.ViewURL,
      });

      commands.push({
        Name: "Kopírovat UUID",
        Handler: () => {
          navigator.clipboard.writeText(file.UUID);
          Prago.notificationCenter.flashNotification(
            "Zkopírováno",
            null,
            true,
            false
          );
        },
        Icon: "glyphicons-basic-611-copy-duplicate.svg",
      });

      cmenu({
        Event: e,
        AlignByElement: true,
        Name: file.ImageName,
        Description: file.ImageDescription,
        Commands: commands,
        Rows: CMenu.rowsFromArray(file.Metadata),
      });

    })
    
    
    
    this.el.appendChild(container);
  }
}

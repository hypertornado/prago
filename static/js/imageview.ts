class ImageView {
  el: HTMLDivElement;
  galleryImagesData: any[];

  constructor(el: HTMLDivElement) {
    this.galleryImagesData = [];
    this.el = el;
    var filesData = JSON.parse(el.getAttribute("data-images"));
    this.addFiles(filesData);
    console.log("done");
  }

  addFiles(filesData: any) {
    this.el.innerHTML = "";
    if (!filesData.Items) {
      return;
    }

    for (var i = 0; i < filesData.Items.length; i++) {
      let file = filesData.Items[i];
      this.addFile(file, i);
    }
  }

  addFile(file: any, index: number) {
    let container = document.createElement("button");
    container.setAttribute("type", "button");
    container.classList.add("imageview_image");
    container.setAttribute("href", file.ViewURL);
    container.setAttribute("title", file.ImageDescription);

    let imgEl = document.createElement("img");
    imgEl.classList.add("imageview_image_img");
    imgEl.setAttribute("src", file.ThumbURL)
    container.appendChild(imgEl);

    let btnEl = document.createElement("div");
    btnEl.classList.add("btn");
    btnEl.classList.add("imageview_image_btn");
    btnEl.innerText = "…";
    container.appendChild(btnEl);

    this.galleryImagesData.push({"URL": file.GiantURL, "Title": file.ImageName + " " + file.ImageDescription})

    container.addEventListener("click", (e: PointerEvent) => {
      e.preventDefault();
      e.stopPropagation();
      //@ts-ignore
      new PragoPhotoGallery(this.galleryImagesData, {"index": index});
    });


    btnEl.addEventListener("click", (e: PointerEvent) => {
      e.preventDefault();
      e.stopPropagation();

      let commands = [];
      commands.push({
        Name: "Náhled",
        Icon: "glyphicons-basic-52-eye.svg",
        Handler: () => {
          //@ts-ignore
          new PragoPhotoGallery(this.galleryImagesData, {"index": index});
        },
      });
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
        //AlignByElement: true,
        Name: file.ImageName,
        Description: file.ImageDescription,
        Commands: commands,
        Rows: CMenu.rowsFromArray(file.Metadata),
      });

    })
    
    
    
    this.el.appendChild(container);
  }
}

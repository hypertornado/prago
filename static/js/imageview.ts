class ImageView {
  el: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    var filesData = JSON.parse(el.getAttribute("data-images"));
    this.addFiles(filesData);
  }

  addFiles(filesData: any) {
    this.el.innerHTML = "";
    for (var i = 0; i < filesData.Items.length; i++) {
      let file = filesData.Items[i];
      this.addFile(file);
    }
  }

  addFile(file: any) {
    var container = document.createElement("button");
    container.classList.add("imageview_image");
    container.setAttribute("href", file.ViewURL);
    container.setAttribute(
      "style",
      "background-image: url('" + file.ThumbURL + "');"
    );
    container.setAttribute("title", file.ImageDescription);

    container.addEventListener("click", (e: PointerEvent) => {
      e.preventDefault();
      e.stopPropagation();

      let commands = [];
      commands.push({
        Name: "Zobrazit",
        URL: file.ViewURL,
      })

      cmenu({
        Event: e,
        AlignByElement: true,
        Name: file.ImageName,
        Description: file.ImageDescription,
        Commands: commands,
      });

    })
    
    
    
    this.el.appendChild(container);
  }
}

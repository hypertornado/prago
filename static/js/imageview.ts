class ImageView {
  el: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    var filesData = JSON.parse(el.getAttribute("data-images"));
    this.addFiles(filesData);
  }

  addFiles(filesData: any) {
    this.el.innerHTML = "";
    for (var i = 0; i < filesData.length; i++) {
      let file = filesData[i];
      this.addFile(file);
    }
  }

  addFile(file: any) {
    var container = document.createElement("a");
    container.classList.add("admin_images_image");
    container.setAttribute("href", file.FileURL);
    container.setAttribute(
      "style",
      "background-image: url('" + file.ThumbnailURL + "');"
    );

    var img = document.createElement("div");
    img.setAttribute("src", file.ThumbnailURL);
    img.setAttribute("draggable", "false");

    var descriptionEl = document.createElement("div");
    descriptionEl.classList.add("admin_images_image_description");
    container.appendChild(descriptionEl);

    descriptionEl.innerText = file.Name;
    container.setAttribute("title", file.Name);
    this.el.appendChild(container);
  }
}

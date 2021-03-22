function bindImageViews() {
  var els = document.querySelectorAll(".admin_item_view_image_content");
  for (var i = 0; i < els.length; i++) {
    new ImageView(<HTMLDivElement>els[i]);
  }
}

class ImageView {
  el: HTMLDivElement;
  adminPrefix: string;

  constructor(el: HTMLDivElement) {
    this.adminPrefix = document.body.getAttribute("data-admin-prefix");
    this.el = el;
    var ids = el.getAttribute("data-images").split(",");
    this.addImages(ids);
  }

  addImages(ids: any) {
    this.el.innerHTML = "";
    for (var i = 0; i < ids.length; i++) {
      if (<string>ids[i] != "") {
        this.addImage(<string>ids[i]);
      }
    }
  }

  addImage(id: string) {
    var container = document.createElement("a");
    container.classList.add("admin_images_image");
    container.setAttribute("href", this.adminPrefix + "/file/api/redirect-uuid/" + id);
    container.setAttribute("style", "background-image: url('" + this.adminPrefix + "/file/api/redirect-thumb/" + id + "');");

    var img = document.createElement("div");
    img.setAttribute("src", this.adminPrefix + "/file/api/redirect-thumb/" + id);
    img.setAttribute("draggable", "false");


    var descriptionEl = document.createElement("div");
    descriptionEl.classList.add("admin_images_image_description")
    container.appendChild(descriptionEl);

    var request = new XMLHttpRequest();
    request.open("GET", this.adminPrefix + "/file/api/imagedata/" + id);
    request.addEventListener("load", (e) => {
      if (request.status == 200) {
        var data = JSON.parse(request.response);
        descriptionEl.innerText = data["Name"];
        container.setAttribute("title", data["Name"])
      } else {
        console.error("Error while loading file metadata.");
      }
    })
    request.send();

    this.el.appendChild(container);
  }

}
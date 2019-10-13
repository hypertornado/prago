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
    container.setAttribute("href", this.adminPrefix + "/file/uuid/" + id);
    container.setAttribute("style", "background-image: url('" + this.adminPrefix + "/_api/image/thumb/" + id + "');");

    var img = document.createElement("div");
    //img.setAttribute("src", this.adminPrefix + "/_api/image/thumb/" + id);
    img.setAttribute("src", this.adminPrefix + "/_api/image/thumb/" + id);
    img.setAttribute("draggable", "false");
    //container.appendChild(img);

    this.el.appendChild(container);
  }

}
class PragoPhotoGallery {
    constructor(data, options) {
        this.gap = 32;
        this.data = data;
        this.el = document.createElement("div");
        this.el.classList.add("prago_photo_gallery");
        this.el.setAttribute("tabindex", "0");
        this.el.innerHTML = `
            <div class="prago_photo_gallery_images"></div>
            <div class="prago_photo_gallery_count"></div>
            <div class="prago_photo_gallery_header">
                <div class="prago_photo_gallery_btn prago_photo_gallery_close">
                    <img src="/admin/api/icons?file=glyphicons-basic-599-menu-close.svg&color=ffffff" class="prago_photo_gallery_btn_icon">
                </div>
            </div>
            <div class="prago_photo_gallery_btn prago_photo_gallery_prev">
                <img src="/admin/api/icons?file=glyphicons-basic-223-chevron-left.svg&color=ffffff" class="prago_photo_gallery_btn_icon">
            </div>
            <div class="prago_photo_gallery_btn prago_photo_gallery_next">
                <img src="/admin/api/icons?file=glyphicons-basic-224-chevron-right.svg&color=ffffff" class="prago_photo_gallery_btn_icon">
            </div>
            <div class="prago_photo_gallery_description prago_photo_gallery_hidden"></div>
        `;
        this.images = this.el.querySelector(".prago_photo_gallery_images");
        this.images.addEventListener("click", (e) => {
            let targetEl = e.target;
            if (targetEl.classList.contains("prago_photo_gallery_image_container")) {
                this.close();
                return;
            }
            if (targetEl.classList.contains("prago_photo_gallery_image")) {
                this.el.classList.toggle("prago_photo_gallery-hiddencontrols");
                return;
            }
        });
        this.descriptionEl = this.el.querySelector(".prago_photo_gallery_description");
        this.countEl = this.el.querySelector(".prago_photo_gallery_count");
        this.nextEl = this.el.querySelector(".prago_photo_gallery_next");
        this.nextEl.addEventListener("click", () => {
            this.next();
        });
        this.prevEl = this.el.querySelector(".prago_photo_gallery_prev");
        this.prevEl.addEventListener("click", () => {
            this.prev();
        });
        this.el.querySelector(".prago_photo_gallery_close").addEventListener("click", (e) => {
            this.close();
            e.preventDefault();
            e.stopPropagation();
        });
        this.el.addEventListener("keydown", (e) => {
            if (e.code == "ArrowLeft") {
                this.prev();
            }
            if (e.code == "ArrowRight") {
                this.next();
            }
            if (e.code == "Escape") {
                this.close();
            }
        });
        this.images.addEventListener("scroll", () => {
            let scrollPosition = this.getScrollPosition();
            this.el.classList.remove("prago_photo_gallery-hiddencontrols");
            this.renderImageDescription(scrollPosition);
        });
        document.body.appendChild(this.el);
        this.renderImages();
        if (options.index) {
            this.setScrollPosition(options.index);
        }
        else {
            this.setScrollPosition(0);
        }
        this.el.focus();
    }
    next() {
        let index = this.getScrollPosition();
        index++;
        this.setScrollPosition(index);
    }
    prev() {
        let index = this.getScrollPosition();
        index--;
        this.setScrollPosition(index);
    }
    getScrollPosition() {
        let scrollLeft = this.images.scrollLeft;
        let viewportWidth = window.innerWidth;
        let currentIndex = Math.round(scrollLeft / (viewportWidth + this.gap));
        return currentIndex;
    }
    setScrollPosition(scrollIndex) {
        this.images.scrollTo({
            left: scrollIndex * (window.innerWidth + this.gap),
            behavior: 'auto'
        });
        this.renderImageDescription(scrollIndex);
        let imgEl = this.images.children[scrollIndex].children[0];
        imgEl.setAttribute("fetchpriority", "high");
    }
    close() {
        this.el.remove();
    }
    imagesCount() {
        return this.data.length;
    }
    renderImages() {
        for (var i = 0; i < this.data.length; i++) {
            let el = this.data[i];
            this.createImage(el);
        }
    }
    createImage(imageData) {
        let imageEl = document.createElement("div");
        imageEl.classList.add("prago_photo_gallery_image_container");
        let imgEl = document.createElement("img");
        imgEl.classList.add("prago_photo_gallery_image");
        imgEl.setAttribute("loading", "lazy");
        imgEl.src = imageData.URL;
        imageEl.appendChild(imgEl);
        this.images.appendChild(imageEl);
    }
    renderImageDescription(index) {
        let imagedata = this.data[index];
        if (imagedata.Title) {
            this.descriptionEl.innerText = imagedata.Title;
            this.descriptionEl.classList.remove("prago_photo_gallery_hidden");
        }
        else {
            this.descriptionEl.classList.add("prago_photo_gallery_hidden");
        }
        if (index == 0) {
            this.prevEl.classList.add("prago_photo_gallery_hidden");
        }
        else {
            this.prevEl.classList.remove("prago_photo_gallery_hidden");
        }
        if (index + 1 >= this.imagesCount()) {
            this.nextEl.classList.add("prago_photo_gallery_hidden");
        }
        else {
            this.nextEl.classList.remove("prago_photo_gallery_hidden");
        }
        this.countEl.innerText = `${index + 1} / ${this.imagesCount()}`;
    }
}

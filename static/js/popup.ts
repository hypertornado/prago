
class Popup {

    el: HTMLDivElement;
    contentEL: HTMLDivElement;
    cancelable: boolean;

    constructor() {
        this.el = document.createElement("div");
        this.el.classList.add("popup_background");
        document.body.appendChild(this.el);

        this.el.innerHTML = `
        <div class="popup">
            <div class="popup_header">
                <div class="popup_header_name"></div>
                <div class="popup_header_cancel">

                </div>
            </div>
            <div class="popup_content"></div>
            <div class="popup_footer"></div>
        </div>
        `

        this.el.querySelector(".popup_header_cancel").addEventListener("click", this.remove.bind(this));
        this.el.addEventListener("click", this.backgroundClicked.bind(this));

        //this.setCancelable();

        //this.addButton("yeah", function () {});
        //this.addButton("Ok", function () {});
    }

    backgroundClicked(e: any) {
        var div = <HTMLDivElement>e.target;
        if (!div.classList.contains("popup_background")) {
            return;
        }
        if (this.cancelable) {
            this.remove();
        }

    }

    remove() {
        this.el.remove();
    }

    setCancelable() {
        this.cancelable = true;
        this.el.querySelector(".popup_header_cancel").classList.add("popup_header_cancel-visible");
        
    }

    setTitle(name: string) {
        this.el.querySelector(".popup_header_name").textContent = name;
    }

    addButton(name: string, handler: any) {
        var button = document.createElement("input");
        button.setAttribute("type", "button");
        button.setAttribute("class", "btn");
        button.setAttribute("value", name);
        button.addEventListener("click", handler);
        this.el.querySelector(".popup_footer").appendChild(button);
    }
}

class Alert extends Popup {
    constructor(text: string) {
        super();
        this.setCancelable();
        this.addButton("OK", this.remove.bind(this));
        this.setTitle(text);
    }
}
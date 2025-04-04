class Popup {
  private el: HTMLDivElement;
  private contentEL: HTMLDivElement;
  private cancelable: boolean;
  protected cancelAction: Function;

  constructor(title: string) {
    this.el = document.createElement("div");
    this.el.classList.add("popup_background");
    document.body.appendChild(this.el);

    this.el.innerHTML = `
        <div class="popup">
            <div class="popup_header">
                <img class="popup_header_icon hidden">
                <div class="popup_header_name"></div>
                <div class="popup_header_cancel"></div>
            </div>
            <div class="popup_content"></div>
            <div class="popup_footer"></div>
        </div>
        `;

    this.el.setAttribute("tabindex", "-1");

    this.el
      .querySelector(".popup_header_cancel")
      .addEventListener("click", this.cancel.bind(this));
    this.el.addEventListener("click", this.backgroundClicked.bind(this));
    this.el.focus();

    this.el.addEventListener("keydown", (e) => {
      if (e.code == "Escape") {
        if (this.cancelable) {
          this.cancel();
        }
      }
    });
    this.setTitle(title);
  }

  private backgroundClicked(e: any) {
    var div = <HTMLDivElement>e.target;
    if (!div.classList.contains("popup_background")) {
      return;
    }
    if (this.cancelable) {
      this.cancel();
    }
  }

  protected wide() {
    this.el.querySelector(".popup").classList.add("popup-wide");
  }

  protected focus() {
    this.el.focus();
  }

  private cancel() {
    if (this.cancelAction) {
      this.cancelAction();
    } else {
      this.remove();
    }
  }

  protected remove() {
    this.el.remove();
  }

  protected setContent(el: HTMLElement) {
    this.el.querySelector(".popup_content").innerHTML = "";
    this.el.querySelector(".popup_content").appendChild(el);
    this.el
      .querySelector(".popup_content")
      .classList.add("popup_content-visible");
  }

  protected setCancelable() {
    this.cancelable = true;
    this.el
      .querySelector(".popup_header_cancel")
      .classList.add("popup_header_cancel-visible");
  }

  protected setTitle(name: string) {
    this.el.querySelector(".popup_header_name").textContent = name;
  }

  public setIcon(iconName: string) {
    if (!iconName) {
      return;
    }
    let iconEl = this.el.querySelector(".popup_header_icon");
    iconEl.classList.remove("hidden");
    iconEl.setAttribute("src", `/admin/api/icons?file=${iconName}&color=444444`)
  }

  protected addButton(
    name: string,
    handler: any,
    style?: ButtonStyle
  ): HTMLInputElement {
    this.el
      .querySelector(".popup_footer")
      .classList.add("popup_footer-visible");

    var button = document.createElement("input");
    button.setAttribute("type", "button");
    button.setAttribute("class", "btn");

    switch (style) {
      case ButtonStyle.Accented:
        button.classList.add("btn-accented");
        break;
      case ButtonStyle.Delete:
        button.classList.add("btn-delete");
        break;
    }
    button.setAttribute("value", name);
    button.addEventListener("click", handler);
    this.el.querySelector(".popup_footer").appendChild(button);
    return button;
  }

  protected present() {
    document.body.appendChild(this.el);
    this.focus();
    this.el.classList.add("popup_background-presented");
  }

  protected unpresent() {
    this.el.classList.remove("popup_background-presented");
  }
}

enum ButtonStyle {
  Default,
  Accented,
  Delete,
}

class Alert extends Popup {
  constructor(title: string) {
    super(title);
    this.setCancelable();
    this.present();
    this.setIcon("glyphicons-basic-79-triangle-empty-alert.svg");
    this.addButton("OK", this.remove.bind(this), ButtonStyle.Accented).focus();
  }
}

class Confirm extends Popup {
  private primaryButton: HTMLInputElement;

  constructor(
    title: string,
    buttonName: string,
    handlerConfirm?: Function,
    handlerCancel?: Function,
    style?: ButtonStyle
  ) {
    super(title);
    this.setCancelable();
    if (!style) {
      style = ButtonStyle.Accented;
    }
    this.cancelAction = () => {
      this.remove();
      if (handlerCancel) {
        handlerCancel();
      }
    };
    this.addButton("Storno", () => {
      this.cancelAction();
    });

    var primaryText = buttonName;
    if (!primaryText) {
      primaryText = "OK";
    }

    this.primaryButton = this.addButton(
      primaryText,
      () => {
        this.remove();
        if (handlerConfirm) {
          handlerConfirm();
        }
      },
      style
    );
    this.present();
    this.primaryButton.focus();
  }
}

class ContentPopup extends Popup {
  isShown: boolean;

  constructor(title: string, content?: HTMLElement) {
    super(title);
    this.isShown = false;
    this.setCancelable();
    if (content) this.setContent(content);
    this.wide();
    this.cancelAction = this.hide.bind(this);
  }

  show() {
    this.present();
    this.focus();
    this.isShown = true;
  }

  hide() {
    this.unpresent();
    this.isShown = false;
  }

  setHiddenHandler(handler: any) {
    this.cancelAction = () => {
      handler();
      this.remove();
    };
  }

  setContent(content: HTMLElement) {
    super.setContent(content);
  }

  setConfirmButtons(handler: any) {
    super.addButton("Storno", () => {
      super.unpresent();
    });
    super.addButton("Upravit", handler, ButtonStyle.Accented);
  }
}

class LoadingPopup extends Popup {
  constructor() {
    super("");

    var contentEl = document.createElement("div");
    contentEl.innerHTML = '<progress class="progress"></progress>';
    this.setContent(contentEl);
    this.present();
  }

  done() {
    this.remove();
  }
}

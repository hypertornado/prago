interface CMenuData {
  Event: Event;
  ImageURL?: string;
  Name?: string;
  Description?: string;

  Commands?: CMenuCommand[];

  DismissHandler?: Function;
}

interface CMenuCommand {
  Icon?: string;
  Name: string;
  Handler: Function;
}

function cmenu(data: CMenuData) {
  Prago.cmenu.showWithData(data);
}

class CMenu {
  lastEl: HTMLDivElement;
  dismissHandler: Function;

  constructor() {
    for (let eventType of ["keydown", "click", "visibilitychange", "blur"]) {
      document.addEventListener(eventType, (e) => {
        this.dismiss();
      });
    }
  }

  dismiss() {
    if (this.lastEl) {
      this.lastEl.remove();
      this.lastEl = null;
    }
    if (this.dismissHandler) {
      this.dismissHandler();
      this.dismissHandler = null;
    }
  }

  showWithData(data: CMenuData) {
    this.dismiss();

    //@ts-ignore
    let y = data.Event.clientY;
    //@ts-ignore
    let x = data.Event.clientX;

    let el = document.createElement("div");
    el.classList.add("cmenu");

    if (data.ImageURL) {
      let imageEl = document.createElement("img");
      imageEl.classList.add("cmenu_image");
      imageEl.setAttribute("src", data.ImageURL);
      el.appendChild(imageEl);
    }

    if (data.Name) {
      let nameEl = document.createElement("div");
      nameEl.classList.add("cmenu_name");
      nameEl.innerText = data.Name;
      el.appendChild(nameEl);
    }

    if (data.Description) {
      let descEl = document.createElement("div");
      descEl.classList.add("cmenu_description");
      descEl.innerText = data.Description;
      el.appendChild(descEl);
    }

    if (data.Commands) {
      let commandsEl = document.createElement("div");
      commandsEl.classList.add("cmenu_commands");

      for (let command of data.Commands) {
        let commandEl = document.createElement("div");
        commandEl.classList.add("cmenu_command");

        let commandNameEl = document.createElement("div");
        commandNameEl.classList.add("cmenu_command_name");
        commandNameEl.innerText = command.Name;
        commandEl.appendChild(commandNameEl);

        if (command.Icon) {
          let commandNameIcon = document.createElement("img");
          commandNameIcon.classList.add("cmenu_command_icon");
          commandNameIcon.setAttribute(
            "src",
            "/admin/api/icons?file=" + command.Icon + "&color=4077bf"
          );
          commandEl.appendChild(commandNameIcon);
        }

        commandEl.addEventListener("click", (e) => {
          command.Handler();
        });
        commandsEl.appendChild(commandEl);
      }

      el.appendChild(commandsEl);
    }

    document.body.appendChild(el);

    let elRect = el.getBoundingClientRect();
    let elWidth = elRect.width;
    let elHeight = elRect.height;

    let viewportWidth = window.innerWidth;
    let viewportHeight = window.innerHeight;

    console.log(x, elWidth, viewportWidth);

    if (x + elWidth > viewportWidth) {
      x = viewportWidth - elWidth;
    }

    if (y + elHeight > viewportHeight) {
      y = viewportHeight - elHeight;
    }

    el.style.left = x + "px";
    el.style.top = y + "px";

    this.lastEl = el;
    this.dismissHandler = data.DismissHandler;
  }
}

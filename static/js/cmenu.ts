interface CMenuData {
  Event: Event;
  AlignByElement?: boolean;
  ImageURL?: string;
  PreName?: string;
  Name?: string;
  Description?: string;

  Commands?: CMenuCommand[];
  Rows?: CMenuTableRow[];

  DismissHandler?: Function;
}

interface CMenuCommand {
  Icon?: string;
  Name: string;
  URL?: string;
  Handler?: Function;
  Style?: string;
}

interface CMenuTableRow {
  Name: string;
  Value: string;
}

function cmenu(data: CMenuData) {
  Prago.cmenu.showWithData(data);
}

class CMenu {
  lastEl: HTMLDivElement;
  dismissHandler: Function;

  constructor() {
    for (let eventType of ["click", "visibilitychange", "blur"]) {
      document.addEventListener(eventType, (e) => {
        this.dismiss();
      });
    }

    document.addEventListener("keydown", (e: KeyboardEvent) => {
      if (e.key == "Escape") {
        this.dismiss();
      }
    });

  }

  static rowsFromArray(inArr: []): CMenuTableRow[] {
    var rows: CMenuTableRow[] = [];
    for (var j = 0; j < inArr.length; j++) {
      rows.push({
        Name: inArr[j][0],
        Value: inArr[j][1],
      })
    }
    return rows;
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

    let containerEl = document.createElement("div");
    containerEl.classList.add("cmenu_container");
    containerEl.addEventListener("contextmenu", (e) => {
      e.preventDefault();
    });

    let el = document.createElement("div");
    el.classList.add("cmenu");

    containerEl.appendChild(el);

    if (data.ImageURL) {
      let imageEl = document.createElement("img");
      imageEl.classList.add("cmenu_image");
      imageEl.setAttribute("src", data.ImageURL);
      el.appendChild(imageEl);
    }

    if (data.PreName) {
      let preNameEl = document.createElement("div");
      preNameEl.classList.add("cmenu_prename");
      preNameEl.innerText = data.PreName;
      preNameEl.setAttribute("title", data.PreName);
      el.appendChild(preNameEl);
    }

    if (data.Name) {
      let nameEl = document.createElement("div");
      nameEl.classList.add("cmenu_name");
      nameEl.innerText = data.Name;
      nameEl.setAttribute("title", data.Name);
      el.appendChild(nameEl);
    }

    if (data.Description) {
      let descEl = document.createElement("div");
      descEl.classList.add("cmenu_description");
      descEl.innerText = data.Description;
      descEl.setAttribute("title", data.Description);
      el.appendChild(descEl);
    }

    if (data.Rows) {
      let rowsEl = document.createElement("div");
      rowsEl.classList.add("cmenu_table_rows");

      for (let i = 0; i < data.Rows.length; i++) {
        let row = data.Rows[i];

        let rowEl = document.createElement("div");
        rowEl.classList.add("cmenu_table_row");
        
        let rowNameEl = document.createElement("div");
        rowNameEl.classList.add("cmenu_table_row_name");
        rowNameEl.innerText = row.Name;
        rowEl.appendChild(rowNameEl);

        let rowValueEl = document.createElement("div");
        rowValueEl.classList.add("cmenu_table_row_value");
        rowValueEl.innerText = row.Value;
        rowEl.appendChild(rowValueEl);

        rowsEl.appendChild(rowEl);
      }

      el.appendChild(rowsEl);
    }

    if (data.Commands) {
      let commandsEl = document.createElement("div");
      commandsEl.classList.add("cmenu_commands");

      for (let command of data.Commands) {
        let commandEl = document.createElement("div");
        commandEl.classList.add("cmenu_command");

        if (command.Style) {
          commandEl.classList.add("cmenu_command-" + command.Style);
        }

        let commandNameEl = document.createElement("div");
        commandNameEl.classList.add("cmenu_command_name");
        commandNameEl.innerText = command.Name;
        commandEl.appendChild(commandNameEl);

        if (command.Icon) {
          let commandNameIcon = document.createElement("img");
          commandNameIcon.classList.add("cmenu_command_icon");
          let color = "4077bf";
          if (command.Style == "destroy") {
            color = "cb2431";
          }
          commandNameIcon.setAttribute(
            "src",
            "/admin/api/icons?file=" + command.Icon + "&color=" + color,
          );
          commandEl.appendChild(commandNameIcon);
        }

        commandEl.addEventListener("click", (e: MouseEvent) => {
          if (command.URL) {
            if (e.shiftKey || e.metaKey || e.ctrlKey) {
              var openedWindow = window.open(command.URL, "newwindow" + new Date() + Math.random());
              openedWindow.focus();
            } else {
              window.location.href = command.URL;
            }
          }
          if (command.Handler) {
            command.Handler();
          }
          this.dismiss();
        });
        commandsEl.appendChild(commandEl);
      }

      el.appendChild(commandsEl);
    }

    document.body.appendChild(containerEl);

    let elWidth = el.clientWidth;
    let elHeight = el.clientHeight;

    let viewportWidth = window.innerWidth;
    let viewportHeight = window.innerHeight;

    if (data.AlignByElement) {
      let targetEl = <HTMLDivElement>data.Event.currentTarget;
      let rect = targetEl.getBoundingClientRect();

      x = rect.left;
      y = rect.top + rect.height;

      if (x + elWidth > viewportWidth) {
        if (x > viewportWidth / 2) {
          x = rect.x + rect.width - elWidth;
        }
      }

      if (y + elHeight > viewportHeight) {
        if (y > viewportHeight / 2) {
          y = rect.y - elHeight;
        }
      }

      if (x < 0) {
        x = 0;
      }

      if (y < 0) {
        y = 0;
      }
    } else {
      if (x + elWidth > viewportWidth) {
        x = viewportWidth - elWidth;
      }

      if (y + elHeight > viewportHeight) {
        y = viewportHeight - elHeight;
      }
    }

    el.style.left = x + "px";
    el.style.top = y + "px";

    el.addEventListener("click", (e: KeyboardEvent) => {
      e.stopPropagation();
    })


    this.lastEl = containerEl;
    this.dismissHandler = data.DismissHandler;
  }

  getOffset(el: HTMLDivElement) {
    var _x = 0;
    var _y = 0;
    while (el && !isNaN(el.offsetLeft) && !isNaN(el.offsetTop)) {
      _x += el.offsetLeft - el.scrollLeft;
      _y += el.offsetTop - el.scrollTop;
      el = <HTMLDivElement>el.offsetParent;
    }
    return { top: _y, left: _x };
  }
}

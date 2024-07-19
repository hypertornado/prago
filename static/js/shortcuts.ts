class Shortcuts {
  private el: HTMLElement;
  private shortcuts: Shortcut[];

  constructor(el: HTMLElement) {
    this.el = el;
    this.shortcuts = [];

    this.el.addEventListener("keydown", (e) => {
      //console.log(e);
      for (let shortcut of this.shortcuts) {
        if (shortcut.match(e)) {
          shortcut.handler();
          e.preventDefault();
          e.stopPropagation();
          return false;
        }
      }
    });
  }

  add(shortcut: ShortcutKeys, description: string, handler: Function) {
    this.shortcuts.push(new Shortcut(shortcut, description, handler));
  }

  addRootShortcuts() {
    let popup = new ContentPopup("Zkratky");
    this.add(
      {
        Key: "?",
      },
      "Zobrazit nápovědu",
      () => {
        let contentEl = document.createElement("div");
        for (let shortcut of this.shortcuts) {
          let shortcutEl = document.createElement("div");
          shortcutEl.innerText = shortcut.getDescription();
          contentEl.appendChild(shortcutEl);
        }

        if (popup.isShown) {
          popup.hide();
        } else {
          popup.setContent(contentEl);
          popup.show();
        }
      }
    );
  }
}

interface ShortcutKeys {
  Key: string;
  Shift?: boolean;
  Control?: boolean;
  Alt?: boolean;
}

class Shortcut {
  private shortcut: ShortcutKeys;
  handler: Function;
  private description: string;

  constructor(shortcut: ShortcutKeys, description: string, handler: Function) {
    this.shortcut = shortcut;
    this.handler = handler;
    this.description = description;
  }

  match(e: KeyboardEvent): boolean {
    if (e.key != this.shortcut.Key) {
      return false;
    }

    if (this.shortcut.Alt && !e.altKey) {
      return false;
    }

    if (this.shortcut.Shift && !e.shiftKey) {
      return false;
    }

    if (this.shortcut.Control && !e.ctrlKey && !e.metaKey) {
      return false;
    }

    return true;
  }

  getDescription(): string {
    let items: string[] = [];
    if (this.shortcut.Control) {
      items.push("Ctrl");
    }
    if (this.shortcut.Alt) {
      items.push("Alt");
    }
    if (this.shortcut.Shift) {
      items.push("Shift");
    }
    items.push(this.shortcut.Key);
    return items.join("+") + ": " + this.description;
  }
}

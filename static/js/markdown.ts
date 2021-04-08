function bindMarkdowns() {
  var elements = document.querySelectorAll(".admin_markdown");
  Array.prototype.forEach.call(elements, function (el: HTMLElement, i: number) {
    new MarkdownEditor(el);
  });
}

class MarkdownEditor {
  textarea: HTMLTextAreaElement;
  preview: HTMLDivElement;

  lastChanged: any;
  changed: boolean;

  el: HTMLElement;

  constructor(el: HTMLElement) {
    this.el = el;
    this.textarea = <HTMLTextAreaElement>el.querySelector(".textarea");
    this.preview = <HTMLDivElement>el.querySelector(".admin_markdown_preview");

    new Autoresize(this.textarea);

    var prefix = document.body.getAttribute("data-admin-prefix");
    var helpLink = <HTMLAnchorElement>(
      el.querySelector(".admin_markdown_show_help")
    );
    helpLink.setAttribute("href", prefix + "/markdown");

    this.lastChanged = Date.now();
    this.changed = false;

    let showChange = <HTMLInputElement>(
      el.querySelector(".admin_markdown_preview_show")
    );
    showChange.addEventListener("change", () => {
      this.preview.classList.toggle("hidden");
    });

    setInterval(() => {
      if (this.changed && Date.now() - this.lastChanged > 500) {
        this.loadPreview();
      }
    }, 100);

    this.textarea.addEventListener("change", this.textareaChanged.bind(this));
    this.textarea.addEventListener("keyup", this.textareaChanged.bind(this));
    this.loadPreview();
    this.bindCommands();
    this.bindShortcuts();
  }

  bindCommands() {
    var btns: any = this.el.querySelectorAll(".admin_markdown_command");
    for (var i = 0; i < btns.length; i++) {
      btns[i].addEventListener("mousedown", (e: any) => {
        var cmd = e.target.getAttribute("data-cmd");
        this.executeCommand(cmd);
        e.preventDefault();
        return false;
      });
    }
  }

  bindShortcuts() {
    this.textarea.addEventListener("keydown", (e) => {
      if (e.metaKey == false && e.ctrlKey == false) {
        return;
      }
      switch (e.keyCode) {
        case 66:
          this.executeCommand("b");
          break;
        case 73:
          this.executeCommand("i");
          break;
        case 75: //k
          this.executeCommand("h2");
          break;
        case 85: //u
          this.executeCommand("a");
          break;
      }
    });
  }

  executeCommand(commandName: string) {
    switch (commandName) {
      case "b":
        this.setAroundMarkdown("**", "**");
        break;
      case "i":
        this.setAroundMarkdown("*", "*");
        break;
      case "a":
        this.setAroundMarkdown("[", "]()");
        var newEnd = this.textarea.selectionEnd + 2;
        this.textarea.selectionStart = newEnd;
        this.textarea.selectionEnd = newEnd;
        break;
      case "h2":
        var start = "## ";
        var end = "";

        var text = this.textarea.value;
        if (text[this.textarea.selectionStart - 1] !== "\n") {
          start = "\n" + start;
        }
        if (text[this.textarea.selectionEnd] !== "\n") {
          end = "\n";
        }
        this.setAroundMarkdown(start, end);
        break;
    }
    this.textareaChanged();
  }

  setAroundMarkdown(before: string, after: string) {
    var text = this.textarea.value;
    var selected = text.substr(
      this.textarea.selectionStart,
      this.textarea.selectionEnd - this.textarea.selectionStart
    );
    var newText = text.substr(0, this.textarea.selectionStart);
    newText += before;
    var newStart = newText.length;
    newText += selected;
    var newEnd = newText.length;
    newText += after;
    newText += text.substr(this.textarea.selectionEnd, text.length);
    this.textarea.value = newText;

    this.textarea.selectionStart = newStart;
    this.textarea.selectionEnd = newEnd;
    this.textarea.focus();
  }

  textareaChanged() {
    this.changed = true;
    this.lastChanged = Date.now();
  }

  loadPreview() {
    this.changed = false;
    var request = new XMLHttpRequest();
    request.open(
      "POST",
      document.body.getAttribute("data-admin-prefix") + "/api/markdown",
      true
    );

    request.addEventListener("load", () => {
      if (request.status == 200) {
        this.preview.innerHTML = JSON.parse(request.response);
      } else {
        console.error("Error while loading markdown preview.");
      }
    });
    request.send(this.textarea.value);
  }
}

class NotificationCenter {
  notifications = new Map<string, NotificationItem>();
  el: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    var data = el.getAttribute("data-notification-views");
    var notifications: NotificationData[] = [];
    if (data) {
      notifications = JSON.parse(data);
    }

    notifications.forEach((item) => {
      this.setData(item);
    });

    this.periodDataLoader();

    /*this.flashNotification("Hello world");
    this.flashNotification("Hello world");
    this.flashNotification("Hello world");
    this.flashNotification("Hello world");
    this.flashNotification("Hello world");
    this.flashNotification("Hello world");*/
  }

  async periodDataLoader() {
    for (;;) {
      if (!document.hidden) this.loadData();
      await sleep(1000);
    }
  }

  loadData() {
    fetch("/admin/api/notifications")
      .then((response) => response.json())
      .then((data: NotificationData[]) =>
        data.forEach((d) => {
          this.setData(d);
        })
      );
  }

  setData(data: NotificationData) {
    var notification: NotificationItem;
    if (this.notifications.has(data.UUID)) {
      notification = this.notifications.get(data.UUID);
    } else {
      notification = new NotificationItem();
      this.notifications.set(data.UUID, notification);
      this.el.appendChild(notification.el);
    }
    notification.setData(data);
  }

  bindNotification(el: HTMLDivElement) {
    el.querySelector(".notification_close").addEventListener("click", () => {
      el.classList.add("notification-closed");
    });
  }

  flashNotification(
    name: string,
    description: string,
    success: boolean,
    fail: boolean
  ) {
    var style = "";
    if (success) {
      style = "success";
    }
    if (fail) {
      style = "fail";
    }
    this.setData({
      UUID: makeid(10),
      Name: name,
      Description: description,
      Flash: false,
      Style: style,
    });
  }
}

interface NotificationData {
  UUID: string;
  PreName?: string;
  Image?: string;
  URL?: string;
  Name: string;
  Description?: string;
  PrimaryAction?: string;
  SecondaryAction?: string;
  DisableCancel?: Boolean;
  Style?: String;
  Progress?: NotificationItemProgress;
  Flash?: Boolean;
}

interface NotificationItemProgress {
  Human: string;
  Percentage: number;
}

class NotificationItem {
  el: HTMLDivElement;
  actionElements: NodeListOf<HTMLDivElement>;
  uuid: string;

  constructor() {
    this.el = document.createElement("div");
    this.el.classList.add("notification");
    this.el.innerHTML = `
      <div class="notification_close"></div>
      <div class="notification_left">
        <div class="notification_left_progress">
          <div class="notification_left_progress_human"></div>
          <progress class="notification_left_progressbar"></progress>
        </div>
      </div>
      <div class="notification_right">
          <div class="notification_prename"></div>
          <div class="notification_name"></div>
          <div class="notification_description"></div>
          <div class="notification_action" data-id="primary"></div>
          <div class="notification_action" data-id="secondary"></div>
      </div>
    `;

    this.actionElements = this.el.querySelectorAll<HTMLDivElement>(
      ".notification_action"
    );
    this.actionElements.forEach((el) => {
      el.addEventListener("click", (e) => {
        var target = <HTMLDivElement>e.currentTarget;
        this.sendAction(target.getAttribute("data-id"));
        return false;
      });
    });

    this.el.querySelector(".notification_left");
    //.setAttribute("style", "background-image: url('/admin/logo');");

    this.el
      .querySelector(".notification_close")
      .addEventListener("click", (e) => {
        this.sendAction("delete");
        this.el.classList.add("notification-closed");
        e.stopPropagation();
        return false;
      });

    this.el.addEventListener("click", () => {
      var url = this.el.getAttribute("data-url");
      if (!url) {
        return;
      }
      window.location.href = url;
    });
  }

  private sendAction(actionID: string) {
    fetch(
      "/admin/api/notifications" +
        encodeParams({
          uuid: this.uuid,
          action: actionID,
        }),
      {
        method: "POST",
      }
    ).then((e) => {
      if (!e.ok) {
        alert("error while deleting notification");
      }
    });
  }

  private setAction(actionEl: HTMLDivElement, action: string) {
    if (!action) {
      actionEl.classList.remove("notification_action-visible");
      return;
    }
    actionEl.classList.add("notification_action-visible");
    actionEl.textContent = action;
  }

  setData(data: NotificationData) {
    this.uuid = data.UUID;
    this.el.querySelector(".notification_prename").textContent = data.PreName;
    this.el.querySelector(".notification_name").textContent = data.Name;
    this.el.querySelector(".notification_description").textContent =
      data.Description;
    this.el
      .querySelector(".notification_description")
      .setAttribute("title", data.Description);

    var left = this.el.querySelector(".notification_left");
    left.classList.remove("notification_left-visible");

    if (data.Image) {
      left.classList.add("notification_left-visible");
      left.setAttribute("style", `background-image: url('${data.Image}');`);
    }

    var closeButton = this.el.querySelector(".notification_close");
    if (data.DisableCancel) {
      closeButton.classList.add("notification_close-disabled");
    } else {
      closeButton.classList.remove("notification_close-disabled");
    }

    this.setAction(this.actionElements[0], data.PrimaryAction);
    this.setAction(this.actionElements[1], data.SecondaryAction);

    this.el.classList.remove("notification-success");
    this.el.classList.remove("notification-fail");
    if (data.Style) {
      this.el.classList.add("notification-" + data.Style);
    }

    var progressEl = this.el.querySelector<HTMLDivElement>(
      ".notification_left_progress"
    );
    if (data.Progress) {
      left.classList.add("notification_left-visible");
      progressEl.classList.add("notification_left_progress-visible");
      this.el.querySelector(".notification_left_progress_human").textContent =
        data.Progress.Human;
      var progressBar = this.el.querySelector<HTMLProgressElement>(
        ".notification_left_progressbar"
      );
      if (data.Progress.Percentage < 0) {
        delete progressBar.value;
        //progressBar.setAttribute("value", "");
      } else {
        progressBar.setAttribute("value", data.Progress.Percentage + "");
      }
    } else {
      progressEl.classList.remove("notification_left_progress-visible");
    }

    if (data.URL) {
      this.el.classList.add("notification-clickable");
      this.el.setAttribute("data-url", data.URL);
    } else {
      this.el.classList.remove("notification-clickable");
      this.el.setAttribute("data-url", "");
    }

    if (data.Flash) {
      window.setTimeout(() => {
        this.close();
      }, 1000);
    }
  }

  close() {
    this.el.classList.add("notification-closed");
  }
}

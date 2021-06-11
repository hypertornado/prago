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

    /*

    var action: NotificationItemAction = {
      Name: "Ukončit",
      ID: "aaa",
    };

    this.setData({
      UUID: "SS",
      Name: "XXX",
      PrimaryAction: action,
      SecondaryAction: {
        Name: "Storno",
        ID: "XXX",
      },
    });

    this.setData({
      UUID: "xXXX",
      PreName: "novinky ze světa",
      Image:
        "https://www.prago-cdn.com/lazne/kRX9YPoKMqD3IKQk1Lmy/2500/48dc142cd1/mvx0017.jpg",
      Name: "Zemřel manžel britské královny. Princi Philipovi bylo 99 let",
      Progress: {
        Human: "76 %",
        Percentage: 0.76,
      },
      Description:
        "Britská královská rodina a celá Velká Británie truchlí. Ve věku 99 let zemřel princ Philip. Manžel britské panovnice Alžběty II., Jeho královská Výsost vévoda z Edinburghu zemřel ráno 9. dubna na zámku Windsor, královská rodina zprávu potvrdila na sociálních sítích.",
    });

    this.setData({
      UUID: "xXsssXXss",
      Name: "OK",
      URL: "/",
    });

    this.setData({
      UUID: "xXXXss",
      Name: "can't cancel",
      DisableCancel: true,
      Progress: {
        Human: "56 %",
        Percentage: 0.56,
      },
      Style: "fail",
    });

    this.setData({
      UUID: "xXXXss2",
      Name: "can't cancel",
      DisableCancel: true,
      Progress: {
        Human: "",
        Percentage: -1,
      },
      Style: "success",
    });*/

    this.periodDataLoader();
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
}

interface NotificationItemProgress {
  Human: string;
  Percentage: Number;
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
  }
}

class NotificationCenter {
  notifications = new Map<string, NotificationItem>();
  el: HTMLDivElement;

  constructor(el: HTMLDivElement) {
    this.el = el;
    var notifications: NotificationData[] = JSON.parse(
      el.getAttribute("data-notification-views")
    );

    notifications.forEach((item) => {
      this.setData(item);
      //this.getItem(item.UUID).setData(item);
    });

    console.log(notifications);

    return;

    /*
    var notifications = el.querySelectorAll(".notification");
    for (var i = 0; i < notifications.length; i++) {
      this.bindNotification(<HTMLDivElement>notifications[i]);
    }*/
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
  Name: string;
}

class NotificationItem {
  el: HTMLDivElement;

  constructor() {
    this.el = document.createElement("div");
    this.el.classList.add("notification");
    this.el.innerHTML = `
      <div class="notification_close"></div>
      <div class="notification_left"></div>
      <div class="notification_right">
          <div class="notification_name"></div>
      </div>
    `;

    this.el
      .querySelector(".notification_close")
      .addEventListener("click", () => {
        this.el.classList.add("notification-closed");
      });
  }

  setData(data: NotificationData) {
    this.el.querySelector(".notification_name").textContent = data.Name;
  }
}

class NotificationCenterOLD {
  el: HTMLDivElement;
  adminPrefix: string;
  notifications: Array<NotificationItemOLD>;

  constructor(el: HTMLDivElement) {
    this.el = el;
    this.notifications = Array<NotificationItemOLD>();
    this.adminPrefix = document.body.getAttribute("data-admin-prefix");
    this.loadNotifications();
  }

  loadNotifications() {
    //console.log("notifications not implemented");
    return;
    var request = new XMLHttpRequest();
    request.open("GET", this.adminPrefix + "/_api/notifications", true);
    request.addEventListener("load", () => {
      if (request.status == 200) {
        var notifications = JSON.parse(request.response);
        notifications.Views.forEach((item: any) => {
          var notification = this.createNotification(item);
          this.notifications.push(notification);
          this.el.appendChild(notification.el);
        });
      } else {
        console.log("failed to load notifications");
      }
    });
    request.send();
  }

  //https://stackoverflow.com/questions/7381974/which-characters-need-to-be-escaped-in-html

  createNotification(data: any): NotificationItemOLD {
    return new NotificationItemOLD(data);
  }
}

class NotificationItemOLD {
  adminPrefix: String;

  el: HTMLDivElement;

  uuid: string;

  constructor(data: any) {
    this.adminPrefix = document.body.getAttribute("data-admin-prefix");
    this.createElement(data);
  }

  createElement(data: any) {
    var ret: HTMLDivElement;
    ret = document.createElement("div");

    ret.innerHTML = `
            <div class="notification">
                <div class="notification_close"></div>
                <div class="notification_left"></div>
                <div class="notification_right">
                    <div class="notification_name">${e(data.Name)}</div>
                </div>
            </div>        
        `;
    this.el = <HTMLDivElement>ret.children[0];

    this.el
      .querySelector(".notification_close")
      .addEventListener("click", this.closeNotification.bind(this));

    //var world = "x<b ls='3'>x</b>x";
    //this.el.innerHTML = `hello ${escapeHTML(world)}s`;
  }

  closeNotification() {
    this.el.classList.add("notification-closed");
    fetch(this.adminPrefix + "/_api/notification/" + this.uuid, {
      method: "DELETE",
    })
      .then(console.log)
      .then((e) => {
        console.log(e);
      });
  }
}

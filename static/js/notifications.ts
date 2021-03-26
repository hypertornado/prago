function bindNotifications() {
    new NotificationCenter2(document.querySelector(".notification_center"));
}

class NotificationCenter2 {

    constructor(el: HTMLDivElement) {
        var notifications = el.querySelectorAll(".notification");
        for (var i = 0; i < notifications.length; i++) {
            this.bindNotification(<HTMLDivElement>notifications[i])
        }
    }

    bindNotification(el: HTMLDivElement) {
        el.querySelector(".notification_close").addEventListener("click", () => {
            el.classList.add("notification-closed");
        })
    }

}

class NotificationCenter {

    el: HTMLDivElement;
    adminPrefix: string;
    notifications: Array<NotificationItem>;

    constructor(el: HTMLDivElement) {
        this.el = el;
        this.notifications = Array<NotificationItem>();
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.loadNotifications();
    }

    loadNotifications() {
        //console.log("notifications not implemented");
        return
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

    createNotification(data: any): NotificationItem {
        return new NotificationItem(data);
    }

}

class NotificationItem {

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

        this.el.querySelector(".notification_close").addEventListener("click", this.closeNotification.bind(this));

        //var world = "x<b ls='3'>x</b>x";
        //this.el.innerHTML = `hello ${escapeHTML(world)}s`;
    }

    closeNotification() {
        this.el.classList.add("notification-closed");
        fetch(this.adminPrefix + "/_api/notification/"+this.uuid, {method: "DELETE"}).then(console.log).then((e) => {
            console.log(e);
        });
    }

}
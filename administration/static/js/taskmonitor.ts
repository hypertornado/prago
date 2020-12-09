function bindTaskMonitor() {
    var el: HTMLDivElement = document.querySelector(".taskmonitorcontainer");
    if (el) {
        new TaskMonitor(el);
    }
}

class TaskMonitor {

    el: HTMLDivElement;

    constructor(el: HTMLDivElement) {
        this.el = el;
        window.setInterval(this.load.bind(this), 1000);
    }

    load() {
        var request = new XMLHttpRequest();
        request.open("GET", "/admin/_tasks/running", true);
        request.addEventListener("load", () => {
            this.el.innerHTML = "";
            if (request.status == 200) {
                this.el.innerHTML = request.response;
            } else {
                console.error("error while loading list");
            }
        });
        request.send();
    }

}
class TaskMonitor {
  el: HTMLDivElement;

  constructor() {
    var el = document.querySelector<HTMLDivElement>(".taskmonitorcontainer");
    if (!el) {
      return;
    }
    this.el = el;
    window.setInterval(this.load.bind(this), 1000);
  }

  load() {
    var request = new XMLHttpRequest();
    request.open("GET", "/admin/api/tasks/running", true);
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

function initTooltips() {
  deleteTooltips();
  let els: any = document.querySelectorAll("[data-tooltip]");

  for (var i = 0; i < els.length; i++) {
    let el = <HTMLDivElement>els[i];
    if (el.getAttribute("data-hastooltip")) {
      continue;
    }
    el.setAttribute("data-hastooltip", "true");

    el.addEventListener("mouseenter", (e: Event) => {
        let targetEl = <HTMLDivElement>e.target;
        showTooltip(targetEl, el.getAttribute("data-tooltip"))
    });

    el.addEventListener("mouseleave", (e: Event) => {
        deleteTooltips();
    })

  }
}

function showTooltip(targetEl: HTMLDivElement, text: string): HTMLDivElement {
    deleteTooltips();
    let tooltipEl = document.createElement("div");
    tooltipEl.classList.add("tooltip");
    tooltipEl.innerText = text;
    document.body.appendChild(tooltipEl);

    let elWidth = tooltipEl.clientWidth;
    let elHeight = tooltipEl.clientHeight;

    let viewportWidth = window.innerWidth;
    let viewportHeight = window.innerHeight;

    //let targetEl = <HTMLDivElement>data.Event.currentTarget;
    let rect = targetEl.getBoundingClientRect();

    var x, y: number = 0;

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

    tooltipEl.style.left = x + "px";
    tooltipEl.style.top = y + "px";

    return tooltipEl;
}

function deleteTooltips() {
    let tooltips = document.querySelectorAll(".tooltip");
    for (var i = 0; i < tooltips.length; i++) {
        tooltips[i].remove();
    }
}
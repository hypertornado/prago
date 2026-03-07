
function initTables() {
    var tables =
    document.querySelectorAll<HTMLDivElement>(".form_table");
    tables.forEach((el) => {
        if (el.getAttribute("data-table-initiated") != "true") {
            new Table(el);
        }
    });
}


class Table {

    el: HTMLDivElement;
    cellsToLoad: HTMLTableCellElement[];

    constructor(el: HTMLDivElement) {

        this.el = el;
        el.setAttribute("data-table-initiated", "true");
        this.cellsToLoad = [];

        let cells = el.querySelectorAll<HTMLTableCellElement>("td.form_table_cell");
        cells.forEach((cell) => {
            this.bindCell(cell);
        });
        this.loadCellsAsync();
    }

    bindCell(cell: HTMLTableCellElement) {
        let cellAsyncURL = cell.getAttribute("data-async-data-url");
        if (!cellAsyncURL) {
            return;
        }

        cell.classList.add("form_table_cell-loading");

        //let textEl = cell.querySelector(".form_table_cell_text");
        //textEl.textContent = "⏳";
        this.cellsToLoad.push(cell);

    }

    loadCellsAsync(){
        if (this.cellsToLoad.length == 0) {
            return
        }

        let cell = this.cellsToLoad.shift();

        let descriptionsBefore = cell.querySelector(".form_table_cell_descriptions_before");
        let descriptionsAfter = cell.querySelector(".form_table_cell_descriptions_after");

        let cellAsyncURL = cell.getAttribute("data-async-data-url");
        let textEl = cell.querySelector(".form_table_cell_text");

        let request = new XMLHttpRequest();
        request.open("GET", cellAsyncURL);

        request.addEventListener("load", (e) => {
            if (request.status == 200) {
                let item = JSON.parse(request.response);
                textEl.textContent = item.Text;
                descriptionsBefore.textContent = "";
                descriptionsAfter.textContent = "";

                if (item.DescriptionsBefore) {
                    for (let i = 0; i < item.DescriptionsBefore.length; i++) {
                        let descText = item.DescriptionsBefore[i];
                        let descDiv = document.createElement("div");
                        descDiv.innerText = descText;
                        descDiv.classList.add("form_table_cell_descriptionbefore");
                        descriptionsBefore.appendChild(descDiv);
                    }
                }

                if (item.DescriptionsAfter) {
                    for (let i = 0; i < item.DescriptionsAfter.length; i++) {
                        let descText = item.DescriptionsAfter[i];
                        let descDiv = document.createElement("div");
                        descDiv.innerText = descText;
                        descDiv.classList.add("form_table_cell_descriptionafter");
                        descriptionsAfter.appendChild(descDiv);
                    }
                }

                if (item.Green) {
                    cell.classList.add("form_table_cell-green");
                }
                if (item.Orange) {
                    cell.classList.add("form_table_cell-orange");
                }
                if (item.Red) {
                    cell.classList.add("form_table_cell-red");
                }

            } else {
                textEl.textContent = "💥";
            }

            cell.classList.remove("form_table_cell-loading");
            this.loadCellsAsync();
        });
        request.send();
    }

}
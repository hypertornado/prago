class Timeline {
    el: HTMLDivElement;
    valuesEl: HTMLDivElement;
    datepicker: HTMLInputElement;
    monthpicker: HTMLInputElement;
    yearpicker: HTMLInputElement;

    typeSelect: HTMLSelectElement;

    constructor(el: HTMLDivElement) {
        this.el = el;
        this.typeSelect = el.querySelector(".timeline_toolbar_type");
        this.valuesEl = el.querySelector(".timeline_values");
        this.datepicker = el.querySelector(".timeline_toolbar_date");
        this.monthpicker = el.querySelector(".timeline_toolbar_month");
        this.yearpicker = el.querySelector(".timeline_toolbar_year");
        
        this.datepicker.valueAsDate = new Date();
        this.monthpicker.valueAsDate = new Date();
        this.yearpicker.value = new Date().getFullYear() + "";


        this.typeSelect.addEventListener("change", this.changedType.bind(this));
        this.datepicker.addEventListener("change", this.changedType.bind(this));
        this.monthpicker.addEventListener("change", this.changedType.bind(this));
        this.yearpicker.addEventListener("change", this.changedType.bind(this));

        this.el.querySelector(".timeline_toolbar_prev").addEventListener("click", () => {
            this.changePosition(false);
        })
        this.el.querySelector(".timeline_toolbar_next").addEventListener("click", () => {
            this.changePosition(true);
        })
        
        this.changedType();
    }

    loadData() {
        this.setLoader();

        var dateStr: string;
        let typ = this.typeSelect.value;
        if (typ == "day") {
            dateStr = this.datepicker.value;
        }
        if (typ == "month") {
            dateStr = this.monthpicker.value;
        }
        if (typ == "year") {
            dateStr = this.yearpicker.value;
        }


        var request = new XMLHttpRequest();
        var params: any = {
            uuid: this.el.getAttribute("data-uuid"),
            date: dateStr,
            width: this.el.clientWidth,
        };


        request.addEventListener("load", () => {
        if (request.status == 200) {
            let data = JSON.parse(request.response)
            this.setData(data);
        } else {
            console.error("Error while loading timeline");
        }
        });

        request.open(
        "GET",
        "/admin/api/timeline" + encodeParams(params),
        true
        );

        request.send();
    }

    setData(data: any) {
        this.valuesEl.innerText = "";
        for (var i = 0; i < data.Values.length; i++) {
            let val = data.Values[i];
            this.setValue(val)
        }
    }

    setLoader() {
        this.valuesEl.innerText = "Loading...";
    }

    setValue(data: any) {
        let valEl = document.createElement("div");
        valEl.innerHTML = `
            <div class="timeline_value_bars"></div>
            <div class="timeline_value_name" title="${data.Name}">${data.Name}</div>
        `
        valEl.classList.add("timeline_value");

        let barsEl: HTMLDivElement = valEl.querySelector(".timeline_value_bars");
        for (var i = 0; i < data.Bars.length; i++) {
            this.addBar(barsEl, data.Bars[i]);
        }

        if (data.IsCurrent) {
            valEl.classList.add("timeline_value-current");
        }
        this.valuesEl.appendChild(valEl);
    }

    addBar(el: HTMLDivElement, barValue: any) {
        let barEl = document.createElement("div");
        barEl.innerHTML = `
            <div class="timeline_value_bar_inner" style="${barValue.StyleCSS}"></div>
        `;
        barEl.setAttribute("title", barValue.ValueText);
        barEl.classList.add("timeline_value_bar");
        el.appendChild(barEl);

    }

    changedType() {

        let typ = this.typeSelect.value;

        this.datepicker.classList.add("hidden");
        this.monthpicker.classList.add("hidden");
        this.yearpicker.classList.add("hidden");

        if (typ == "day") {
            this.datepicker.classList.remove("hidden");
        }
        if (typ == "month") {
            this.monthpicker.classList.remove("hidden");
        }
        if (typ == "year") {
            this.yearpicker.classList.remove("hidden");
        }
        this.loadData();
    }

    changePosition(next: boolean) {
        let typ = this.typeSelect.value;
        if (typ == "day") {
            let date = new Date(this.datepicker.value);    
            if (isNaN(date.getTime())) {
                console.error("Invalid date format. Please use YYYY-MM-DD.");
                return
            }
            
            var addNumber = -1;
            if (next) {
                addNumber = 1;
            }
            date.setDate(date.getDate() + addNumber);
            this.datepicker.value = date.toISOString().split("T")[0];
        }
        if (typ == "month") {
            let vals = this.monthpicker.value.split("-");
            let year = parseInt(vals[0]);
            let month = parseInt(vals[1]);
            if (next) {
                month += 1
                if (month > 12) {
                    month = 1;
                    year += 1;
                }
            } else {
                month -= 1
                if (month < 1) {
                    month = 12;
                    year -= 1;
                }
            }

            let format = year + "-";
            if (month < 10) {
                format += "0"
            }
            format += month + ""

            this.monthpicker.value = format;
        }
        if (typ == "year") {
            let year = parseInt(this.yearpicker.value);
            if (next) {
                year += 1;
            } else {
                year -= 1;
            }
            this.yearpicker.value = year + "";
        }
        this.loadData();
    }



}
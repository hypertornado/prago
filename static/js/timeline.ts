class Timeline {
    el: HTMLDivElement;
    valuesEl: HTMLDivElement;
    datepicker: HTMLInputElement;
    monthpicker: HTMLInputElement;
    yearpicker: HTMLInputElement;

    lastRequest: XMLHttpRequest;

    typeValue: string;
    alignmentValue: string;

    settingsOptions: any;

    filtersEl: HTMLDivElement;


    cache: any;

    constructor(el: HTMLDivElement) {
        this.el = el;
        this.valuesEl = el.querySelector(".timeline_values");
        this.datepicker = el.querySelector(".timeline_toolbar_date");
        this.monthpicker = el.querySelector(".timeline_toolbar_month");
        this.yearpicker = el.querySelector(".timeline_toolbar_year");

        const now = new Date();
        const year = now.getFullYear();
        const month = String(now.getMonth() + 1).padStart(2, '0'); 
        
        this.datepicker.valueAsDate = new Date();
        //this.monthpicker.valueAsDate = new Date();
        this.monthpicker.value = `${year}-${month}`;
        this.yearpicker.value = new Date().getFullYear() + "";
        this.cache = {};
        this.settingsOptions = {};

        this.typeValue = "day";
        this.alignmentValue = this.el.getAttribute("data-alignment");
        this.datepicker.addEventListener("change", this.changedType.bind(this));
        this.monthpicker.addEventListener("change", this.changedType.bind(this));
        this.yearpicker.addEventListener("change", this.changedType.bind(this));

        this.el.querySelector(".timeline_toolbar_prev").addEventListener("click", () => {
            this.changePosition(false);
        });
        this.el.querySelector(".timeline_toolbar_next").addEventListener("click", () => {
            this.changePosition(true);
        });

        window.addEventListener("resize", this.loadData.bind(this));

        this.el.querySelector(".timeline_toolbar_settings").addEventListener("click", this.settingsClicked.bind(this));

        this.filtersEl = this.el.querySelector(".timeline_filters");

        this.el.querySelector(".timeline_toolbar_fullscreen_open").addEventListener("click", () => {
            this.el.classList.add("timeline-fullscreen");
            this.loadData();
        });
        this.el.querySelector(".timeline_toolbar_fullscreen_close").addEventListener("click", () => {
            this.el.classList.remove("timeline-fullscreen");
            this.loadData();
        });
        
        this.changedType();
    }

    settingsClicked() {
        /*let paramsS = {
            "_uuid": this.el.getAttribute("data-uuid"),
            "_alignment": this.alignmentValue,
            "_type": this.typeValue,
        };*/
        let params = JSON.parse(JSON.stringify(this.settingsOptions));
        params["_uuid"] = this.el.getAttribute("data-uuid");
        params["_alignment"] = this.alignmentValue;
        params["_type"] = this.typeValue;
        new PopupForm("/admin/_timeline-settings" + encodeParams(params), (data: any) => {
            this.cache = {};
            this.settingsOptions = data.Data;
            console.log(data);
            this.alignmentValue = data.Data["_alignment"];
            this.typeValue = data.Data["_type"];
            this.changedType();
        })
    }

    loadData() {
        this.setLoader();

        var dateStr: string;
        let typ = this.typeValue;
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
        if (this.lastRequest) {
            this.lastRequest.abort();
        }
        this.lastRequest = request;

        var params: any = {
            UUID: this.el.getAttribute("data-uuid"),
            DateStr: dateStr,
            Width: this.el.clientWidth,
            ValueCache: this.cache,
            Alignment: this.alignmentValue,
            Options: this.settingsOptions,
        };


        request.addEventListener("load", () => {
            if (request.status == 200) {
                let data = JSON.parse(request.response)
                this.setData(data);
            } else {
                this.valuesEl.innerText = "Error while loading timeline";
                console.error("Error while loading timeline");
            }
            this.lastRequest = null;
        });

        request.open(
            "POST",
            "/admin/api/timeline",
            true
        );
        request.send(JSON.stringify(params));
    }

    setData(data: any) {
        this.valuesEl.innerText = "";

        let linesEl = document.createElement("div");
        linesEl.classList.add("timeline_lines");
        this.valuesEl.appendChild(linesEl);

        this.filtersEl.innerHTML = "";
        this.filtersEl.classList.add("hidden");
        for (var i = 0; i < data.Filters.length; i++) {
            this.setFilter(data.Filters[i]);
        }

        /*this.setFilter({
            "KeyName": "Neco",
            "ValueName": "1234",
        });

        this.setFilter({
            "KeyName": "B",
            "ValueName": "1dwdw234",
        });*/

        for (var i = 0; i < data.Lines.length; i++) {
            this.drawLine(linesEl, data.Lines[i]);
        }

        for (var i = 0; i < data.Values.length; i++) {
            let val = data.Values[i];
            this.setValue(val)
        }
    }

    setFilter(data: any) {
        this.filtersEl.classList.remove("hidden");
        let el = document.createElement("div");
        el.innerText = `${data.KeyName}: ${data.ValueName}`;
        el.classList.add("timeline_filter");
        this.filtersEl.appendChild(el);
        console.log("filter", data);
    }

    drawLine(linesEl: HTMLDivElement, data: any) {
        let lineEl = document.createElement("div");
        lineEl.classList.add("timeline_line");
        lineEl.setAttribute("style", data.StyleCSS);
        if (data.IsZero) {
            lineEl.classList.add("timeline_line-iszero");
        }
        //lineEl.setAttribute("")

        let lineNameEl = document.createElement("div");
        lineNameEl.classList.add("timeline_line_name");
        lineNameEl.textContent = data.Name;

        lineEl.appendChild(lineNameEl);

        linesEl.appendChild(lineEl);
    }

    setLoader() {
        this.valuesEl.innerHTML = '<progress class="progress"></progress>';
    }

    setValue(data: any) {
        this.cache[data.DateID] = data.Value;

        let valEl = document.createElement("div");
        valEl.innerHTML = `
            <div class="timeline_value_bars"></div>
            <div class="timeline_value_name">
                <span class="timeline_value_name_inner">${data.Name}</span>
            </div>
        `
        valEl.classList.add("timeline_value");

        let barsEl: HTMLDivElement = valEl.querySelector(".timeline_value_bars");
        this.addBar(barsEl, data);

        if (data.IsCurrent) {
            valEl.classList.add("timeline_value-current");
        }
        if (data.IsSelected) {
            valEl.classList.add("timeline_value-selected");
        }
        this.valuesEl.appendChild(valEl);
    }

    addBar(el: HTMLDivElement, barValue: any) {
        let barEl = document.createElement("div");

        var styleName: string;
        if (barValue.Value >= 0) {
            styleName = "timeline_value_bar_inner-positive";
        } else {
            styleName = "timeline_value_bar_inner-negative";
        }

        barEl.innerHTML = `
            <div class="timeline_value_bar_inner ${styleName}" style="${barValue.StyleCSS}"></div>
        `;
        barEl.classList.add("timeline_value_bar");
        el.appendChild(barEl);

        let labelEl = document.createElement("div");
        labelEl.classList.add("timeline_value_label");
        labelEl.innerText = barValue.ValueText;
        labelEl.setAttribute("style", barValue.LabelStyleCSS);
        barEl.appendChild(labelEl);

    }

    changedType() {
        let typ = this.typeValue;
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
        let typ = this.typeValue;
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
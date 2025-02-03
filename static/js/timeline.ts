class Timeline {
    el: HTMLDivElement;
    valuesEl: HTMLDivElement;
    datepicker: HTMLInputElement;

    constructor(el: HTMLDivElement) {
        this.el = el;
        this.valuesEl = el.querySelector(".timeline_values");
        this.datepicker = el.querySelector(".timeline_toolbar_date");
        this.datepicker.valueAsDate = new Date();

        this.datepicker.addEventListener("change", this.loadData.bind(this));
        
        this.loadData();
    }

    loadData() {

        let totalWidth = this.el.clientWidth;
        let optimalSize = 40;
        let columnsCount = Math.floor(totalWidth / optimalSize);
        if (columnsCount < 10) {
            columnsCount = 10;
        }



        this.setLoader();

        var request = new XMLHttpRequest();
        var params: any = {
            uuid: this.el.getAttribute("data-uuid"),
            date: this.datepicker.value,
            columns: columnsCount,
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
            <div class="timeline_value_top">
                <div class="timeline_value_graph" style="${data.StyleCSS}"></div>
            </div>
            <div class="timeline_value_human" title="${data.ValueText}">${data.ValueText}</div>
            <div class="timeline_value_name" title="${data.Name}">${data.Name}</div>
        `
        valEl.classList.add("timeline_value");
        if (data.IsCurrent) {
            valEl.classList.add("timeline_value-current");
        }
        this.valuesEl.appendChild(valEl);

    }

}
class ListFilterItem {
    filter: ListFilter;
    el: HTMLDivElement;
    filterLayout: string;

    isListFilter2: boolean;
    filter2El: HTMLDivElement;
    filter2NameEl: HTMLDivElement;
    filter2Data: any;
    value: string;

    constructor(filter: ListFilter, el: HTMLDivElement, params: URLSearchParams) {
        this.filter = filter;
        this.el = el;

        var fieldName = el.getAttribute("data-name");
        this.filterLayout = el.getAttribute("data-filter-layout");
        var fieldInput = el.querySelector("input");
        var fieldSelect = el.querySelector("select");
        var fieldValue = params.get(fieldName);

        this.value = fieldValue;

        if (fieldValue) {
            if (fieldInput) {
                fieldInput.value = fieldValue;
            }
            if (fieldSelect) {
                fieldSelect.value = fieldValue;
            }
        }

        if (fieldInput) {
            fieldInput.addEventListener("input", this.inputListener.bind(this));
            fieldInput.addEventListener("change", this.inputListener.bind(this));
        }

        if (fieldSelect) {
            fieldSelect.addEventListener("input", this.inputListener.bind(this));
            fieldSelect.addEventListener("change", this.inputListener.bind(this));
        }

        if (this.filterLayout == "filter_layout_text" || this.filterLayout == "filter_layout_number" || this.filterLayout == "filter_layout_select" || this.filterLayout == "filter_layout_relation" || this.filterLayout == "filter_layout_date" || this.filterLayout == "filter_layout_boolean") {
            this.initFilter2();
            return;
        }

        /*if (this.filterLayout == "filter_layout_relation") {
            new ListFilterRelations(el, fieldValue, this.filter.list);
        }*/

        /*if (this.filterLayout == "filter_layout_date") {
            new ListFilterDate(el, fieldValue);
        }*/
    }

    inputListener(e: any) {
        if (
            e.keyCode == 9 ||
            e.keyCode == 16 ||
            e.keyCode == 17 ||
            e.keyCode == 18
        ) {
            return;
        }
        this.filter.filterChanged();
    }

    initFilter2() {
        this.isListFilter2 = true
        this.filter2El = this.el.querySelector(".list_filter2");
        this.filter2NameEl = this.el.querySelector(".list_filter2_name");
        this.filter2El.addEventListener("click", this.filter2Clicked.bind(this));

        let data = JSON.parse(this.filter2El.getAttribute("data-filter-content"));
        this.setFilter2Data(data);
    }

    filter2Clicked() {
        new PopupForm("/admin/_list-fiter-item?field=" + this.getFieldID() + "&resource=" + this.filter.list.typeName + "&value=" + this.getFieldValue(), (data: any) => {
            this.setFilter2Data(data.Data);
        });
    }

    setFilter2Data(data: any) {
        this.value = "";
        this.setFilter2Name("");
        if (data) {
            this.value = data.ID;
            this.setFilter2Name(data.Name);
        }

        this.filter.filterChanged();
    }

    setFilter2Name(name: string) {
        this.filter2El.title = name;
        this.filter2NameEl.innerText = name;
    }

    getFieldID(): string {
        return this.el.getAttribute("data-name");
    }

    getFieldValue(): string {
        if (this.isListFilter2) {
            return this.value;
        }

        if (this.filterLayout == "filter_layout_select") {
            return this.el.querySelector("input").value;
        }

        if (this.filterLayout == "filter_layout_relation") {
            let hiddenEl: HTMLInputElement = this.el.querySelector(".filter_relations_hidden");
            return hiddenEl.value;
        }

        let input: HTMLInputElement = this.el.querySelector(".list_filter_input");
        if (!input) {
            return "";
        }
        return input.value.trim();
    }

    setColor() {
        var val = this.getFieldValue();
        if (val) {
            this.el.classList.add("list_filteritem-colored");
        } else {
            this.el.classList.remove("list_filteritem-colored");
        }
    }
}
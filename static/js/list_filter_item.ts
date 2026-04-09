class ListFilterItem {
    filter: ListFilter;
    el: HTMLDivElement;
    filterLayout: string;
    isListFilter2: boolean = false;
    filter2El: HTMLDivElement | null = null;
    filter2NameEl: HTMLDivElement | null = null;
    filter2Data: any;

    closeButton: HTMLButtonElement | null = null;

    filterInput: HTMLInputElement | null = null;

    key: string;
    value: string;

    constructor(filter: ListFilter, el: HTMLDivElement, params: URLSearchParams) {
        this.filter = filter;
        this.el = el;

        this.key = <string>el.getAttribute("data-name");
        this.filterLayout = <string>el.getAttribute("data-filter-layout");

        this.value = <string>params.get(<string>this.key);

        if (this.filterLayout) {
            this.initFilter2();
        }
    }


    initFilter2() {
        this.isListFilter2 = true
        this.filter2El = <HTMLDivElement>this.el.querySelector(".list_filter2");
        this.filterInput = <HTMLInputElement>this.el.querySelector(".list_filter2_input");
        this.filter2NameEl = <HTMLDivElement>this.el.querySelector(".list_filter2_name");
        this.filter2El.addEventListener("click", this.filter2Clicked.bind(this));

        this.closeButton = <HTMLButtonElement>this.el.querySelector(".list_filter2_close");
        this.closeButton.addEventListener("click", this.closeButtonClicked.bind(this));

        if (this.isInlineItem()) {
            this.filterInput.classList.remove("hidden");
            this.filter2NameEl.classList.add("hidden");
            this.filterInput.addEventListener("input", this.inlineInputChange.bind(this));
            this.filterInput.addEventListener("keydown", (e: KeyboardEvent) => {
                if (e.keyCode == 27) {
                    this.filterInput!.value = "";
                    this.inlineInputChange();
                }
            })
        } else {
            this.filterInput.classList.add("hidden");
            this.filter2NameEl.classList.remove("hidden");
        }

        let data = JSON.parse(<any>this.filter2El.getAttribute("data-filter-content"));
        this.setFilter2Data(data);
    }

    inlineInputChange() {
        let val = this.filterInput!.value;
        this.setFilter2Data({
            ID: val,
            Name: val,
        });
        this.filter.filterChanged();

    }

    closeButtonClicked(e: Event) {
        this.setFilter2Data(null);
        this.filter.filterChanged();

        e.preventDefault();
        e.stopPropagation();

    }

    filter2Clicked() {
        if (this.isInlineItem()) {
            return;
        }
        new PopupForm("/admin/_list-fiter-item?field=" + this.getFieldID() + "&resource=" + this.filter.list.typeName + "&value=" + this.getFieldValue(), (data: any) => {
            this.setFilter2Data(data.Data);
            this.filter.filterChanged();
        });
    }

    setFilter2Data(data: any) {
        this.value = "";
        this.closeButton!.classList.add("hidden");
        if (data) {
            this.value = data.ID;
            if (this.value) {
                this.closeButton!.classList.remove("hidden");
            }
            this.setFilter2Name(data.Name);
            this.setInlineValue(this.value);
        } else {
            this.setFilter2Name("");
            this.setInlineValue("");
        }
    }

    setFilter2Name(name: string) {
        this.filter2El!.title = name;
        this.filter2NameEl!.innerText = name;
    }

    setInlineValue(name: string) {
        if (this.isInlineItem() && this.filterInput!.value != name) {
            this.filterInput!.value = name;
        }
    }

    getFieldID(): string {
        return this.key;
    }

    getFieldValue(): string | null {
        if (this.isListFilter2) {
            return this.value;
        }
        return null;
    }

    setColor() {
        var val = this.getFieldValue();
        if (val) {
            this.el.classList.add("list_filteritem-colored");
        } else {
            this.el.classList.remove("list_filteritem-colored");
        }
    }

    isInlineItem(): boolean {
        if (this.filterLayout == "filter_layout_text" || this.filterLayout == "filter_layout_number") {
            return true;
        }
        return false;
    }
}
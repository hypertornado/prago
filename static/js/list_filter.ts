class ListFilter {
    list: List;
    listFilterItems: ListFilterItem[];

    constructor(list: List, params: URLSearchParams) {
        this.list = list;
        this.listFilterItems = [];
        var filterFields = this.list.listEl.querySelectorAll(".list_header_item_filter");
        for (var i = 0; i < filterFields.length; i++) {
            var field: HTMLDivElement = <HTMLDivElement>filterFields[i];
            this.listFilterItems.push(new ListFilterItem(this, field, params));
        }
    }

    filterChanged() {
        this.colorActiveFilterItems();
        this.list.page = 1;
        this.list.changed = true;
        this.list.changedTimestamp = Date.now();
        this.list.listEl.classList.add("list-loading");
    }

    getFilterData(): any {
        var ret: any = {};
        this.listFilterItems.forEach((item) => {
            var val = item.getFieldValue();
            if (val) {
                ret[item.getFieldID()] = val;
            }
        });
        return ret;
    }

    colorActiveFilterItems() {
        let itemsToColor = this.getFilterData();
        this.listFilterItems.forEach((item) => {
            item.setColor();
        });
    }
}

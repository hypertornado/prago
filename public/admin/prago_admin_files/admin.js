function bindStats() {
}
class Autoresize {
    constructor(el) {
        this.el = el;
        this.el.addEventListener('input', this.resizeIt.bind(this));
        this.resizeIt();
    }
    resizeIt() {
        var height = this.el.scrollHeight + 2;
        this.el.style.height = height + 'px';
    }
}
function DOMinsertChildAtIndex(parent, child, index) {
    if (index >= parent.children.length) {
        parent.appendChild(child);
    }
    else {
        parent.insertBefore(child, parent.children[index]);
    }
}
function encodeParams(data) {
    var ret = "";
    for (var k in data) {
        if (!data[k]) {
            continue;
        }
        if (ret != "") {
            ret += "&";
        }
        ret += encodeURIComponent(k) + "=" + encodeURIComponent(data[k]);
    }
    if (ret != "") {
        ret = "?" + ret;
    }
    return ret;
}
function e(str) {
    return escapeHTML(str);
}
function escapeHTML(str) {
    str = str.split("&").join("&amp;");
    str = str.split("<").join("&lt;");
    str = str.split(">").join("&gt;");
    str = str.split("\"").join("&quot;");
    str = str.split("'").join("&#39;");
    return str;
}
function bindImageViews() {
    var els = document.querySelectorAll(".admin_item_view_image_content");
    for (var i = 0; i < els.length; i++) {
        new ImageView(els[i]);
    }
}
class ImageView {
    constructor(el) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.el = el;
        var ids = el.getAttribute("data-images").split(",");
        this.addImages(ids);
    }
    addImages(ids) {
        this.el.innerHTML = "";
        for (var i = 0; i < ids.length; i++) {
            if (ids[i] != "") {
                this.addImage(ids[i]);
            }
        }
    }
    addImage(id) {
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("href", this.adminPrefix + "/file/api/redirect-uuid/" + id);
        container.setAttribute("style", "background-image: url('" + this.adminPrefix + "/file/api/redirect-thumb/" + id + "');");
        var img = document.createElement("div");
        img.setAttribute("src", this.adminPrefix + "/file/api/redirect-thumb/" + id);
        img.setAttribute("draggable", "false");
        var descriptionEl = document.createElement("div");
        descriptionEl.classList.add("admin_images_image_description");
        container.appendChild(descriptionEl);
        var request = new XMLHttpRequest();
        request.open("GET", this.adminPrefix + "/file/api/imagedata/" + id);
        request.addEventListener("load", (e) => {
            if (request.status == 200) {
                var data = JSON.parse(request.response);
                descriptionEl.innerText = data["Name"];
                container.setAttribute("title", data["Name"]);
            }
            else {
                console.error("Error while loading file metadata.");
            }
        });
        request.send();
        this.el.appendChild(container);
    }
}
function bindImagePickers() {
    var els = document.querySelectorAll(".admin_images");
    for (var i = 0; i < els.length; i++) {
        new ImagePicker(els[i]);
    }
}
class ImagePicker {
    constructor(el) {
        this.el = el;
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.hiddenInput = el.querySelector(".admin_images_hidden");
        this.preview = el.querySelector(".admin_images_preview");
        this.fileInput = this.el.querySelector(".admin_images_fileinput input");
        this.progress = this.el.querySelector("progress");
        this.el.querySelector(".admin_images_loaded").classList.remove("hidden");
        this.hideProgress();
        var ids = this.hiddenInput.value.split(",");
        this.el.addEventListener("click", (e) => {
            if (e.altKey) {
                var ids = window.prompt("IDs of images", this.hiddenInput.value);
                this.hiddenInput.value = ids;
                e.preventDefault();
                return false;
            }
        });
        this.fileInput.addEventListener("dragenter", (ev) => {
            this.fileInput.classList.add("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragleave", (ev) => {
            this.fileInput.classList.remove("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragover", (ev) => {
            ev.preventDefault();
        });
        this.fileInput.addEventListener("drop", (ev) => {
            var text = ev.dataTransfer.getData('Text');
            return;
        });
        for (var i = 0; i < ids.length; i++) {
            var id = ids[i];
            if (id) {
                this.addImage(id);
            }
        }
        this.fileInput.addEventListener("change", (e) => {
            var files = this.fileInput.files;
            var formData = new FormData();
            if (files.length == 0) {
                return;
            }
            for (var i = 0; i < files.length; i++) {
                formData.append("file", files[i]);
            }
            var request = new XMLHttpRequest();
            request.open("POST", this.adminPrefix + "/file/api/upload");
            request.addEventListener("load", (e) => {
                this.hideProgress();
                if (request.status == 200) {
                    var data = JSON.parse(request.response);
                    for (var i = 0; i < data.length; i++) {
                        this.addImage(data[i].UID);
                    }
                }
                else {
                    new Alert("Chyba při nahrávání souboru.");
                    console.error("Error while loading item.");
                }
            });
            this.showProgress();
            request.send(formData);
        });
    }
    updateHiddenData() {
        var ids = [];
        for (var i = 0; i < this.preview.children.length; i++) {
            var item = this.preview.children[i];
            var uuid = item.getAttribute("data-uuid");
            ids.push(uuid);
        }
        this.hiddenInput.value = ids.join(",");
    }
    addImage(id) {
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("data-uuid", id);
        container.setAttribute("draggable", "true");
        container.setAttribute("target", "_blank");
        container.setAttribute("href", this.adminPrefix + "/file/api/redirect-uuid/" + id);
        container.setAttribute("style", "background-image: url('" + this.adminPrefix + "/file/api/redirect-thumb/" + id + "');");
        var descriptionEl = document.createElement("div");
        descriptionEl.classList.add("admin_images_image_description");
        container.appendChild(descriptionEl);
        var request = new XMLHttpRequest();
        request.open("GET", this.adminPrefix + "/file/api/imagedata/" + id);
        request.addEventListener("load", (e) => {
            if (request.status == 200) {
                var data = JSON.parse(request.response);
                descriptionEl.innerText = data["Name"];
                container.setAttribute("title", data["Name"]);
            }
            else {
                console.error("Error while loading file metadata.");
            }
        });
        request.send();
        container.addEventListener("dragstart", (e) => {
            this.draggedElement = e.target;
        });
        container.addEventListener("drop", (e) => {
            var droppedElement = e.toElement;
            if (!droppedElement) {
                droppedElement = e.originalTarget;
            }
            for (var i = 0; i < 3; i++) {
                if (droppedElement.nodeName == "A") {
                    break;
                }
                else {
                    droppedElement = droppedElement.parentElement;
                }
            }
            var draggedIndex = -1;
            var droppedIndex = -1;
            var parent = this.draggedElement.parentElement;
            for (var i = 0; i < parent.children.length; i++) {
                var child = parent.children[i];
                if (child == this.draggedElement) {
                    draggedIndex = i;
                }
                if (child == droppedElement) {
                    droppedIndex = i;
                }
            }
            if (draggedIndex == -1 || droppedIndex == -1) {
                return;
            }
            if (draggedIndex <= droppedIndex) {
                droppedIndex += 1;
            }
            DOMinsertChildAtIndex(parent, this.draggedElement, droppedIndex);
            this.updateHiddenData();
            e.preventDefault();
            return false;
        });
        container.addEventListener("dragover", (e) => {
            e.preventDefault();
        });
        container.addEventListener("click", (e) => {
            var target = e.target;
            if (target.classList.contains("admin_images_image_delete")) {
                var parent = e.currentTarget.parentNode;
                parent.removeChild(e.currentTarget);
                this.updateHiddenData();
                e.preventDefault();
                return false;
            }
        });
        var del = document.createElement("div");
        del.textContent = "×";
        del.classList.add("admin_images_image_delete");
        container.appendChild(del);
        this.preview.appendChild(container);
        this.updateHiddenData();
    }
    hideProgress() {
        this.progress.classList.add("hidden");
    }
    showProgress() {
        this.progress.classList.remove("hidden");
    }
}
class ListFilterRelations {
    constructor(el, value, list) {
        this.valueInput = el.querySelector(".filter_relations_hidden");
        this.input = el.querySelector(".filter_relations_search_input");
        this.search = el.querySelector(".filter_relations_search");
        this.suggestions = el.querySelector(".filter_relations_suggestions");
        this.preview = el.querySelector(".filter_relations_preview");
        this.previewName = el.querySelector(".filter_relations_preview_name");
        this.previewClose = el.querySelector(".filter_relations_preview_close");
        this.previewClose.addEventListener("click", this.closePreview.bind(this));
        this.preview.classList.add("hidden");
        let hiddenEl = el.querySelector("input");
        this.relatedResourceName = el.querySelector(".admin_table_filter_item-relations").getAttribute("data-related-resource");
        this.input.addEventListener("input", () => {
            this.dirty = true;
            this.lastChanged = Date.now();
            return false;
        });
        window.setInterval(() => {
            if (this.dirty && Date.now() - this.lastChanged > 100) {
                this.loadSuggestions();
            }
        }, 30);
        if (this.valueInput.value) {
            this.loadPreview(this.valueInput.value);
        }
    }
    loadPreview(value) {
        var request = new XMLHttpRequest();
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        request.open("GET", adminPrefix + "/" + this.relatedResourceName + "/api/preview-relation/" + value, true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                this.renderPreview(JSON.parse(request.response));
            }
            else {
                console.error("not found");
            }
        });
        request.send();
    }
    renderPreview(item) {
        this.valueInput.value = item.ID;
        this.preview.classList.remove("hidden");
        this.search.classList.add("hidden");
        this.previewName.textContent = item.Name;
        this.dispatchChange();
    }
    dispatchChange() {
        var event = new Event('change');
        this.valueInput.dispatchEvent(event);
    }
    closePreview() {
        this.valueInput.value = "";
        this.preview.classList.add("hidden");
        this.search.classList.remove("hidden");
        this.input.value = "";
        this.suggestions.innerHTML = "";
        this.suggestions.classList.add("filter_relations_suggestions-empty");
        this.dispatchChange();
        this.input.focus();
    }
    loadSuggestions() {
        this.getSuggestions(this.input.value);
        this.dirty = false;
    }
    getSuggestions(q) {
        var request = new XMLHttpRequest();
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        request.open("GET", adminPrefix + "/" + this.relatedResourceName + "/api/searchresource" + "?q=" + encodeURIComponent(q), true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                this.renderSuggestions(JSON.parse(request.response));
            }
            else {
                console.error("not found");
            }
        });
        request.send();
    }
    renderSuggestions(data) {
        this.suggestions.innerHTML = "";
        this.suggestions.classList.add("filter_relations_suggestions-empty");
        for (var i = 0; i < data.length; i++) {
            this.suggestions.classList.remove("filter_relations_suggestions-empty");
            let item = data[i];
            let el = this.renderSuggestion(item);
            this.suggestions.appendChild(el);
            let index = i;
            el.addEventListener("mousedown", (e) => {
                this.renderPreview(item);
            });
        }
    }
    renderSuggestion(data) {
        var ret = document.createElement("div");
        ret.classList.add("list_filter_suggestion");
        ret.setAttribute("href", data.URL);
        var image = document.createElement("div");
        image.classList.add("list_filter_suggestion_image");
        image.setAttribute("style", "background-image: url('" + data.Image + "');");
        var right = document.createElement("div");
        right.classList.add("list_filter_suggestion_right");
        var name = document.createElement("div");
        name.classList.add("list_filter_suggestion_name");
        name.textContent = data.Name;
        var description = document.createElement("div");
        description.classList.add("list_filter_suggestion_description");
        description.textContent = data.Description;
        ret.appendChild(image);
        right.appendChild(name);
        right.appendChild(description);
        ret.appendChild(right);
        return ret;
    }
}
class ListFilterDate {
    constructor(el, value) {
        this.hidden = el.querySelector(".admin_table_filter_item");
        this.from = el.querySelector(".admin_filter_layout_date_from");
        this.to = el.querySelector(".admin_filter_layout_date_to");
        this.from.addEventListener("input", this.changed.bind(this));
        this.from.addEventListener("change", this.changed.bind(this));
        this.to.addEventListener("input", this.changed.bind(this));
        this.to.addEventListener("change", this.changed.bind(this));
        this.setValue(value);
    }
    setValue(value) {
        if (!value) {
            return;
        }
        var splited = value.split(",");
        if (splited.length == 2) {
            this.from.value = splited[0];
            this.to.value = splited[1];
        }
        this.hidden.value = value;
    }
    changed() {
        var val = "";
        if (this.from.value || this.to.value) {
            val = this.from.value + "," + this.to.value;
        }
        this.hidden.value = val;
        var event = new Event('change');
        this.hidden.dispatchEvent(event);
    }
}
function bindLists() {
    var els = document.getElementsByClassName("admin_list");
    for (var i = 0; i < els.length; i++) {
        new List(els[i], document.querySelector(".admin_tablesettings_buttons"));
    }
}
class List {
    constructor(el, openbutton) {
        this.el = el;
        this.settingsRow = document.querySelector(".admin_list_settingsrow");
        this.settingsRowColumn = document.querySelector(".admin_list_settingsrow_column");
        this.settingsEl = document.querySelector(".admin_tablesettings");
        this.settingsPopup = new ContentPopup("Možnosti", this.settingsEl);
        this.settingsButton = this.el.querySelector(".admin_list_settings");
        this.settingsButton.addEventListener("click", () => {
            this.settingsPopup.show();
        });
        this.exportButton = document.querySelector(".admin_exportbutton");
        let urlParams = new URLSearchParams(window.location.search);
        this.page = parseInt(urlParams.get("_page"));
        if (!this.page) {
            this.page = 1;
        }
        this.typeName = el.getAttribute("data-type");
        if (!this.typeName) {
            return;
        }
        this.progress = el.querySelector(".admin_table_progress");
        this.tbody = el.querySelector("tbody");
        this.tbody.textContent = "";
        this.bindFilter(urlParams);
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.defaultOrderColumn = el.getAttribute("data-order-column");
        if (el.getAttribute("data-order-desc") == "true") {
            this.defaultOrderDesc = true;
        }
        else {
            this.defaultOrderDesc = false;
        }
        this.orderColumn = this.defaultOrderColumn;
        this.orderDesc = this.defaultOrderDesc;
        if (urlParams.get("_order")) {
            this.orderColumn = urlParams.get("_order");
        }
        if (urlParams.get("_desc") == "true") {
            this.orderDesc = true;
        }
        if (urlParams.get("_desc") == "false") {
            this.orderDesc = false;
        }
        this.defaultVisibleColumnsStr = el.getAttribute("data-visible-columns");
        var visibleColumnsStr = this.defaultVisibleColumnsStr;
        if (urlParams.get("_columns")) {
            visibleColumnsStr = urlParams.get("_columns");
        }
        let visibleColumnsArr = visibleColumnsStr.split(",");
        let visibleColumnsMap = {};
        for (var i = 0; i < visibleColumnsArr.length; i++) {
            visibleColumnsMap[visibleColumnsArr[i]] = true;
        }
        this.itemsPerPage = parseInt(el.getAttribute("data-items-per-page"));
        this.paginationSelect = document.querySelector(".admin_tablesettings_pages");
        this.paginationSelect.addEventListener("change", this.load.bind(this));
        this.statsCheckbox = document.querySelector(".admin_tablesettings_stats");
        this.statsCheckbox.addEventListener("change", () => {
            this.filterChanged();
        });
        this.statsCheckboxSelectCount = document.querySelector(".admin_tablesettings_stats_limit");
        this.statsCheckboxSelectCount.addEventListener("change", () => {
            this.filterChanged();
        });
        this.statsContainer = document.querySelector(".admin_tablesettings_stats_container");
        if (this.hasMultipleActions()) {
            this.bindMultipleActions();
        }
        this.bindOptions(visibleColumnsMap);
        this.bindOrder();
    }
    hasMultipleActions() {
        if (this.el.classList.contains("admin_list-hasmultipleactions")) {
            return true;
        }
        return false;
    }
    bindMultipleActions() {
        var actions = this.el.querySelectorAll(".admin_list_multiple_action");
        for (var i = 0; i < actions.length; i++) {
            actions[i].addEventListener("click", this.multipleActionSelected.bind(this));
        }
    }
    multipleActionSelected(e) {
        var target = e.target;
        var actionName = target.getAttribute("name");
        switch (actionName) {
            case "cancel":
                this.multipleUncheckAll();
                break;
            case "delete":
                var ids = this.multipleGetIDs();
                new Confirm(`Opravdu chcete smazat ${ids.length} položek?`, () => {
                    var loader = new LoadingPopup();
                    var params = {};
                    params["action"] = "delete";
                    params["ids"] = ids.join(",");
                    var url = this.adminPrefix + "/" + this.typeName + "/api/multipleaction" + encodeParams(params);
                    fetch(url, {
                        method: "POST",
                    }).then((e) => {
                        loader.done();
                        if (e.status != 200) {
                            new Alert("Error while doing multipleaction delete");
                            return;
                        }
                        this.load();
                    });
                }, Function(), ButtonStyle.Delete);
                break;
            default:
                console.log("other");
        }
    }
    bindMultipleActionCheckboxes() {
        this.checkboxesAr = document.querySelectorAll(".admin_table_cell-multiple_checkbox");
        for (var i = 0; i < this.checkboxesAr.length; i++) {
            var checkbox = this.checkboxesAr[i];
            checkbox.addEventListener("change", this.multipleCheckboxChanged.bind(this));
        }
        this.multipleCheckboxChanged();
    }
    multipleGetIDs() {
        var ret = [];
        for (var i = 0; i < this.checkboxesAr.length; i++) {
            var checkbox = this.checkboxesAr[i];
            if (checkbox.checked) {
                ret.push(checkbox.getAttribute("data-id"));
            }
        }
        return ret;
    }
    multipleCheckboxChanged() {
        var checkedCount = 0;
        for (var i = 0; i < this.checkboxesAr.length; i++) {
            var checkbox = this.checkboxesAr[i];
            if (checkbox.checked) {
                checkedCount++;
            }
        }
        var multipleActionsPanel = this.el.querySelector(".admin_list_multiple_actions");
        if (checkedCount > 0) {
            multipleActionsPanel.classList.add("admin_list_multiple_actions-visible");
        }
        else {
            multipleActionsPanel.classList.remove("admin_list_multiple_actions-visible");
        }
        this.el.querySelector(".admin_list_multiple_actions_description").textContent = `Vybráno ${checkedCount} položek`;
    }
    multipleUncheckAll() {
        for (var i = 0; i < this.checkboxesAr.length; i++) {
            var checkbox = this.checkboxesAr[i];
            checkbox.checked = false;
        }
        this.multipleCheckboxChanged();
    }
    settingsCheckboxChange() {
        if (this.settingsCheckbox.checked) {
            this.settingsRow.classList.add("admin_list_settingsrow-visible");
        }
        else {
            this.settingsRow.classList.remove("admin_list_settingsrow-visible");
        }
    }
    load() {
        this.progress.classList.remove("admin_table_progress-inactive");
        var request = new XMLHttpRequest();
        var params = {};
        if (this.page > 1) {
            params["_page"] = this.page;
        }
        if (this.orderColumn != this.defaultOrderColumn) {
            params["_order"] = this.orderColumn;
        }
        if (this.orderDesc != this.defaultOrderDesc) {
            params["_desc"] = this.orderDesc + "";
        }
        var columns = this.getSelectedColumnsStr();
        if (columns != this.defaultVisibleColumnsStr) {
            params["_columns"] = columns;
        }
        let filterData = this.getFilterData();
        for (var k in filterData) {
            params[k] = filterData[k];
        }
        this.colorActiveFilterItems();
        let selectedPages = parseInt(this.paginationSelect.value);
        if (selectedPages != this.itemsPerPage) {
            params["_pagesize"] = selectedPages;
        }
        var encoded = encodeParams(params);
        window.history.replaceState(null, null, document.location.pathname + encoded);
        if (this.statsCheckbox.checked) {
            params["_stats"] = "true";
            params["_statslimit"] = this.statsCheckboxSelectCount.value;
        }
        params["_format"] = "xlsx";
        if (this.exportButton) {
            this.exportButton.setAttribute("href", this.adminPrefix + "/" + this.typeName + "/api/list" + encodeParams(params));
        }
        params["_format"] = "json";
        encoded = encodeParams(params);
        request.open("GET", this.adminPrefix + "/" + this.typeName + "/api/list" + encoded, true);
        request.addEventListener("load", () => {
            this.tbody.innerHTML = "";
            if (request.status == 200) {
                var response = JSON.parse(request.response);
                this.tbody.innerHTML = response.Content;
                var countStr = response.CountStr;
                this.el.querySelector(".admin_table_count").textContent = countStr;
                this.statsContainer.innerHTML = response.StatsStr;
                bindOrder();
                this.bindPagination();
                this.bindClick();
                if (this.hasMultipleActions()) {
                    this.bindMultipleActionCheckboxes();
                }
                this.tbody.classList.remove("admin_table_loading");
            }
            else {
                console.error("error while loading list");
            }
            this.progress.classList.add("admin_table_progress-inactive");
        });
        request.send(JSON.stringify({}));
    }
    bindOptions(visibleColumnsMap) {
        var columns = document.querySelectorAll(".admin_tablesettings_column");
        for (var i = 0; i < columns.length; i++) {
            let columnName = columns[i].getAttribute("data-column-name");
            if (visibleColumnsMap[columnName]) {
                columns[i].checked = true;
            }
            columns[i].addEventListener("change", () => {
                this.changedOptions();
            });
        }
        this.changedOptions();
    }
    changedOptions() {
        var columns = this.getSelectedColumnsMap();
        var headers = document.querySelectorAll(".admin_list_orderitem");
        for (var i = 0; i < headers.length; i++) {
            var name = headers[i].getAttribute("data-name");
            if (columns[name]) {
                headers[i].classList.remove("hidden");
            }
            else {
                headers[i].classList.add("hidden");
            }
        }
        var filters = document.querySelectorAll(".admin_list_filteritem");
        for (var i = 0; i < filters.length; i++) {
            var name = filters[i].getAttribute("data-name");
            if (columns[name]) {
                filters[i].classList.remove("hidden");
            }
            else {
                filters[i].classList.add("hidden");
            }
        }
        this.settingsRowColumn.setAttribute("colspan", Object.keys(columns).length + "");
        this.load();
    }
    colorActiveFilterItems() {
        let itemsToColor = this.getFilterData();
        var filterItems = this.el.querySelectorAll(".admin_list_filteritem");
        for (var i = 0; i < filterItems.length; i++) {
            var item = filterItems[i];
            let name = item.getAttribute("data-name");
            if (itemsToColor[name]) {
                item.classList.add("admin_list_filteritem-colored");
            }
            else {
                item.classList.remove("admin_list_filteritem-colored");
            }
        }
    }
    paginationChange(e) {
        var el = e.target;
        var page = parseInt(el.getAttribute("data-page"));
        this.page = page;
        this.load();
        e.preventDefault();
        return false;
    }
    bindPagination() {
        var paginationEl = this.el.querySelector(".pagination");
        var totalPages = parseInt(paginationEl.getAttribute("data-total"));
        var selectedPage = parseInt(paginationEl.getAttribute("data-selected"));
        for (var i = 1; i <= totalPages; i++) {
            var pEl = document.createElement("a");
            pEl.setAttribute("href", "#");
            pEl.textContent = i + "";
            if (i == selectedPage) {
                pEl.classList.add("pagination_page_current");
            }
            else {
                pEl.classList.add("pagination_page");
                pEl.setAttribute("data-page", i + "");
                pEl.addEventListener("click", this.paginationChange.bind(this));
            }
            paginationEl.appendChild(pEl);
        }
    }
    bindClick() {
        var rows = this.el.querySelectorAll(".admin_table_row");
        for (var i = 0; i < rows.length; i++) {
            var row = rows[i];
            var id = row.getAttribute("data-id");
            row.addEventListener("click", (e) => {
                if (e.target.classList.contains("admin_table_cell-multiple_checkbox")) {
                    return false;
                }
                var target = e.target;
                if (target.classList.contains("preventredirect")) {
                    return;
                }
                var el = e.currentTarget;
                var url = el.getAttribute("data-url");
                if (e.shiftKey || e.metaKey || e.ctrlKey) {
                    var openedWindow = window.open(url, "newwindow" + (new Date()));
                    openedWindow.focus();
                    return;
                }
                window.location.href = url;
            });
            var buttons = row.querySelector(".admin_list_buttons");
            buttons.addEventListener("click", (e) => {
                var url = e.target.getAttribute("href");
                if (url != "") {
                    window.location.href = url;
                    e.preventDefault();
                    e.stopPropagation();
                    return false;
                }
            });
        }
    }
    bindOrder() {
        this.renderOrder();
        var headers = this.el.querySelectorAll(".admin_list_orderitem-canorder");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.addEventListener("click", (e) => {
                var el = e.target;
                var name = el.getAttribute("data-name");
                if (name == this.orderColumn) {
                    if (this.orderDesc) {
                        this.orderDesc = false;
                    }
                    else {
                        this.orderDesc = true;
                    }
                }
                else {
                    this.orderColumn = name;
                    this.orderDesc = false;
                }
                this.renderOrder();
                this.load();
                e.preventDefault();
                return false;
            });
        }
    }
    renderOrder() {
        var headers = this.el.querySelectorAll(".admin_list_orderitem-canorder");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.classList.remove("ordered");
            header.classList.remove("ordered-desc");
            var name = header.getAttribute("data-name");
            if (name == this.orderColumn) {
                header.classList.add("ordered");
                if (this.orderDesc) {
                    header.classList.add("ordered-desc");
                }
            }
        }
    }
    getSelectedColumnsStr() {
        var ret = [];
        var checked = document.querySelectorAll(".admin_tablesettings_column:checked");
        for (var i = 0; i < checked.length; i++) {
            ret.push(checked[i].getAttribute("data-column-name"));
        }
        return ret.join(",");
    }
    getSelectedColumnsMap() {
        var columns = {};
        var checked = document.querySelectorAll(".admin_tablesettings_column:checked");
        for (var i = 0; i < checked.length; i++) {
            columns[checked[i].getAttribute("data-column-name")] = true;
        }
        return columns;
    }
    getFilterData() {
        var ret = {};
        var items = this.el.querySelectorAll(".admin_table_filter_item");
        for (var i = 0; i < items.length; i++) {
            var item = items[i];
            var typ = item.getAttribute("data-typ");
            var layout = item.getAttribute("data-filter-layout");
            if (item.classList.contains("admin_table_filter_item-relations")) {
                ret[typ] = item.querySelector("input").value;
            }
            else {
                var val = item.value.trim();
                if (val) {
                    ret[typ] = val;
                }
            }
        }
        return ret;
    }
    bindFilter(params) {
        var filterFields = this.el.querySelectorAll(".admin_list_filteritem");
        for (var i = 0; i < filterFields.length; i++) {
            var field = filterFields[i];
            var fieldName = field.getAttribute("data-name");
            var fieldLayout = field.getAttribute("data-filter-layout");
            var fieldInput = field.querySelector("input");
            var fieldSelect = field.querySelector("select");
            var fieldValue = params.get(fieldName);
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
            if (fieldLayout == "filter_layout_relation") {
                this.bindFilterRelation(field, fieldValue);
            }
            if (fieldLayout == "filter_layout_date") {
                this.bindFilterDate(field, fieldValue);
            }
        }
        this.inputPeriodicListener();
    }
    inputListener(e) {
        if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
            return;
        }
        this.filterChanged();
    }
    filterChanged() {
        this.colorActiveFilterItems();
        this.tbody.classList.add("admin_table_loading");
        this.page = 1;
        this.changed = true;
        this.changedTimestamp = Date.now();
        this.progress.classList.remove("admin_table_progress-inactive");
    }
    bindFilterRelation(el, value) {
        new ListFilterRelations(el, value, this);
    }
    bindFilterDate(el, value) {
        new ListFilterDate(el, value);
    }
    inputPeriodicListener() {
        setInterval(() => {
            if (this.changed == true && Date.now() - this.changedTimestamp > 500) {
                this.changed = false;
                this.load();
            }
        }, 200);
    }
}
function bindOrder() {
    function orderTable(el) {
        var rows = el.getElementsByClassName("admin_table_row");
        Array.prototype.forEach.call(rows, function (item, i) {
            bindDraggable(item);
        });
        var draggedElement;
        function bindDraggable(row) {
            row.setAttribute("draggable", "true");
            row.addEventListener("dragstart", function (ev) {
                row.classList.add("admin_table_row-selected");
                draggedElement = this;
                ev.dataTransfer.setData('text/plain', '');
                ev.dataTransfer.effectAllowed = "move";
                var d = document.createElement("div");
                d.style.display = "none";
                ev.dataTransfer.setDragImage(d, 0, 0);
            });
            row.addEventListener("dragenter", function (ev) {
                var targetEl = this;
                if (this != draggedElement) {
                    var draggedIndex = -1;
                    var thisIndex = -1;
                    Array.prototype.forEach.call(el.getElementsByClassName("admin_table_row"), function (item, i) {
                        if (item == draggedElement) {
                            draggedIndex = i;
                        }
                        if (item == targetEl) {
                            thisIndex = i;
                        }
                    });
                    if (draggedIndex < thisIndex) {
                        thisIndex += 1;
                    }
                    DOMinsertChildAtIndex(targetEl.parentElement, draggedElement, thisIndex);
                }
                return false;
            });
            row.addEventListener("drop", function (ev) {
                saveOrder();
                row.classList.remove("admin_table_row-selected");
                return false;
            });
            row.addEventListener("dragover", function (ev) {
                ev.preventDefault();
            });
        }
        function saveOrder() {
            var adminPrefix = document.body.getAttribute("data-admin-prefix");
            var typ = document.querySelector(".admin_list-order").getAttribute("data-type");
            var ajaxPath = adminPrefix + "/" + typ + "/api/set-order";
            var order = [];
            var rows = el.getElementsByClassName("admin_table_row");
            Array.prototype.forEach.call(rows, function (item, i) {
                order.push(parseInt(item.getAttribute("data-id")));
            });
            var request = new XMLHttpRequest();
            request.open("POST", ajaxPath, true);
            request.addEventListener("load", () => {
                if (request.status != 200) {
                    console.error("Error while saving order.");
                }
            });
            request.send(JSON.stringify({ "order": order }));
        }
    }
    var elements = document.querySelectorAll(".admin_list-order");
    Array.prototype.forEach.call(elements, function (el, i) {
        orderTable(el);
    });
}
function bindMarkdowns() {
    var elements = document.querySelectorAll(".admin_markdown");
    Array.prototype.forEach.call(elements, function (el, i) {
        new MarkdownEditor(el);
    });
}
class MarkdownEditor {
    constructor(el) {
        this.el = el;
        this.textarea = el.querySelector(".textarea");
        this.preview = el.querySelector(".admin_markdown_preview");
        new Autoresize(this.textarea);
        var prefix = document.body.getAttribute("data-admin-prefix");
        var helpLink = el.querySelector(".admin_markdown_show_help");
        helpLink.setAttribute("href", prefix + "/markdown");
        this.lastChanged = Date.now();
        this.changed = false;
        let showChange = el.querySelector(".admin_markdown_preview_show");
        showChange.addEventListener("change", () => {
            this.preview.classList.toggle("hidden");
        });
        setInterval(() => {
            if (this.changed && (Date.now() - this.lastChanged > 500)) {
                this.loadPreview();
            }
        }, 100);
        this.textarea.addEventListener("change", this.textareaChanged.bind(this));
        this.textarea.addEventListener("keyup", this.textareaChanged.bind(this));
        this.loadPreview();
        this.bindCommands();
        this.bindShortcuts();
    }
    bindCommands() {
        var btns = this.el.querySelectorAll(".admin_markdown_command");
        for (var i = 0; i < btns.length; i++) {
            btns[i].addEventListener("mousedown", (e) => {
                var cmd = e.target.getAttribute("data-cmd");
                this.executeCommand(cmd);
                e.preventDefault();
                return false;
            });
        }
    }
    bindShortcuts() {
        this.textarea.addEventListener("keydown", (e) => {
            if (e.metaKey == false && e.ctrlKey == false) {
                return;
            }
            switch (e.keyCode) {
                case 66:
                    this.executeCommand("b");
                    break;
                case 73:
                    this.executeCommand("i");
                    break;
                case 75:
                    this.executeCommand("h2");
                    break;
                case 85:
                    this.executeCommand("a");
                    break;
            }
        });
    }
    executeCommand(commandName) {
        switch (commandName) {
            case "b":
                this.setAroundMarkdown("**", "**");
                break;
            case "i":
                this.setAroundMarkdown("*", "*");
                break;
            case "a":
                this.setAroundMarkdown("[", "]()");
                var newEnd = this.textarea.selectionEnd + 2;
                this.textarea.selectionStart = newEnd;
                this.textarea.selectionEnd = newEnd;
                break;
            case "h2":
                var start = "## ";
                var end = "";
                var text = this.textarea.value;
                if (text[this.textarea.selectionStart - 1] !== "\n") {
                    start = "\n" + start;
                }
                if (text[this.textarea.selectionEnd] !== "\n") {
                    end = "\n";
                }
                this.setAroundMarkdown(start, end);
                break;
        }
        this.textareaChanged();
    }
    setAroundMarkdown(before, after) {
        var text = this.textarea.value;
        var selected = text.substr(this.textarea.selectionStart, this.textarea.selectionEnd - this.textarea.selectionStart);
        var newText = text.substr(0, this.textarea.selectionStart);
        newText += before;
        var newStart = newText.length;
        newText += selected;
        var newEnd = newText.length;
        newText += after;
        newText += text.substr(this.textarea.selectionEnd, text.length);
        this.textarea.value = newText;
        this.textarea.selectionStart = newStart;
        this.textarea.selectionEnd = newEnd;
        this.textarea.focus();
    }
    textareaChanged() {
        this.changed = true;
        this.lastChanged = Date.now();
    }
    loadPreview() {
        this.changed = false;
        var request = new XMLHttpRequest();
        request.open("POST", document.body.getAttribute("data-admin-prefix") + "/api/markdown", true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                this.preview.innerHTML = JSON.parse(request.response);
            }
            else {
                console.error("Error while loading markdown preview.");
            }
        });
        request.send(this.textarea.value);
    }
}
function bindTimestamps() {
    var elements = document.querySelectorAll(".admin_timestamp");
    Array.prototype.forEach.call(elements, function (el, i) {
        new Timestamp(el);
    });
}
class Timestamp {
    constructor(el) {
        this.elTsInput = el.getElementsByTagName("input")[0];
        this.elTsDate = el.getElementsByClassName("admin_timestamp_date")[0];
        this.elTsHour = el.getElementsByClassName("admin_timestamp_hour")[0];
        this.elTsMinute = el.getElementsByClassName("admin_timestamp_minute")[0];
        this.initClock();
        var v = this.elTsInput.value;
        this.setTimestamp(v);
        this.elTsDate.addEventListener("change", this.saveValue.bind(this));
        this.elTsHour.addEventListener("change", this.saveValue.bind(this));
        this.elTsMinute.addEventListener("change", this.saveValue.bind(this));
        this.saveValue();
    }
    setTimestamp(v) {
        if (v == "") {
            return;
        }
        var date = v.split(" ")[0];
        var hour = parseInt(v.split(" ")[1].split(":")[0]);
        var minute = parseInt(v.split(" ")[1].split(":")[1]);
        this.elTsDate.value = date;
        var minuteOption = this.elTsMinute.children[minute];
        minuteOption.selected = true;
        var hourOption = this.elTsHour.children[hour];
        hourOption.selected = true;
    }
    initClock() {
        for (var i = 0; i < 24; i++) {
            var newEl = document.createElement("option");
            var addVal = "" + i;
            if (i < 10) {
                addVal = "0" + addVal;
            }
            newEl.innerText = addVal;
            newEl.setAttribute("value", addVal);
            this.elTsHour.appendChild(newEl);
        }
        for (var i = 0; i < 60; i++) {
            var newEl = document.createElement("option");
            var addVal = "" + i;
            if (i < 10) {
                addVal = "0" + addVal;
            }
            newEl.innerText = addVal;
            newEl.setAttribute("value", addVal);
            this.elTsMinute.appendChild(newEl);
        }
    }
    saveValue() {
        var str = this.elTsDate.value + " " + this.elTsHour.value + ":" + this.elTsMinute.value;
        if (this.elTsDate.value == "") {
            str = "";
        }
        this.elTsInput.value = str;
    }
}
function bindRelations() {
    var elements = document.querySelectorAll(".admin_item_relation");
    Array.prototype.forEach.call(elements, function (el, i) {
        new RelationPicker(el);
    });
}
class RelationPicker {
    constructor(el) {
        this.selectedClass = "admin_item_relation_picker_suggestion-selected";
        this.input = el.getElementsByTagName("input")[0];
        this.previewContainer = el.querySelector(".admin_item_relation_preview");
        this.relationName = el.getAttribute("data-relation");
        this.progress = el.querySelector("progress");
        this.changeSection = el.querySelector(".admin_item_relation_change");
        this.changeButton = el.querySelector(".admin_item_relation_change_btn");
        this.changeButton.addEventListener("click", () => {
            this.input.value = "0";
            this.showSearch();
            this.pickerInput.focus();
        });
        this.suggestionsEl = el.querySelector(".admin_item_relation_picker_suggestions_content");
        this.suggestions = [];
        this.picker = el.querySelector(".admin_item_relation_picker");
        this.pickerInput = this.picker.querySelector("input");
        this.pickerInput.addEventListener("input", () => {
            this.getSuggestions(this.pickerInput.value);
        });
        this.pickerInput.addEventListener("blur", () => {
            this.suggestionsEl.classList.add("hidden");
        });
        this.pickerInput.addEventListener("focus", () => {
            this.suggestionsEl.classList.remove("hidden");
            this.getSuggestions(this.pickerInput.value);
        });
        this.pickerInput.addEventListener("keydown", this.suggestionInput.bind(this));
        if (this.input.value != "0") {
            this.getData();
        }
        else {
            this.progress.classList.add("hidden");
            this.showSearch();
        }
    }
    getData() {
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/" + this.relationName + "/api/preview-relation/" + this.input.value, true);
        request.addEventListener("load", () => {
            this.progress.classList.add("hidden");
            if (request.status == 200) {
                this.showPreview(JSON.parse(request.response));
            }
            else {
                this.showSearch();
            }
        });
        request.send();
    }
    showPreview(data) {
        this.previewContainer.textContent = "";
        this.input.value = data.ID;
        var el = this.createPreview(data, true);
        this.previewContainer.appendChild(el);
        this.previewContainer.classList.remove("hidden");
        this.changeSection.classList.remove("hidden");
        this.picker.classList.add("hidden");
    }
    showSearch() {
        this.previewContainer.classList.add("hidden");
        this.changeSection.classList.add("hidden");
        this.picker.classList.remove("hidden");
        this.suggestions = [];
        this.suggestionsEl.innerText = "";
        this.pickerInput.value = "";
    }
    getSuggestions(q) {
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/" + this.relationName + "/api/searchresource" + "?q=" + encodeURIComponent(q), true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                if (q != this.pickerInput.value) {
                    return;
                }
                var data = JSON.parse(request.response);
                this.suggestions = data;
                this.suggestionsEl.innerText = "";
                for (var i = 0; i < data.length; i++) {
                    var item = data[i];
                    var el = this.createPreview(item, false);
                    el.classList.add("admin_item_relation_picker_suggestion");
                    el.setAttribute("data-position", i + "");
                    el.addEventListener("mousedown", this.suggestionClick.bind(this));
                    el.addEventListener("mouseenter", this.suggestionSelect.bind(this));
                    this.suggestionsEl.appendChild(el);
                }
            }
            else {
                console.log("Error while searching");
            }
        });
        request.send();
    }
    suggestionClick() {
        var selected = this.getSelected();
        if (selected >= 0) {
            this.showPreview(this.suggestions[selected]);
        }
    }
    suggestionSelect(e) {
        var target = e.currentTarget;
        var position = parseInt(target.getAttribute("data-position"));
        this.select(position);
    }
    getSelected() {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        return parseInt(selected.getAttribute("data-position"));
    }
    unselect() {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        selected.classList.remove(this.selectedClass);
        return parseInt(selected.getAttribute("data-position"));
    }
    select(i) {
        this.unselect();
        this.suggestionsEl.querySelectorAll(".admin_preview")[i].classList.add(this.selectedClass);
    }
    suggestionInput(e) {
        switch (e.keyCode) {
            case 13:
                this.suggestionClick();
                e.preventDefault();
                return true;
            case 38:
                var i = this.getSelected();
                if (i < 1) {
                    i = this.suggestions.length - 1;
                }
                else {
                    i = i - 1;
                }
                this.select(i);
                e.preventDefault();
                return false;
            case 40:
                var i = this.getSelected();
                if (i >= 0) {
                    i += 1;
                    i = i % this.suggestions.length;
                }
                else {
                    i = 0;
                }
                this.select(i);
                e.preventDefault();
                return false;
        }
    }
    createPreview(data, anchor) {
        var ret = document.createElement("div");
        if (anchor) {
            ret = document.createElement("a");
        }
        ret.classList.add("admin_preview");
        ret.setAttribute("href", data.URL);
        var image = document.createElement("div");
        image.classList.add("admin_preview_image");
        image.setAttribute("style", "background-image: url('" + data.Image + "');");
        var right = document.createElement("div");
        right.classList.add("admin_preview_right");
        var name = document.createElement("div");
        name.classList.add("admin_preview_name");
        name.textContent = data.Name;
        var description = document.createElement("description");
        description.classList.add("admin_preview_description");
        description.textContent = data.Description;
        ret.appendChild(image);
        right.appendChild(name);
        right.appendChild(description);
        ret.appendChild(right);
        return ret;
    }
}
function bindPlacesView() {
    var els = document.querySelectorAll(".admin_item_view_place");
    for (var i = 0; i < els.length; i++) {
        new PlacesView(els[i]);
    }
}
class PlacesView {
    constructor(el) {
        var val = el.getAttribute("data-value");
        el.innerText = "";
        var coords = val.split(",");
        if (coords.length != 2) {
            el.classList.remove("admin_item_view_place");
            return;
        }
        var position = { lat: parseFloat(coords[0]), lng: parseFloat(coords[1]) };
        var zoom = 18;
        var map = new google.maps.Map(el, {
            center: position,
            zoom: zoom
        });
        var marker = new google.maps.Marker({
            position: position,
            map: map
        });
    }
}
function bindPlaces() {
    bindPlacesView();
    function bindPlace(el) {
        var mapEl = document.createElement("div");
        mapEl.classList.add("admin_place_map");
        el.appendChild(mapEl);
        var position = { lat: 0, lng: 0 };
        var zoom = 1;
        var visible = false;
        var input = el.getElementsByTagName("input")[0];
        var inVal = input.value;
        var inVals = inVal.split(",");
        if (inVals.length == 2) {
            inVals[0] = parseFloat(inVals[0]);
            inVals[1] = parseFloat(inVals[1]);
            if (!isNaN(inVals[0]) && !isNaN(inVals[1])) {
                position.lat = inVals[0];
                position.lng = inVals[1];
                zoom = 11;
                visible = true;
            }
        }
        var map = new google.maps.Map(mapEl, {
            center: position,
            zoom: zoom
        });
        var marker = new google.maps.Marker({
            position: position,
            map: map,
            draggable: true,
            title: "",
            visible: visible
        });
        var searchInput = document.createElement("input");
        searchInput.classList.add("input", "input-placesearch");
        var searchBox = new google.maps.places.SearchBox(searchInput);
        map.controls[google.maps.ControlPosition.LEFT_TOP].push(searchInput);
        searchBox.addListener('places_changed', () => {
            var places = searchBox.getPlaces();
            if (places.length > 0) {
                map.fitBounds(places[0].geometry.viewport);
                marker.setPosition({ lat: places[0].geometry.location.lat(), lng: places[0].geometry.location.lng() });
                marker.setVisible(true);
            }
        });
        searchInput.addEventListener("keydown", (e) => {
            if (e.keyCode == 13) {
                e.preventDefault();
                return false;
            }
        });
        marker.addListener("position_changed", function () {
            var p = marker.getPosition();
            var str = stringifyPosition(p.lat(), p.lng());
            input.value = str;
        });
        marker.addListener("click", function () {
            marker.setVisible(false);
            input.value = "";
        });
        map.addListener('click', function (e) {
            position.lat = e.latLng.lat();
            position.lng = e.latLng.lng();
            marker.setPosition(position);
            marker.setVisible(true);
        });
        function stringifyPosition(lat, lng) {
            return lat + "," + lng;
        }
    }
    var elements = document.querySelectorAll(".admin_place");
    Array.prototype.forEach.call(elements, function (el, i) {
        bindPlace(el);
    });
}
function bindForm() {
    var els = document.querySelectorAll(".form_leavealert");
    for (var i = 0; i < els.length; i++) {
        new Form(els[i]);
    }
}
class Form {
    constructor(el) {
        this.dirty = false;
        el.addEventListener("submit", () => {
            this.dirty = false;
        });
        let els = el.querySelectorAll(".form_watcher");
        for (var i = 0; i < els.length; i++) {
            var input = els[i];
            input.addEventListener("input", () => {
                this.dirty = true;
            });
            input.addEventListener("change", () => {
                this.dirty = true;
            });
        }
        window.addEventListener("beforeunload", (e) => {
            if (this.dirty) {
                var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
                e.returnValue = confirmationMessage;
                return confirmationMessage;
            }
        });
    }
}
function bindDatePicker() {
    var dates = document.querySelectorAll(".form_input-date");
    for (var i = 0; i < dates.length; i++) {
        var dateEl = dates[i];
        new DatePicker(dateEl);
    }
}
class DatePicker {
    constructor(el) {
        var language = "cs";
        var i18n = {
            previousMonth: 'Previous Month',
            nextMonth: 'Next Month',
            months: ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"],
            weekdays: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],
            weekdaysShort: ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']
        };
        if (language == "de") {
            i18n = {
                previousMonth: 'Vorheriger Monat',
                nextMonth: 'Nächsten Monat',
                months: ["Januar", "Februar", "März", "April", "Kann", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"],
                weekdays: ['Sonntag', 'Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag'],
                weekdaysShort: ['So', 'Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa']
            };
        }
        if (language == "ru") {
            var i18n = {
                previousMonth: 'Предыдущий месяц',
                nextMonth: 'В следующем месяце',
                months: ["Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"],
                weekdays: ["Воскресенье", "Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота"],
                weekdaysShort: ['Во', 'По', 'Вт', 'Ср', 'Че', 'Пя', 'Су']
            };
        }
        if (language == "cs") {
            i18n = {
                previousMonth: 'Předchozí měsíc',
                nextMonth: 'Další měsíc',
                months: ["Leden", "Únor", "Březen", "Duben", "Květen", "Červen", "Červenec", "Srpen", "Září", "Říjen", "Listopad", "Prosinec"],
                weekdays: ['Neděle', 'Pondělí', 'Úterý', 'Středa', 'Čtvrtek', 'Pátek', 'Sobota'],
                weekdaysShort: ['Ne', 'Po', 'Út', 'St', 'Čt', 'Pá', 'So']
            };
        }
        var self = this;
        var pd = new Pikaday({
            field: el,
            setDefaultDate: false,
            i18n: i18n,
            onSelect: (date) => {
                el.value = pd.toString();
            },
            toString: (date) => {
                const day = date.getDate();
                var dayStr = "" + day;
                if (day < 10) {
                    dayStr = "0" + dayStr;
                }
                const month = date.getMonth() + 1;
                var monthStr = "" + month;
                if (month < 10) {
                    monthStr = "0" + monthStr;
                }
                const year = date.getFullYear();
                var ret = `${year}-${monthStr}-${dayStr}`;
                return ret;
            }
        });
    }
}
function prettyDate(date) {
    const day = date.getDate();
    const month = date.getMonth() + 1;
    const year = date.getFullYear();
    return `${day}. ${month}. ${year}`;
}
function bindDropdowns() {
    var els = document.querySelectorAll(".admin_dropdown");
    for (var i = 0; i < els.length; i++) {
        new Dropdown(els[i]);
    }
}
class Dropdown {
    constructor(el) {
        this.targetEl = el.querySelector(".admin_dropdown_target");
        this.contentEl = el.querySelector(".admin_dropdown_content");
        this.targetEl.addEventListener("mousedown", (e) => {
            if (document.activeElement == el) {
                el.blur();
                e.preventDefault();
                return false;
            }
        });
    }
}
function bindSearch() {
    var els = document.querySelectorAll(".admin_header_search");
    for (var i = 0; i < els.length; i++) {
        new SearchForm(els[i]);
    }
}
class SearchForm {
    constructor(el) {
        this.searchForm = el;
        this.searchInput = el.querySelector(".admin_header_search_input");
        this.suggestionsEl = el.querySelector(".admin_header_search_suggestions");
        this.searchInput.value = document.body.getAttribute("data-search-query");
        this.searchInput.addEventListener("input", () => {
            this.suggestions = [];
            this.dirty = true;
            this.lastChanged = Date.now();
            return false;
        });
        this.searchInput.addEventListener("blur", () => {
        });
        window.setInterval(() => {
            if (this.dirty && Date.now() - this.lastChanged > 100) {
                this.loadSuggestions();
            }
        }, 30);
        this.searchInput.addEventListener("keydown", (e) => {
            if (!this.suggestions || this.suggestions.length == 0) {
                return;
            }
            switch (e.keyCode) {
                case 13:
                    var i = this.getSelected();
                    if (i >= 0) {
                        var child = this.suggestions[i];
                        if (child) {
                            window.location.href = child.getAttribute("href");
                        }
                        e.preventDefault();
                        return true;
                    }
                    return false;
                case 38:
                    var i = this.getSelected();
                    if (i < 1) {
                        i = this.suggestions.length - 1;
                    }
                    else {
                        i = i - 1;
                    }
                    this.setSelected(i);
                    e.preventDefault();
                    return false;
                case 40:
                    var i = this.getSelected();
                    if (i >= 0) {
                        i += 1;
                        i = i % this.suggestions.length;
                    }
                    else {
                        i = 0;
                    }
                    this.setSelected(i);
                    e.preventDefault();
                    return false;
            }
        });
    }
    loadSuggestions() {
        this.dirty = false;
        var suggestText = this.searchInput.value;
        var request = new XMLHttpRequest();
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var url = adminPrefix + "/api/search-suggest" + encodeParams({ "q": this.searchInput.value });
        request.open("GET", url);
        request.addEventListener("load", () => {
            if (suggestText != this.searchInput.value) {
                return;
            }
            if (request.status == 200) {
                this.addSuggestions(request.response);
            }
            else {
                this.suggestionsEl.classList.add("hidden");
                console.error("Error while loading item.");
            }
        });
        request.send();
    }
    addSuggestions(content) {
        this.suggestionsEl.innerHTML = content;
        this.suggestionsEl.classList.remove("hidden");
        this.suggestions = this.suggestionsEl.querySelectorAll(".admin_search_suggestion");
        for (var i = 0; i < this.suggestions.length; i++) {
            var suggestion = this.suggestions[i];
            suggestion.addEventListener("touchend", (e) => {
                var el = e.currentTarget;
                window.location.href = el.getAttribute("href");
            });
            suggestion.addEventListener("click", (e) => {
                return false;
            });
            suggestion.addEventListener("mouseenter", (e) => {
                this.deselect();
                var el = e.currentTarget;
                this.setSelected(parseInt(el.getAttribute("data-position")));
            });
        }
    }
    deselect() {
        var el = this.suggestionsEl.querySelector(".admin_search_suggestion-selected");
        if (el) {
            el.classList.remove("admin_search_suggestion-selected");
        }
    }
    getSelected() {
        var el = this.suggestionsEl.querySelector(".admin_search_suggestion-selected");
        if (el) {
            return parseInt(el.getAttribute("data-position"));
        }
        return -1;
    }
    setSelected(position) {
        this.deselect();
        if (position >= 0) {
            var els = this.suggestionsEl.querySelectorAll(".admin_search_suggestion");
            els[position].classList.add("admin_search_suggestion-selected");
        }
    }
}
function bindMainMenu() {
    var el = document.querySelector(".admin_layout_left");
    if (el) {
        new MainMenu(el);
    }
}
class MainMenu {
    constructor(leftEl) {
        this.leftEl = leftEl;
        this.menuEl = document.querySelector(".admin_header_container_menu");
        this.menuEl.addEventListener("click", this.menuClick.bind(this));
        this.scrollTo(this.loadFromStorage());
        this.leftEl.addEventListener("scroll", this.scrollHandler.bind(this));
    }
    scrollHandler() {
        this.saveToStorage(this.leftEl.scrollTop);
    }
    saveToStorage(position) {
        window.localStorage["left_menu_position"] = position;
    }
    menuClick() {
        this.leftEl.classList.toggle("admin_layout_left-visible");
    }
    loadFromStorage() {
        var pos = window.localStorage["left_menu_position"];
        if (pos) {
            return parseInt(pos);
        }
        return 0;
    }
    scrollTo(position) {
        this.leftEl.scrollTo(0, position);
    }
}
function bindRelationList() {
    var els = document.getElementsByClassName("admin_relationlist");
    for (var i = 0; i < els.length; i++) {
        new RelationList(els[i]);
    }
}
class RelationList {
    constructor(el) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.targetEl = el.querySelector(".admin_relationlist_target");
        this.sourceResource = el.getAttribute("data-source-resource");
        this.targetResource = el.getAttribute("data-target-resource");
        this.targetField = el.getAttribute("data-target-field");
        this.idValue = parseInt(el.getAttribute("data-id-value"));
        this.count = parseInt(el.getAttribute("data-count"));
        this.offset = 0;
        this.loadingEl = el.querySelector(".admin_relationlist_loading");
        this.moreEl = el.querySelector(".admin_relationlist_more");
        this.moreButton = el.querySelector(".admin_relationlist_more .btn");
        this.moreButton.addEventListener("click", this.load.bind(this));
        this.load();
    }
    load() {
        this.loadingEl.classList.remove("hidden");
        this.moreEl.classList.add("hidden");
        var request = new XMLHttpRequest();
        request.open("POST", this.adminPrefix + "/api/relationlist", true);
        request.addEventListener("load", () => {
            this.loadingEl.classList.add("hidden");
            if (request.status == 200) {
                this.offset += 10;
                var parentEl = document.createElement("div");
                parentEl.innerHTML = request.response;
                var parentAr = [];
                for (var i = 0; i < parentEl.children.length; i++) {
                    parentAr.push(parentEl.children[i]);
                }
                for (var i = 0; i < parentAr.length; i++) {
                    this.targetEl.appendChild(parentAr[i]);
                }
                if (this.offset < this.count) {
                    this.moreEl.classList.remove("hidden");
                }
            }
            else {
                console.error("Error while RelationList request");
            }
        });
        request.send(JSON.stringify({
            SourceResource: this.sourceResource,
            TargetResource: this.targetResource,
            TargetField: this.targetField,
            IDValue: this.idValue,
            Offset: this.offset,
            Count: 10,
        }));
    }
}
function bindTaskMonitor() {
    var el = document.querySelector(".taskmonitorcontainer");
    if (el) {
        new TaskMonitor(el);
    }
}
class TaskMonitor {
    constructor(el) {
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
            }
            else {
                console.error("error while loading list");
            }
        });
        request.send();
    }
}
function bindNotifications() {
    new NotificationCenter2(document.querySelector(".notification_center"));
}
class NotificationCenter2 {
    constructor(el) {
        var notifications = el.querySelectorAll(".notification");
        for (var i = 0; i < notifications.length; i++) {
            this.bindNotification(notifications[i]);
        }
    }
    bindNotification(el) {
        el.querySelector(".notification_close").addEventListener("click", () => {
            el.classList.add("notification-closed");
        });
    }
}
class NotificationCenter {
    constructor(el) {
        this.el = el;
        this.notifications = Array();
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.loadNotifications();
    }
    loadNotifications() {
        return;
        var request = new XMLHttpRequest();
        request.open("GET", this.adminPrefix + "/_api/notifications", true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                var notifications = JSON.parse(request.response);
                notifications.Views.forEach((item) => {
                    var notification = this.createNotification(item);
                    this.notifications.push(notification);
                    this.el.appendChild(notification.el);
                });
            }
            else {
                console.log("failed to load notifications");
            }
        });
        request.send();
    }
    createNotification(data) {
        return new NotificationItem(data);
    }
}
class NotificationItem {
    constructor(data) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.createElement(data);
    }
    createElement(data) {
        var ret;
        ret = document.createElement("div");
        ret.innerHTML = `
            <div class="notification">
                <div class="notification_close"></div>
                <div class="notification_left"></div>
                <div class="notification_right">
                    <div class="notification_name">${e(data.Name)}</div>
                </div>
            </div>        
        `;
        this.el = ret.children[0];
        this.el.querySelector(".notification_close").addEventListener("click", this.closeNotification.bind(this));
    }
    closeNotification() {
        this.el.classList.add("notification-closed");
        fetch(this.adminPrefix + "/_api/notification/" + this.uuid, { method: "DELETE" }).then(console.log).then((e) => {
            console.log(e);
        });
    }
}
class Popup {
    constructor(title) {
        this.el = document.createElement("div");
        this.el.classList.add("popup_background");
        document.body.appendChild(this.el);
        this.el.innerHTML = `
        <div class="popup">
            <div class="popup_header">
                <div class="popup_header_name"></div>
                <div class="popup_header_cancel">

                </div>
            </div>
            <div class="popup_content"></div>
            <div class="popup_footer"></div>
        </div>
        `;
        this.el.setAttribute("tabindex", "-1");
        this.el.querySelector(".popup_header_cancel").addEventListener("click", this.cancel.bind(this));
        this.el.addEventListener("click", this.backgroundClicked.bind(this));
        this.el.focus();
        this.el.addEventListener("keydown", (e) => {
            if (e.code == "Escape") {
                if (this.cancelable) {
                    this.cancel();
                }
            }
        });
        this.setTitle(title);
    }
    backgroundClicked(e) {
        var div = e.target;
        if (!div.classList.contains("popup_background")) {
            return;
        }
        if (this.cancelable) {
            this.cancel();
        }
    }
    wide() {
        this.el.querySelector(".popup").classList.add("popup-wide");
    }
    focus() {
        this.el.focus();
    }
    cancel() {
        this.remove();
        if (this.cancelAction) {
            this.cancelAction();
        }
    }
    remove() {
        this.el.remove();
    }
    setContent(el) {
        this.el.querySelector(".popup_content").appendChild(el);
        this.el.querySelector(".popup_content").classList.add("popup_content-visible");
    }
    setCancelable() {
        this.cancelable = true;
        this.el.querySelector(".popup_header_cancel").classList.add("popup_header_cancel-visible");
    }
    setTitle(name) {
        this.el.querySelector(".popup_header_name").textContent = name;
    }
    addButton(name, handler, style) {
        this.el.querySelector(".popup_footer").classList.add("popup_footer-visible");
        var button = document.createElement("input");
        button.setAttribute("type", "button");
        button.setAttribute("class", "btn");
        switch (style) {
            case ButtonStyle.Accented:
                button.classList.add("btn-accented");
                break;
            case ButtonStyle.Delete:
                button.classList.add("btn-delete");
                break;
        }
        button.setAttribute("value", name);
        button.addEventListener("click", handler);
        this.el.querySelector(".popup_footer").appendChild(button);
        return button;
    }
    present() {
        document.body.appendChild(this.el);
        this.focus();
        console.log("heer");
        this.el.classList.add("popup_background-presented");
    }
    unpresent() {
        this.el.classList.remove("popup_background-presented");
    }
}
var ButtonStyle;
(function (ButtonStyle) {
    ButtonStyle[ButtonStyle["Default"] = 0] = "Default";
    ButtonStyle[ButtonStyle["Accented"] = 1] = "Accented";
    ButtonStyle[ButtonStyle["Delete"] = 2] = "Delete";
})(ButtonStyle || (ButtonStyle = {}));
class Alert extends Popup {
    constructor(title) {
        super(title);
        this.setCancelable();
        this.present();
        this.addButton("OK", this.remove.bind(this), ButtonStyle.Accented).focus();
    }
}
class Confirm extends Popup {
    constructor(title, handlerConfirm, handlerCancel, style) {
        super(title);
        this.setCancelable();
        if (!style) {
            style = ButtonStyle.Accented;
        }
        this.cancelAction = handlerCancel;
        this.addButton("Storno", () => {
            this.remove();
            if (handlerCancel) {
                handlerCancel();
            }
        });
        var primaryText = "OK";
        if (style = ButtonStyle.Delete) {
            primaryText = "Smazat";
        }
        this.primaryButton = this.addButton(primaryText, () => {
            this.remove();
            if (handlerConfirm) {
                handlerConfirm();
            }
        }, style);
        this.present();
        this.primaryButton.focus();
    }
}
class ContentPopup extends Popup {
    constructor(title, content) {
        super(title);
        this.setCancelable();
        this.setContent(content);
        this.wide();
    }
    show() {
        this.present();
    }
}
class LoadingPopup extends Popup {
    constructor() {
        super("");
        var contentEl = document.createElement("div");
        contentEl.innerHTML = "<progress class=\"progress\"></progress>";
        this.setContent(contentEl);
        this.present();
    }
    done() {
        this.remove();
    }
}
document.addEventListener("DOMContentLoaded", () => {
    bindStats();
    bindMarkdowns();
    bindTimestamps();
    bindRelations();
    bindImagePickers();
    bindLists();
    bindForm();
    bindImageViews();
    bindFlashMessages();
    bindScrolled();
    bindDatePicker();
    bindDropdowns();
    bindSearch();
    bindMainMenu();
    bindRelationList();
    bindTaskMonitor();
    bindNotifications();
    var content = document.createElement("div");
    content.innerHTML = "<h2>hello world</h2><br><textarea rows='10'></textarea>";
});
function bindFlashMessages() {
    var messages = document.querySelectorAll(".flash_message");
    for (var i = 0; i < messages.length; i++) {
        var message = messages[i];
        message.addEventListener("click", (e) => {
            var target = e.target;
            if (target.classList.contains("flash_message_close")) {
                var current = e.currentTarget;
                current.classList.add("hidden");
            }
        });
    }
}
function bindScrolled() {
    var lastScrollPosition = 0;
    var header = document.querySelector(".admin_header");
    document.addEventListener("scroll", (event) => {
        if (document.body.clientWidth < 1100) {
            return;
        }
        var scrollPosition = window.scrollY;
        if (scrollPosition > 0) {
            header.classList.add("admin_header-scrolled");
        }
        else {
            header.classList.remove("admin_header-scrolled");
        }
        lastScrollPosition = scrollPosition;
    });
}

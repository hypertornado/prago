var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
class Autoresize {
    constructor(el) {
        this.el = el;
        this.el.addEventListener("input", this.resizeIt.bind(this));
        this.resizeIt();
    }
    resizeIt() {
        var height = this.el.scrollHeight + 2;
        this.el.style.height = height + "px";
    }
}
function DOMinsertChildAtIndex(parent, child, index) {
    if (index >= parent.children.length) {
        parent.appendChild(child);
    }
    else {
        if (index < 0) {
            index = 0;
        }
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
    str = str.split('"').join("&quot;");
    str = str.split("'").join("&#39;");
    return str;
}
function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}
class ImageView {
    constructor(el) {
        this.el = el;
        var filesData = JSON.parse(el.getAttribute("data-images"));
        this.addFiles(filesData);
    }
    addFiles(filesData) {
        this.el.innerHTML = "";
        if (!filesData.Items) {
            return;
        }
        for (var i = 0; i < filesData.Items.length; i++) {
            let file = filesData.Items[i];
            this.addFile(file);
        }
    }
    addFile(file) {
        let container = document.createElement("button");
        container.setAttribute("type", "button");
        container.classList.add("imageview_image");
        container.setAttribute("href", file.ViewURL);
        container.setAttribute("title", file.ImageDescription);
        let imgEl = document.createElement("img");
        imgEl.classList.add("imageview_image_img");
        imgEl.setAttribute("src", file.ThumbURL);
        container.appendChild(imgEl);
        container.addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            let commands = [];
            commands.push({
                Name: "Zobrazit",
                URL: file.ViewURL,
            });
            commands.push({
                Name: "Kopírovat UUID",
                Handler: () => {
                    navigator.clipboard.writeText(file.UUID);
                    Prago.notificationCenter.flashNotification("Zkopírováno", null, true, false);
                },
                Icon: "glyphicons-basic-611-copy-duplicate.svg",
            });
            cmenu({
                Event: e,
                AlignByElement: true,
                Name: file.ImageName,
                Description: file.ImageDescription,
                Commands: commands,
                Rows: CMenu.rowsFromArray(file.Metadata),
            });
        });
        this.el.appendChild(container);
    }
}
class ImagePicker {
    constructor(el) {
        this.el = el;
        this.hiddenInput = (el.querySelector(".admin_images_hidden"));
        this.preview2 = el.querySelector(".imagepicker_preview");
        this.fileInput = (this.el.querySelector(".imagepicker_input"));
        this.progress = this.el.querySelector("progress");
        this.el.querySelector(".imagepicker_content").classList.remove("hidden");
        this.hideProgress();
        this.el.querySelector(".imagepicker_btn").addEventListener("click", (e) => {
            e.preventDefault();
            e.stopPropagation();
            cmenu({
                Event: e,
                AlignByElement: true,
                Commands: [
                    {
                        Name: "Nahrát nový soubor",
                        Icon: "glyphicons-basic-301-square-upload.svg",
                        Handler: () => {
                            this.fileInput.click();
                        },
                    },
                    {
                        Name: "Vložit UUID",
                        Icon: "glyphicons-basic-613-paste.svg",
                        Handler: () => {
                            new PopupForm("/admin/validate-uuid-files", (data) => {
                                this.addUUID(data.Data);
                                this.load();
                            });
                        },
                    },
                ],
            });
        });
        this.el.addEventListener("click", (e) => {
            if (e.altKey) {
                var ids = window.prompt("IDs of images", this.hiddenInput.value);
                this.hiddenInput.value = ids;
                this.load();
                e.preventDefault();
                return false;
            }
        });
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
            request.open("POST", "/admin/file/api/upload");
            request.addEventListener("load", (e) => {
                this.hideProgress();
                if (request.status == 200) {
                    var data = JSON.parse(request.response);
                    console.log(data);
                    for (var i = 0; i < data.length; i++) {
                        console.log(data[i]);
                        this.addUUID(data[i]);
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
        this.load();
    }
    hideProgress() {
        this.progress.classList.add("hidden");
    }
    showProgress() {
        this.progress.classList.remove("hidden");
    }
    load() {
        this.showProgress();
        var request = new XMLHttpRequest();
        request.open("GET", "/admin/api/imagepicker" + encodeParams({
            "ids": this.hiddenInput.value,
        }));
        request.addEventListener("load", (e) => {
            this.hideProgress();
            if (request.status == 200) {
                var data = JSON.parse(request.response);
                this.addFiles2(data);
            }
            else {
                new Alert("Chyba při načítání dat obrázků.");
                console.error("Error while loading item.");
            }
        });
        request.send();
    }
    addFiles2(data) {
        this.preview2.innerHTML = "";
        for (let i = 0; i < data.Items.length; i++) {
            let item = data.Items[i];
            let itemEl = document.createElement("button");
            itemEl.setAttribute("type", "button");
            itemEl.setAttribute("data-uuid", item.UUID);
            itemEl.setAttribute("title", item.ImageName);
            itemEl.classList.add("imagepicker_preview_item");
            itemEl.setAttribute("style", "background-image: url('" + item.ThumbURL + "');");
            itemEl.addEventListener("click", (e) => {
                e.stopPropagation();
                e.preventDefault();
                var commands = [];
                commands.push({
                    Name: "Zobrazit",
                    Icon: "glyphicons-basic-588-book-open-text.svg",
                    URL: item.ViewURL,
                });
                commands.push({
                    Name: "Upravit popis",
                    Icon: "glyphicons-basic-31-pencil.svg",
                    Handler: () => {
                        new PopupForm(item.EditURL, () => {
                            this.load();
                        });
                    }
                });
                commands.push({
                    Name: "První",
                    Icon: "glyphicons-basic-212-arrow-up.svg",
                    Handler: () => {
                        DOMinsertChildAtIndex(this.preview2, itemEl, 0);
                        this.updateHiddenData2();
                        this.load();
                    },
                });
                commands.push({
                    Name: "Nahoru",
                    Icon: "glyphicons-basic-828-arrow-thin-up.svg",
                    Handler: () => {
                        DOMinsertChildAtIndex(this.preview2, itemEl, i - 1);
                        this.updateHiddenData2();
                        this.load();
                    },
                });
                commands.push({
                    Name: "Dolů",
                    Icon: "glyphicons-basic-827-arrow-thin-down.svg",
                    Handler: () => {
                        DOMinsertChildAtIndex(this.preview2, itemEl, i + 2);
                        this.updateHiddenData2();
                        this.load();
                    },
                });
                commands.push({
                    Name: "Kopírovat UUID",
                    Handler: () => {
                        navigator.clipboard.writeText(item.UUID);
                        Prago.notificationCenter.flashNotification("Zkopírováno", null, true, false);
                    },
                    Icon: "glyphicons-basic-611-copy-duplicate.svg",
                });
                commands.push({
                    Name: "Smazat",
                    Handler: () => {
                        itemEl.remove();
                        this.updateHiddenData2();
                    },
                    Icon: "glyphicons-basic-17-bin.svg",
                    Style: "destroy",
                });
                var rows = [];
                for (var j = 0; j < item.Metadata.length; j++) {
                    rows.push({
                        Name: item.Metadata[j][0],
                        Value: item.Metadata[j][1],
                    });
                }
                cmenu({
                    Event: e,
                    AlignByElement: true,
                    Name: item.ImageName,
                    Description: item.ImageDescription,
                    Commands: commands,
                    Rows: CMenu.rowsFromArray(item.Metadata),
                });
            });
            this.preview2.appendChild(itemEl);
        }
    }
    updateHiddenData2() {
        var ids = [];
        for (var i = 0; i < this.preview2.children.length; i++) {
            let item = this.preview2.children[i];
            var uuid = item.getAttribute("data-uuid");
            ids.push(uuid);
        }
        this.hiddenInput.value = ids.join(",");
    }
    addUUID(uuid) {
        if (!uuid) {
            return;
        }
        let val = this.hiddenInput.value;
        if (val) {
            val += ",";
        }
        val += uuid;
        this.hiddenInput.value = val;
        this.load();
    }
}
class ListFilterRelations {
    constructor(el, value, list) {
        this.valueInput = el.querySelector(".filter_relations_hidden");
        this.input = el.querySelector(".filter_relations_search_input");
        this.search = el.querySelector(".filter_relations_search");
        this.suggestions = el.querySelector(".filter_relations_suggestions");
        this.preview = el.querySelector(".filter_relations_preview");
        this.previewImage = el.querySelector(".filter_relations_preview_image");
        this.previewName = el.querySelector(".filter_relations_preview_name");
        this.previewClose = el.querySelector(".filter_relations_preview_close");
        this.previewClose.addEventListener("click", this.closePreview.bind(this));
        this.preview.classList.add("hidden");
        this.preview.addEventListener("click", this.previewClicked.bind(this));
        let hiddenEl = el.querySelector("input");
        this.relatedResourceName = el
            .querySelector(".list_filter_item-relations")
            .getAttribute("data-related-resource");
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
        let apiURL = "/admin/" + this.relatedResourceName + "/api/preview-relation/" + value;
        request.open("GET", apiURL, true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                let respData = JSON.parse(request.response);
                if (respData.length > 0) {
                    this.renderPreview(respData[0]);
                }
            }
            else {
                console.error("not found");
            }
        });
        request.send();
    }
    renderPreview(item) {
        this.currentDataItem = item;
        this.valueInput.value = item.ID;
        this.preview.classList.remove("hidden");
        this.search.classList.add("hidden");
        this.preview.setAttribute("title", item.Name);
        if (item.Image) {
            this.previewImage.classList.remove("hidden");
            this.previewImage.setAttribute("style", "background-image: url('" + item.Image + "');");
        }
        else {
            this.previewImage.classList.add("hidden");
        }
        this.previewName.textContent = item.Name;
        this.dispatchChange();
    }
    previewClicked(e) {
        e.preventDefault();
        e.stopPropagation();
        cmenu({
            Event: e,
            AlignByElement: true,
            ImageURL: this.currentDataItem.Image,
            Name: this.currentDataItem.Name,
            Description: this.currentDataItem.Description,
            Commands: [
                {
                    Name: "Detail",
                    URL: this.currentDataItem.URL,
                }
            ]
        });
    }
    dispatchChange() {
        var event = new Event("change");
        this.valueInput.dispatchEvent(event);
    }
    closePreview(e) {
        e.preventDefault();
        e.stopPropagation();
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
        request.open("GET", "/admin/" +
            this.relatedResourceName +
            "/api/searchresource" +
            "?q=" +
            encodeURIComponent(q), true);
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
        for (var i = 0; i < data.Previews.length; i++) {
            this.suggestions.classList.remove("filter_relations_suggestions-empty");
            let item = data.Previews[i];
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
        var right = document.createElement("div");
        right.classList.add("list_filter_suggestion_right");
        var name = document.createElement("div");
        name.classList.add("list_filter_suggestion_name");
        name.textContent = data.Name;
        var description = document.createElement("div");
        description.classList.add("list_filter_suggestion_description");
        description.textContent = data.Description;
        var image = document.createElement("div");
        image.classList.add("list_filter_suggestion_image");
        if (data.Image) {
            image.setAttribute("style", "background-image: url('" + data.Image + "');");
        }
        ret.appendChild(image);
        right.appendChild(name);
        right.appendChild(description);
        ret.appendChild(right);
        return ret;
    }
}
class ListFilterDate {
    constructor(el, value) {
        this.hidden = el.querySelector(".list_filter_item");
        this.from = (el.querySelector(".list_filter_layout_date_from"));
        this.to = el.querySelector(".list_filter_layout_date_to");
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
        var event = new Event("change");
        this.hidden.dispatchEvent(event);
    }
}
class List {
    constructor(list) {
        this.minCellWidth = 50;
        this.normalCellWidth = 100;
        this.maxCellWidth = 500;
        this.list = list;
        this.rootContent = document.querySelector(".root_content");
        this.listHeaderContainer = this.rootContent.querySelector(".list_header_container");
        this.listTable = this.list.querySelector(".list_table");
        this.listHeader = this.list.querySelector(".list_header");
        this.listFooter = this.list.querySelector(".list_footer");
        this.settings = new ListSettings(this);
        let urlParams = new URLSearchParams(window.location.search);
        this.page = parseInt(urlParams.get("_page"));
        if (!this.page) {
            this.page = 1;
        }
        this.typeName = list.getAttribute("data-type");
        if (!this.typeName) {
            return;
        }
        this.progress = list.querySelector(".list_progress");
        this.tableContent = list.querySelector(".list_table_content");
        this.bindFilter(urlParams);
        this.defaultOrderColumn = list.getAttribute("data-order-column");
        if (list.getAttribute("data-order-desc") == "true") {
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
        this.defaultVisibleColumnsStr = list.getAttribute("data-visible-columns");
        var visibleColumnsStr = this.defaultVisibleColumnsStr;
        let visibleColumnsArr = visibleColumnsStr.split(",");
        let visibleColumnsMap = {};
        for (var i = 0; i < visibleColumnsArr.length; i++) {
            visibleColumnsMap[visibleColumnsArr[i]] = true;
        }
        this.itemsPerPage = parseInt(list.getAttribute("data-items-per-page"));
        this.paginationSelect = (document.querySelector(".list_settings_pages"));
        this.paginationSelect.addEventListener("change", this.load.bind(this));
        this.multiple = new ListMultiple(this);
        this.settings.bindOptions(visibleColumnsMap);
        this.bindOrder();
        this.bindInitialHeaderWidths();
        this.bindResizer();
        this.bindHeaderPositionCalculator();
    }
    copyColumnWidths() {
        let totalWidth = this.listHeader.getBoundingClientRect().width;
        let headerItems = this.list.querySelectorAll(".list_header > :not(.hidden)");
        let widths = [];
        for (let j = 0; j < headerItems.length; j++) {
            let headerEl = headerItems[j];
            var clientRect = headerEl.getBoundingClientRect();
            var elWidth = clientRect.width;
            widths.push(elWidth);
        }
        let tableRows = this.list.querySelectorAll(".list_row");
        for (let i = 0; i < tableRows.length; i++) {
            let rowItems = tableRows[i].children;
            for (let j = 0; j < widths.length; j++) {
                if (j >= rowItems.length) {
                    break;
                }
                let tableEl = rowItems[j];
                tableEl.style.width = widths[j] + "px";
            }
        }
        let placeholderItems = this.list.querySelectorAll(".list_tableplaceholder_row");
        if (placeholderItems.length > 0) {
            let placeholderWidth = totalWidth;
            for (let i = 0; i < placeholderItems.length; i++) {
                let item = placeholderItems[i];
                item.style.width = placeholderWidth + "px";
            }
        }
    }
    load() {
        if (this.currentRequest) {
            this.currentRequest.abort();
        }
        this.list.classList.add("list-loading");
        var request = new XMLHttpRequest();
        this.currentRequest = request;
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
        let filterData = this.getFilterData();
        for (var k in filterData) {
            params[k] = filterData[k];
        }
        this.colorActiveFilterItems();
        var encoded = encodeParams(params);
        window.history.replaceState(null, null, document.location.pathname + encoded);
        var columns = this.settings.getSelectedColumnsStr();
        if (columns != this.defaultVisibleColumnsStr) {
            params["_columns"] = columns;
        }
        let selectedPages = parseInt(this.paginationSelect.value);
        if (selectedPages != this.itemsPerPage) {
            params["_pagesize"] = selectedPages;
        }
        encoded = encodeParams(params);
        request.open("GET", "/admin/" + this.typeName + "/api/list" + encoded, true);
        request.addEventListener("load", () => {
            this.currentRequest = null;
            this.tableContent.innerHTML = "";
            if (request.status == 200) {
                var response = JSON.parse(request.response);
                this.tableContent.innerHTML = response.Content;
                this.listFooter.innerHTML = response.FooterStr;
                bindReOrder();
                this.bindSettingsButton();
                this.bindPagination();
                this.bindClick();
                this.bindFetchStats();
                if (this.multiple.hasMultipleActions()) {
                    this.multiple.bindMultipleActionCheckboxes();
                }
            }
            else {
                new Alert("Chyba při načítání položek.");
                console.error("error while loading list");
            }
            this.copyColumnWidths();
            this.list.classList.remove("list-loading");
            this.listHeaderContainer.classList.add("list_header_container-visible");
        });
        request.send(JSON.stringify({}));
    }
    colorActiveFilterItems() {
        let itemsToColor = this.getFilterData();
        var filterItems = this.list.querySelectorAll(".list_header_item_filter");
        for (var i = 0; i < filterItems.length; i++) {
            var item = filterItems[i];
            let name = item.getAttribute("data-name");
            if (itemsToColor[name]) {
                item.classList.add("list_filteritem-colored");
            }
            else {
                item.classList.remove("list_filteritem-colored");
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
    bindSettingsButton() {
        let btn = this.list.querySelector(".list_settings_btn2");
        this.settings.bindSettingsBtn(btn);
    }
    bindPagination() {
        var paginationEl = this.list.querySelector(".pagination");
        var totalPages = parseInt(paginationEl.getAttribute("data-total"));
        var selectedPage = parseInt(paginationEl.getAttribute("data-selected"));
        if (totalPages < 2) {
            return;
        }
        let beforeItemWasShown = true;
        for (var i = 1; i <= totalPages; i++) {
            let shouldBeShown = false;
            let maxDistance = 2;
            if (Math.abs(1 - i) <= maxDistance) {
                shouldBeShown = true;
            }
            if (Math.abs(totalPages - i) <= maxDistance) {
                shouldBeShown = true;
            }
            if (Math.abs(selectedPage - i) <= maxDistance) {
                shouldBeShown = true;
            }
            if (shouldBeShown) {
                if (!beforeItemWasShown) {
                    let delimiterEl = document.createElement("div");
                    delimiterEl.classList.add("pagination_page_delimiter");
                    delimiterEl.innerText = "…";
                    paginationEl.appendChild(delimiterEl);
                }
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
                beforeItemWasShown = true;
            }
            else {
                beforeItemWasShown = false;
            }
        }
    }
    bindFetchStats() {
        var cells = this.list.querySelectorAll(".list_cell[data-fetch-url]");
        for (var i = 0; i < cells.length; i++) {
            let cell = cells[i];
            let url = cell.getAttribute("data-fetch-url");
            if (!url) {
                continue;
            }
            if (cell.classList.contains("list_cell-fetched")) {
                continue;
            }
            if (!document.contains(cell)) {
                continue;
            }
            let cellContentSpan = (cell.querySelector(".list_cell_name"));
            fetch(url)
                .then((data) => {
                return data.json();
            })
                .then((data) => {
                cellContentSpan.innerText = data.Value;
                cell.setAttribute("title", data.Value);
                cell.classList.add("list_cell-fetched");
                this.bindFetchStats();
            })
                .catch((error) => {
                cellContentSpan.innerText = "⚠️";
                console.error("cant fetch data:", error);
            });
            return;
        }
    }
    bindClick() {
        var rows = this.list.querySelectorAll(".list_row");
        for (var i = 0; i < rows.length; i++) {
            let row = rows[i];
            row.addEventListener("contextmenu", this.contextClick.bind(this));
            row.addEventListener("click", (e) => {
                var el = e.currentTarget;
                var url = el.getAttribute("data-url");
                if (e.altKey) {
                    url += "/edit";
                    let targetEl = e.target;
                    targetEl = targetEl.closest(".list_cell");
                    let focusID = targetEl.getAttribute("data-cell-id");
                    if (focusID) {
                        url += "?_focus=" + focusID;
                    }
                }
                if (e.shiftKey || e.metaKey || e.ctrlKey) {
                    var openedWindow = window.open(url, "newwindow" + new Date() + Math.random());
                    openedWindow.focus();
                    return;
                }
                window.location.href = url;
            });
        }
    }
    createCmenu(e, rowEl, alignByElement) {
        rowEl.classList.add("list_row-context");
        let actions = JSON.parse(rowEl.getAttribute("data-actions"));
        var commands = [];
        let allowPopupForm = true;
        if (e.altKey || e.metaKey || e.shiftKey || e.ctrlKey) {
            allowPopupForm = false;
        }
        for (let action of actions.MenuButtons) {
            let actionURL = null;
            let handler = null;
            if (action.FormURL && allowPopupForm) {
                handler = () => {
                    new PopupForm(action.FormURL, (data) => {
                        this.load();
                    });
                };
            }
            else {
                actionURL = action.URL;
            }
            commands.push({
                Icon: action.Icon,
                Name: action.Name,
                URL: actionURL,
                Style: action.Style,
                Handler: handler,
            });
        }
        let name = rowEl.getAttribute("data-name");
        let preName = rowEl.getAttribute("data-prename");
        if (name == preName) {
            preName = null;
        }
        cmenu({
            Event: e,
            AlignByElement: alignByElement,
            ImageURL: rowEl.getAttribute("data-image-url"),
            PreName: preName,
            Name: name,
            Description: rowEl.getAttribute("data-description"),
            Commands: commands,
            DismissHandler: () => {
                rowEl.classList.remove("list_row-context");
            },
        });
        e.preventDefault();
    }
    contextClick(e) {
        let rowEl = e.currentTarget;
        this.createCmenu(e, rowEl, false);
    }
    bindOrder() {
        this.renderOrder();
        var headers = this.list.querySelectorAll(".list_header_item_name-canorder");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.addEventListener("click", (e) => {
                var el = e.currentTarget;
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
    bindResizer() {
        var resizers = this.list.querySelectorAll(".list_header_item_resizer");
        for (var i = 0; i < resizers.length; i++) {
            var resizer = resizers[i];
            let parentEl = resizer.parentElement;
            let naturalWidth = parseInt(parentEl.getAttribute("data-natural-width"));
            resizer.addEventListener("drag", (e) => {
                var clientRect = parentEl.getBoundingClientRect();
                var clientX = clientRect.left;
                if (e.clientX == 0) {
                    return false;
                }
                let width = e.clientX - clientX;
                this.setCellWidth(parentEl, width);
            });
            resizer.addEventListener("dblclick", (e) => {
                let width = this.getCellWidth(parentEl);
                if (width == this.minCellWidth) {
                    this.setCellWidth(parentEl, naturalWidth);
                }
                else {
                    if (width == naturalWidth) {
                        this.setCellWidth(parentEl, this.maxCellWidth);
                    }
                    else {
                        if (width == this.maxCellWidth) {
                            this.setCellWidth(parentEl, this.minCellWidth);
                        }
                        else {
                            this.setCellWidth(parentEl, naturalWidth);
                        }
                    }
                }
                this.copyColumnWidths();
            });
            resizer.addEventListener("dragend", (e) => {
                this.copyColumnWidths();
            });
        }
        this.copyColumnWidths();
    }
    getCellWidth(cell) {
        return cell.getBoundingClientRect().width;
    }
    setCellWidth(cell, width) {
        if (width < this.minCellWidth) {
            width = this.minCellWidth;
        }
        if (width > this.maxCellWidth) {
            width = this.maxCellWidth;
        }
        let cellName = cell.getAttribute("data-name");
        if (width + "" == cell.getAttribute("data-natural-width")) {
            this.webStorageDeleteWidth(cellName);
        }
        else {
            this.webStorageSetWidth(cellName, width);
        }
        cell.setAttribute("style", "width: " + width + "px;");
    }
    bindInitialHeaderWidths() {
        let headerItems = this.list.querySelectorAll(".list_header_item");
        for (var i = 0; i < headerItems.length; i++) {
            var itemEl = headerItems[i];
            let width = parseInt(itemEl.getAttribute("data-natural-width"));
            let cellName = itemEl.getAttribute("data-name");
            let savedWidth = this.webStorageLoadWidth(cellName);
            if (savedWidth > 0) {
                width = savedWidth;
            }
            this.setCellWidth(itemEl, width);
        }
    }
    webStorageWidthName(cell) {
        let tableName = this.typeName;
        return "prago_cellwidth_" + tableName + "_" + cell;
    }
    webStorageLoadWidth(cell) {
        let val = window.localStorage[this.webStorageWidthName(cell)];
        if (val) {
            return parseInt(val);
        }
        return 0;
    }
    webStorageSetWidth(cell, width) {
        window.localStorage[this.webStorageWidthName(cell)] = width;
    }
    webStorageDeleteWidth(cell) {
        window.localStorage.removeItem(this.webStorageWidthName(cell));
    }
    renderOrder() {
        var headers = this.list.querySelectorAll(".list_header_item_name-canorder");
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
    getFilterData() {
        var ret = {};
        var items = this.list.querySelectorAll(".list_filter_item");
        for (var i = 0; i < items.length; i++) {
            var item = items[i];
            var typ = item.getAttribute("data-typ");
            var layout = item.getAttribute("data-filter-layout");
            if (item.classList.contains("list_filter_item-relations")) {
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
        var filterFields = this.list.querySelectorAll(".list_header_item_filter");
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
        if (e.keyCode == 9 ||
            e.keyCode == 16 ||
            e.keyCode == 17 ||
            e.keyCode == 18) {
            return;
        }
        this.filterChanged();
    }
    filterChanged() {
        this.colorActiveFilterItems();
        this.page = 1;
        this.changed = true;
        this.changedTimestamp = Date.now();
        this.list.classList.add("list-loading");
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
    bindHeaderPositionCalculator() {
        this.listHeaderPositionChanged();
        window.addEventListener("resize", this.listHeaderPositionChanged.bind(this));
        this.list.addEventListener("scroll", this.listHeaderPositionChanged.bind(this));
        this.rootContent.addEventListener("scroll", this.listHeaderPositionChanged.bind(this));
        this.listTable.addEventListener("scroll", this.listHeaderPositionChanged.bind(this));
    }
    listHeaderPositionChanged() {
        let rect = this.rootContent.getBoundingClientRect();
        let scrolledClassName = "list_header_container-scrolled";
        if (this.rootContent.scrollTop > 50) {
            this.listHeaderContainer.classList.add(scrolledClassName);
        }
        else {
            this.listHeaderContainer.classList.remove(scrolledClassName);
        }
        return true;
    }
}
class ListSettings {
    constructor(list) {
        this.list = list;
        this.settingsEl = document.querySelector(".list_settings");
        this.settingsPopup = new ContentPopup("Nastavení", this.settingsEl);
        this.settingsPopup.setIcon("glyphicons-basic-137-cogwheel.svg");
        this.statsContainer = document.querySelector(".list_stats_container");
        this.statsEl = document.querySelector(".list_stats");
        this.statsPopup = new ContentPopup("Statistiky", this.statsEl);
        this.statsPopup.setIcon("glyphicons-basic-43-stats-circle.svg");
        this.statsCheckboxSelectCount = document.querySelector(".list_stats_limit");
        this.statsCheckboxSelectCount.addEventListener("change", () => {
            this.loadStats();
        });
    }
    bindSettingsBtn(btn) {
        btn.addEventListener("click", (e) => {
            e.stopPropagation();
            cmenu({
                Event: e,
                AlignByElement: true,
                Commands: [
                    {
                        Name: "Nastavení",
                        Icon: "glyphicons-basic-137-cogwheel.svg",
                        Handler: () => {
                            this.settingsPopup.show();
                        },
                    },
                    {
                        Name: "Statistiky",
                        Icon: "glyphicons-basic-43-stats-circle.svg",
                        Handler: () => {
                            this.loadStats();
                            this.statsPopup.show();
                        },
                    },
                    {
                        Name: "Export CSV",
                        Icon: "glyphicons-basic-302-square-download.svg",
                        Handler: () => {
                            window.open("/admin/" + this.list.typeName + "/api/export.csv");
                        }
                    },
                ],
            });
        });
    }
    loadStats() {
        let filterData = this.list.getFilterData();
        var params = {};
        params["_statslimit"] = this.statsCheckboxSelectCount.value;
        for (var k in filterData) {
            params[k] = filterData[k];
        }
        var request = new XMLHttpRequest();
        var encoded = encodeParams(params);
        request.open("GET", "/admin/" + this.list.typeName + "/api/list-stats" + encoded, true);
        this.statsContainer.innerHTML = "Loading...";
        request.addEventListener("load", () => {
            if (request.status == 200) {
                this.statsContainer.innerHTML = request.response;
            }
        });
        request.send();
    }
    bindOptions(visibleColumnsMap) {
        var columns = document.querySelectorAll(".list_settings_column");
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
        var headers = document.querySelectorAll(".list_header_item");
        for (var i = 0; i < headers.length; i++) {
            var name = headers[i].getAttribute("data-name");
            if (columns[name]) {
                headers[i].classList.remove("hidden");
            }
            else {
                headers[i].classList.add("hidden");
            }
        }
        var filters = document.querySelectorAll(".list_header_item_filter");
        for (var i = 0; i < filters.length; i++) {
            var name = filters[i].getAttribute("data-name");
            if (columns[name] === true) {
                filters[i].classList.remove("hidden");
            }
            if (columns[name] === false) {
            }
        }
        this.list.load();
    }
    getSelectedColumnsStr() {
        var ret = [];
        var checked = document.querySelectorAll(".list_settings_column:checked");
        for (var i = 0; i < checked.length; i++) {
            ret.push(checked[i].getAttribute("data-column-name"));
        }
        return ret.join(",");
    }
    getSelectedColumnsMap() {
        var columns = {};
        var inputs = document.querySelectorAll(".list_settings_column");
        for (var i = 0; i < inputs.length; i++) {
            if (inputs[i].checked) {
                columns[inputs[i].getAttribute("data-column-name")] = true;
            }
            else {
                columns[inputs[i].getAttribute("data-column-name")] = false;
            }
        }
        return columns;
    }
}
class ListMultiple {
    constructor(list) {
        this.list = list;
        if (this.hasMultipleActions()) {
            this.bindMultipleActions();
        }
    }
    hasMultipleActions() {
        if (this.list.list.classList.contains("list-hasmultipleactions")) {
            return true;
        }
        return false;
    }
    bindMultipleActions() {
        this.listHeaderAllSelect = this.list.list.querySelector(".list_header_multiple");
        this.listHeaderAllSelect.addEventListener("click", () => {
            if (this.isAllChecked()) {
                this.multipleUncheckAll();
            }
            else {
                this.multipleCheckAll();
            }
        });
        var actions = this.list.list.querySelectorAll(".list_multiple_action");
        for (var i = 0; i < actions.length; i++) {
            actions[i].addEventListener("click", this.multipleActionSelected.bind(this));
        }
        this.list.list
            .querySelector(".list_multiple_actions_cancel")
            .addEventListener("click", () => {
            this.multipleUncheckAll();
        });
    }
    multipleActionSelected(e) {
        var ids = this.multipleGetIDs();
        this.multipleActionStart(e.target, ids);
    }
    multipleActionStart(btn, ids) {
        let actionID = btn.getAttribute("data-id");
        let actionName = btn.getAttribute("data-name");
        switch (btn.getAttribute("data-action-type")) {
            case "mutiple_edit":
                new ListMultipleEdit(this, ids);
                break;
            case "mutiple_export":
                let urlStr = "/admin/" + this.list.typeName + "/api/export?ids=" + ids.join(",");
                window.open(urlStr);
                break;
            default:
                let confirm = new Confirm(`${actionName}: Opravdu chcete provést tuto akci na ${ids.length} položek?`, actionName, () => {
                    var loader = new LoadingPopup();
                    var params = {};
                    params["action"] = actionID;
                    params["ids"] = ids.join(",");
                    var url = "/admin/" +
                        this.list.typeName +
                        "/api/multipleaction" +
                        encodeParams(params);
                    fetch(url, {
                        method: "POST",
                    }).then((e) => {
                        loader.done();
                        if (e.status == 200) {
                            e.json().then((data) => {
                                if (data.ErrorStr) {
                                    new Alert(data.ErrorStr);
                                }
                                if (data.FlashMessage) {
                                    Prago.notificationCenter.flashNotification(actionName, data.FlashMessage, true, false);
                                }
                                if (data.RedirectURL) {
                                    window.location = data.RedirectURL;
                                }
                                this.list.load();
                            });
                        }
                        else {
                            Prago.notificationCenter.flashNotification(actionName, "Chyba " + e, false, true);
                            this.list.load();
                        }
                    });
                }, Function());
        }
    }
    bindMultipleActionCheckboxes() {
        this.lastCheckboxIndexClicked = -1;
        this.pseudoCheckboxesAr = document.querySelectorAll(".list_row_multiple");
        for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
            var checkbox = this.pseudoCheckboxesAr[i];
            checkbox.addEventListener("click", this.multipleCheckboxClicked.bind(this));
        }
        this.multipleCheckboxChanged();
    }
    multipleGetIDs() {
        var ret = [];
        for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
            var checkbox = this.pseudoCheckboxesAr[i];
            if (checkbox.classList.contains("list_row_multiple-checked")) {
                ret.push(checkbox.getAttribute("data-id"));
            }
        }
        return ret;
    }
    multipleCheckboxClicked(e) {
        var cell = e.currentTarget;
        var index = this.indexOfClickedCheckbox(cell);
        if (e.shiftKey && this.lastCheckboxIndexClicked >= 0) {
            var start = Math.min(index, this.lastCheckboxIndexClicked);
            var end = Math.max(index, this.lastCheckboxIndexClicked);
            for (var i = start; i <= end; i++) {
                this.checkPseudocheckbox(i);
            }
        }
        else {
            this.lastCheckboxIndexClicked = index;
            if (this.isCheckedPseudocheckbox(index)) {
                this.uncheckPseudocheckbox(index);
            }
            else {
                this.checkPseudocheckbox(index);
            }
        }
        e.preventDefault();
        e.stopPropagation();
        this.multipleCheckboxChanged();
        return false;
    }
    isCheckedPseudocheckbox(index) {
        var sb = this.pseudoCheckboxesAr[index];
        return sb.classList.contains("list_row_multiple-checked");
    }
    checkPseudocheckbox(index) {
        var sb = this.pseudoCheckboxesAr[index];
        sb.classList.add("list_row_multiple-checked");
    }
    uncheckPseudocheckbox(index) {
        var sb = this.pseudoCheckboxesAr[index];
        sb.classList.remove("list_row_multiple-checked");
    }
    multipleCheckboxChanged() {
        var checkedCount = 0;
        for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
            var checkbox = this.pseudoCheckboxesAr[i];
            if (checkbox.classList.contains("list_row_multiple-checked")) {
                checkedCount++;
            }
        }
        var multipleActionsPanel = this.list.list.querySelector(".list_multiple_actions");
        if (checkedCount > 0) {
            multipleActionsPanel.classList.add("list_multiple_actions-visible");
        }
        else {
            multipleActionsPanel.classList.remove("list_multiple_actions-visible");
        }
        if (this.isAllChecked()) {
            this.listHeaderAllSelect.classList.add("list_row_multiple-checked");
        }
        else {
            this.listHeaderAllSelect.classList.remove("list_row_multiple-checked");
        }
        this.list.list.querySelector(".list_multiple_actions_description").textContent = `Vybráno ${checkedCount} položek`;
    }
    multipleUncheckAll() {
        this.lastCheckboxIndexClicked = -1;
        for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
            var checkbox = this.pseudoCheckboxesAr[i];
            checkbox.classList.remove("list_row_multiple-checked");
        }
        this.multipleCheckboxChanged();
    }
    multipleCheckAll() {
        this.lastCheckboxIndexClicked = -1;
        for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
            var checkbox = this.pseudoCheckboxesAr[i];
            checkbox.classList.add("list_row_multiple-checked");
        }
        this.multipleCheckboxChanged();
    }
    isAllChecked() {
        if (this.pseudoCheckboxesAr.length == 0) {
            return false;
        }
        for (var i = 0; i < this.pseudoCheckboxesAr.length; i++) {
            var checkbox = this.pseudoCheckboxesAr[i];
            if (!checkbox.classList.contains("list_row_multiple-checked")) {
                return false;
            }
        }
        return true;
    }
    indexOfClickedCheckbox(el) {
        var ret = -1;
        this.pseudoCheckboxesAr.forEach((v, k) => {
            if (v == el) {
                ret = k;
            }
        });
        return ret;
    }
}
class ListMultipleEdit {
    constructor(multiple, ids) {
        this.listMultiple = multiple;
        var typeID = document.querySelector(".list").getAttribute("data-type");
        var progress = document.createElement("progress");
        this.popup = new ContentPopup(`Hromadná úprava položek (${ids.length} položek)`, progress);
        this.popup.show();
        fetch("/admin/" + typeID + "/api/multiple_edit?ids=" + ids.join(","))
            .then((response) => {
            if (response.ok) {
                return response.text();
            }
            else {
                this.popup.hide();
                new Alert("Operaci nelze nahrát.");
            }
        })
            .then((val) => {
            var div = document.createElement("div");
            div.innerHTML = val;
            this.popup.setContent(div);
            this.initFormPopup(div.querySelector("form"));
            this.popup.setConfirmButtons(this.confirm.bind(this));
        });
    }
    initFormPopup(form) {
        this.form = form;
        this.form.addEventListener("submit", this.confirm.bind(this));
        new Form(this.form);
        this.initCheckboxes();
    }
    initCheckboxes() {
        var checkboxes = this.form.querySelectorAll(".multiple_edit_field_checkbox");
        checkboxes.forEach((cb) => {
            cb.addEventListener("change", (e) => {
                var item = cb.parentElement.parentElement;
                if (cb.checked) {
                    item.classList.add("multiple_edit_field-selected");
                }
                else {
                    item.classList.remove("multiple_edit_field-selected");
                }
            });
        });
    }
    confirm(e) {
        var typeID = document.querySelector(".list").getAttribute("data-type");
        var data = new FormData(this.form);
        var loader = new LoadingPopup();
        fetch("/admin/" + typeID + "/api/multiple_edit", {
            method: "POST",
            body: data,
        }).then((response) => {
            loader.done();
            if (response.ok) {
                this.popup.hide();
                this.listMultiple.list.load();
            }
            else {
                if (response.status == 403) {
                    response.json().then((data) => {
                        new Alert(data.error.Text);
                    });
                    return;
                }
                else {
                    new Alert("Chyba při ukládání.");
                }
            }
        });
        e.preventDefault();
    }
}
function bindReOrder() {
    function orderTable(el) {
        var rows = el.getElementsByClassName("list_row");
        Array.prototype.forEach.call(rows, function (item, i) {
            bindDraggable(item);
        });
        var draggedElement;
        function bindDraggable(row) {
            row.setAttribute("draggable", "true");
            row.addEventListener("dragstart", function (ev) {
                row.classList.add("list_row-reorder");
                draggedElement = this;
                ev.dataTransfer.setData("text/plain", "");
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
                    Array.prototype.forEach.call(el.getElementsByClassName("list_row"), function (item, i) {
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
                row.classList.remove("list_row-reorder");
                return false;
            });
            row.addEventListener("dragover", function (ev) {
                ev.preventDefault();
            });
        }
        function saveOrder() {
            var typ = document.querySelector(".list-order").getAttribute("data-type");
            var ajaxPath = "/admin/" + typ + "/api/set-order";
            var order = [];
            var rows = el.getElementsByClassName("list_row");
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
            request.send(JSON.stringify({ order: order }));
        }
    }
    var elements = document.querySelectorAll(".list-order");
    Array.prototype.forEach.call(elements, function (el, i) {
        orderTable(el);
    });
}
class MarkdownEditor {
    constructor(el) {
        this.el = el;
        this.textarea = el.querySelector(".textarea");
        this.preview = el.querySelector(".admin_markdown_preview");
        new Autoresize(this.textarea);
        this.lastChanged = Date.now();
        this.changed = false;
        let showChange = (el.querySelector(".admin_markdown_preview_show"));
        showChange.addEventListener("change", () => {
            this.preview.classList.toggle("hidden");
        });
        setInterval(() => {
            if (this.changed && Date.now() - this.lastChanged > 500) {
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
        request.open("POST", "/admin/api/markdown", true);
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
class Timestamp {
    constructor(el) {
        this.elTsInput = el.getElementsByTagName("input")[0];
        this.elTsDate = (el.getElementsByClassName("admin_timestamp_date")[0]);
        this.elTsHour = (el.getElementsByClassName("admin_timestamp_hour")[0]);
        this.elTsMinute = (el.getElementsByClassName("admin_timestamp_minute")[0]);
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
        var minuteOption = (this.elTsMinute.children[minute]);
        minuteOption.selected = true;
        var hourOption = (this.elTsHour.children[hour]);
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
        var str = this.elTsDate.value +
            " " +
            this.elTsHour.value +
            ":" +
            this.elTsMinute.value;
        if (this.elTsDate.value == "") {
            str = "";
        }
        this.elTsInput.value = str;
    }
}
class RelationPicker {
    constructor(el) {
        this.selectedClass = "admin_item_relation_picker_suggestion-selected";
        if (el.getAttribute("data-autofocus") == "true") {
            this.autofocus = true;
        }
        if (el.getAttribute("data-multiple") == "true") {
            this.multipleInputs = true;
        }
        else {
            this.multipleInputs = false;
        }
        this.input = el.getElementsByTagName("input")[0];
        this.previewsContainer = (el.querySelector(".admin_relation_previews"));
        this.relationName = el.getAttribute("data-relation");
        this.progress = el.querySelector("progress");
        this.suggestionsEl = (el.querySelector(".admin_item_relation_picker_suggestions_content"));
        this.suggestions = [];
        this.picker = (el.querySelector(".admin_item_relation_picker"));
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
        if (parseInt(this.input.value) > 0) {
            this.getData();
        }
        else {
            this.progress.classList.add("hidden");
            this.showSearch();
        }
    }
    getData() {
        var request = new XMLHttpRequest();
        request.open("GET", "/admin/" +
            this.relationName +
            "/api/preview-relation/" +
            this.input.value, true);
        request.addEventListener("load", () => {
            this.progress.classList.add("hidden");
            if (request.status == 200) {
                let items = JSON.parse(request.response);
                for (var i = 0; i < items.length; i++) {
                    this.addPreview(items[i]);
                }
            }
            else {
                this.showSearch();
            }
        });
        request.send();
    }
    addPreview(data) {
        let previewEl = document.createElement("div");
        previewEl.classList.add("admin_relation_preview");
        var el = this.createPreview(data, true);
        this.previewsContainer.appendChild(previewEl);
        previewEl.appendChild(el);
        let upButton = document.createElement("div");
        upButton.classList.add("admin_relation_preview_action", "admin_relation_preview_action-up");
        upButton.innerText = "↑";
        previewEl.appendChild(upButton);
        upButton.addEventListener("click", (e) => {
            this.updateOrder(e, false);
        });
        let downButton = document.createElement("div");
        downButton.classList.add("admin_relation_preview_action", "admin_relation_preview_action-down");
        downButton.innerText = "↓";
        previewEl.appendChild(downButton);
        downButton.addEventListener("click", (e) => {
            this.updateOrder(e, true);
        });
        let deleteButton = document.createElement("div");
        deleteButton.classList.add("admin_relation_preview_action");
        deleteButton.innerText = "×";
        previewEl.appendChild(deleteButton);
        deleteButton.addEventListener("click", () => {
            previewEl.remove();
            this.updateLayout();
        });
        previewEl.setAttribute("data-id", data.ID);
        this.pickerInput.value = "";
        this.updateLayout();
    }
    numberOfItems() {
        return this.previewsContainer.children.length;
    }
    updateOrder(e, down) {
        let target = e.target;
        let previewEl = target.parentElement;
        let sibling;
        if (down) {
            sibling = previewEl.nextElementSibling;
        }
        else {
            sibling = previewEl.previousElementSibling;
        }
        if (!sibling) {
            return;
        }
        let parent = previewEl.parentElement;
        if (down) {
            parent.insertBefore(sibling, previewEl);
        }
        else {
            parent.insertBefore(previewEl, sibling);
        }
        this.updateLayout();
    }
    updateLayout() {
        if (this.multipleInputs || this.numberOfItems() == 0) {
            this.picker.classList.remove("hidden");
        }
        else {
            this.picker.classList.add("hidden");
        }
        this.updateInput();
    }
    updateInput() {
        var valItems = [];
        for (var i = 0; i < this.previewsContainer.children.length; i++) {
            let child = this.previewsContainer.children[i];
            let val = child.getAttribute("data-id");
            valItems.push(val);
        }
        let val = valItems.join(";");
        if (this.multipleInputs) {
            val = ";" + val + ";";
        }
        this.input.value = val;
    }
    showSearch() {
        this.picker.classList.remove("hidden");
        this.suggestions = [];
        this.suggestionsEl.innerText = "";
        this.pickerInput.value = "";
        if (this.autofocus) {
            this.pickerInput.focus();
        }
    }
    getSuggestions(q) {
        var request = new XMLHttpRequest();
        request.open("GET", "/admin/" +
            this.relationName +
            "/api/searchresource" +
            "?q=" +
            encodeURIComponent(q), true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                if (q != this.pickerInput.value) {
                    return;
                }
                var data = JSON.parse(request.response);
                this.suggestions = data.Previews;
                this.suggestionsEl.innerText = "";
                if (data.Message) {
                    let messageEl = document.createElement("div");
                    messageEl.innerText = data.Message;
                    messageEl.classList.add("relation_message");
                    this.suggestionsEl.appendChild(messageEl);
                }
                for (var i = 0; i < data.Previews.length; i++) {
                    var item = data.Previews[i];
                    var el = this.createPreview(item, false);
                    el.classList.add("admin_item_relation_picker_suggestion");
                    el.setAttribute("data-position", i + "");
                    el.addEventListener("mousedown", this.suggestionClick.bind(this));
                    el.addEventListener("mouseenter", this.suggestionSelect.bind(this));
                    this.suggestionsEl.appendChild(el);
                }
                if (data.Button) {
                    let buttonEl = document.createElement("a");
                    let buttonElIcon = document.createElement("img");
                    buttonElIcon.setAttribute("src", "/admin/api/icons?file=glyphicons-basic-371-plus.svg");
                    buttonElIcon.classList.add("btn_icon");
                    let buttonElText = document.createElement("span");
                    buttonElText.innerText = data.Button.Name;
                    buttonEl.appendChild(buttonElIcon);
                    buttonEl.appendChild(buttonElText);
                    buttonEl.classList.add("btn", "relation_button");
                    buttonEl.addEventListener("click", (e) => {
                        this.suggestionsEl.classList.add("hidden");
                        let popupForm = new PopupForm(data.Button.FormURL, (data) => {
                            this.addPreview(data.Data);
                        });
                        e.preventDefault();
                        e.stopPropagation();
                    });
                    buttonEl.addEventListener("mousedown", (e) => {
                        e.preventDefault();
                        e.stopPropagation();
                    });
                    this.suggestionsEl.appendChild(buttonEl);
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
            this.addPreview(this.suggestions[selected]);
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
        this.suggestionsEl
            .querySelectorAll(".admin_preview")[i].classList.add(this.selectedClass);
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
        ret.addEventListener("mouseleave", () => {
            this.unselect();
        });
        var right = document.createElement("div");
        right.classList.add("admin_preview_right");
        var name = document.createElement("div");
        name.classList.add("admin_preview_name");
        name.textContent = data.Name;
        var description = document.createElement("description");
        description.classList.add("admin_preview_description");
        description.setAttribute("title", data.Description);
        description.textContent = data.Description;
        if (data.Image) {
            let image = document.createElement("img");
            image.classList.add("admin_preview_image");
            image.setAttribute("src", data.Image);
            image.setAttribute("loading", "lazy");
            ret.appendChild(image);
        }
        else {
            let imageDiv = document.createElement("div");
            imageDiv.classList.add("admin_preview_image");
            ret.appendChild(imageDiv);
        }
        right.appendChild(name);
        right.appendChild(description);
        ret.appendChild(right);
        return ret;
    }
}
class Form {
    constructor(form) {
        this.dirty = false;
        this.dirty = false;
        this.formEl = form;
        this.fixAufofocus();
        var elements = form.querySelectorAll(".admin_markdown");
        elements.forEach((el) => {
            new MarkdownEditor(el);
        });
        var timestamps = form.querySelectorAll(".admin_timestamp");
        timestamps.forEach((form) => {
            new Timestamp(form);
        });
        var relations = form.querySelectorAll(".admin_item_relation");
        relations.forEach((form) => {
            new RelationPicker(form);
        });
        var imagePickers = form.querySelectorAll(".imagepicker");
        imagePickers.forEach((form) => {
            new ImagePicker(form);
        });
        form.addEventListener("submit", () => {
            this.dirty = false;
        });
        let els = form.querySelectorAll(".form_watcher");
        for (var i = 0; i < els.length; i++) {
            var input = els[i];
            input.addEventListener("keyup", this.messageChanged.bind(this));
            input.addEventListener("change", this.changed.bind(this));
        }
        window.setInterval(() => {
            if (this.dirty && Date.now() - this.lastChanged > 500) {
                this.changed();
            }
        }, 100);
    }
    messageChanged() {
        if (this.willChangeHandler) {
            this.willChangeHandler();
        }
        this.dirty = true;
        this.lastChanged = Date.now();
    }
    changed() {
        if (this.changeHandler) {
            this.dirty = false;
            this.changeHandler();
        }
        else {
            this.dirty = true;
        }
    }
    fixAufofocus() {
        let input = this.formEl.querySelector('[autofocus]');
        if (input) {
            let value = input.value;
            let typ = input.getAttribute("type");
            if (input.nodeName == "TEXTAREA" || typ == "text" || typ == "password" || typ == "tel" || typ == "search" || typ == "url") {
                input.focus();
                input.setSelectionRange(value.length, value.length);
            }
        }
    }
}
class FormContainer {
    constructor(formContainer, okHandler) {
        this.formContainer = formContainer;
        this.okHandler = okHandler;
        this.progress = formContainer.querySelector(".form_progress");
        var formEl = formContainer.querySelector("form");
        this.form = new Form(formEl);
        this.form.formEl.addEventListener("submit", this.submitFormAJAX.bind(this));
        if (this.isAutosubmitFirstTime()) {
            this.sendForm();
        }
        if (this.isAutosubmit()) {
            this.form.changeHandler = this.formChanged.bind(this);
            this.form.willChangeHandler = this.formWillChange.bind(this);
            this.sendForm();
        }
    }
    isAutosubmitFirstTime() {
        if (this.formContainer.classList.contains("form_container-autosubmitfirsttime")) {
            return true;
        }
        else {
            return false;
        }
    }
    isAutosubmit() {
        if (this.formContainer.classList.contains("form_container-autosubmit")) {
            return true;
        }
        else {
            return false;
        }
    }
    formWillChange() {
        this.progress.classList.remove("hidden");
    }
    formChanged() {
        this.sendForm();
    }
    submitFormAJAX(event) {
        event.preventDefault();
        this.sendForm();
    }
    sendForm() {
        let formData = new FormData(this.form.formEl);
        let request = new XMLHttpRequest();
        request.open("POST", this.form.formEl.getAttribute("action"));
        let requestID = makeid(10);
        this.lastAJAXID = requestID;
        if (this.activeRequest) {
            if (this.isAutosubmit()) {
                this.activeRequest.abort();
            }
            else {
                return;
            }
        }
        this.activeRequest = request;
        this.form.formEl.classList.remove("form-errors");
        request.addEventListener("load", (e) => {
            if (requestID != this.lastAJAXID) {
                return;
            }
            this.activeRequest = null;
            if (request.status == 200) {
                let contentType = request.getResponseHeader("Content-Type");
                if (contentType == "application/json") {
                    var data = JSON.parse(request.response);
                    if (data.RedirectionLocation || data.Preview || data.Data) {
                        this.okHandler(data);
                    }
                    else {
                        this.progress.classList.add("hidden");
                        this.setFormErrors(data.Errors);
                        if (data.AfterContent)
                            this.setAfterContent(data.AfterContent);
                    }
                }
                else {
                    var blob = new Blob([request.response], {
                        type: "application/octet-stream",
                    });
                    var downloadUrl = URL.createObjectURL(blob);
                    var a = document.createElement("a");
                    a.href = downloadUrl;
                    a.download = "data.xlsx";
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    URL.revokeObjectURL(downloadUrl);
                    this.progress.classList.add("hidden");
                }
            }
            else {
                this.progress.classList.add("hidden");
                new Alert("Chyba při nahrávání souboru.");
            }
        });
        this.progress.classList.remove("hidden");
        request.send(formData);
    }
    setAfterContent(text) {
        this.formContainer.querySelector(".form_after_content").innerHTML = text;
    }
    setFormErrors(errors) {
        this.deleteItemErrors();
        let errorsDiv = this.form.formEl.querySelector(".form_errors");
        errorsDiv.innerText = "";
        errorsDiv.classList.add("hidden");
        if (errors) {
            for (let i = 0; i < errors.length; i++) {
                if (errors[i].Field) {
                    this.setItemError(errors[i]);
                }
                else {
                    let errorDiv = document.createElement("div");
                    errorDiv.classList.add("form_errors_error");
                    errorDiv.innerText = errors[i].Text;
                    errorsDiv.appendChild(errorDiv);
                }
            }
            if (errors.length > 0) {
                this.form.formEl.classList.add("form-errors");
                errorsDiv.classList.remove("hidden");
            }
        }
    }
    deleteItemErrors() {
        let labels = this.form.formEl.querySelectorAll(".form_label");
        for (let i = 0; i < labels.length; i++) {
            let label = labels[i];
            label.classList.remove("form_label-errors");
            let labelErrors = label.querySelector(".form_label_errors");
            labelErrors.innerHTML = "";
            labelErrors.classList.add("hidden");
        }
    }
    setItemError(itemError) {
        let labels = this.form.formEl.querySelectorAll(".form_label");
        for (let i = 0; i < labels.length; i++) {
            let label = labels[i];
            let id = label.getAttribute("data-id");
            if (label.getAttribute("data-id") == itemError.Field) {
                label.classList.add("form_label-errors");
                let labelErrors = label.querySelector(".form_label_errors");
                labelErrors.classList.remove("hidden");
                let errorDiv = document.createElement("div");
                errorDiv.classList.add("form_label_errors_error");
                errorDiv.innerText = itemError.Text;
                labelErrors.appendChild(errorDiv);
            }
        }
    }
}
function makeid(length) {
    var result = "";
    var characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    var charactersLength = characters.length;
    for (var i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}
class SearchForm {
    constructor(el) {
        this.searchForm = el;
        this.searchInput = el.querySelector(".searchbox_input");
        this.suggestionsEl = (el.querySelector(".searchbox_suggestions"));
        Prago.shortcuts.add({
            Key: "f",
            Alt: true,
        }, "Vyhledávání", () => {
            this.searchInput.focus();
        });
        this.searchInput.addEventListener("input", () => {
            this.suggestions = [];
            this.dirty = true;
            this.deleteSuggestions();
            this.lastChanged = Date.now();
            return false;
        });
        window.setInterval(() => {
            if (this.dirty && Date.now() - this.lastChanged > 100) {
                this.loadSuggestions();
            }
        }, 30);
        this.searchInput.addEventListener("keydown", (e) => {
            if (e.keyCode == 27) {
                this.searchInput.blur();
                e.preventDefault();
                return false;
            }
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
    deleteSuggestions() {
        this.suggestionsEl.innerHTML = "";
        this.searchForm.classList.remove("searchbox-showsuggestions");
    }
    loadSuggestions() {
        this.dirty = false;
        var suggestText = this.searchInput.value;
        var request = new XMLHttpRequest();
        var url = "/admin/api/search-suggest" + encodeParams({ q: this.searchInput.value });
        request.open("GET", url);
        request.addEventListener("load", () => {
            if (suggestText != this.searchInput.value) {
                return;
            }
            if (request.status == 200) {
                this.addSuggestions(request.response);
            }
            else {
                this.deleteSuggestions();
                console.error("Error while loading item.");
            }
        });
        request.send();
    }
    addSuggestions(content) {
        this.suggestionsEl.innerHTML = content;
        this.suggestions = this.suggestionsEl.querySelectorAll(".admin_search_suggestion");
        if (this.suggestions.length > 0) {
            this.searchForm.classList.add("searchbox-showsuggestions");
        }
        else {
            this.searchForm.classList.remove("searchbox-showsuggestions");
        }
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
class Menu {
    constructor() {
        this.rootEl = document.querySelector(".root");
        this.rootLeft = document.querySelector(".root_left");
        this.hamburgerMenuEl = document.querySelector(".root_hamburger");
        this.hamburgerMenuEl.addEventListener("click", this.menuClick.bind(this));
        var searchFormEl = document.querySelector(".searchbox");
        if (searchFormEl) {
            this.search = new SearchForm(searchFormEl);
        }
        this.scrollTo(this.loadFromStorage());
        this.rootLeft.addEventListener("scroll", this.scrollHandler.bind(this));
        this.bindSubmenus();
        this.bindResourceCounts();
    }
    scrollHandler() {
        this.saveToStorage(this.rootLeft.scrollTop);
    }
    saveToStorage(position) {
        window.localStorage["left_menu_position"] = position;
    }
    menuClick() {
        this.rootEl.classList.toggle("root-visible");
    }
    loadFromStorage() {
        var pos = window.localStorage["left_menu_position"];
        if (pos) {
            return parseInt(pos);
        }
        return 0;
    }
    scrollTo(position) {
        this.rootLeft.scrollTo(0, position);
    }
    bindSubmenus() {
        let triangleIcons = document.querySelectorAll(".menu2_item_icon");
        for (var i = 0; i < triangleIcons.length; i++) {
            let triangleIcon = triangleIcons[i];
            triangleIcon.addEventListener("click", () => {
                let parent = triangleIcon.parentElement.parentElement;
                parent.classList.toggle("menu2_item-expanded");
            });
        }
    }
    bindResourceCounts() {
        this.setResourceCountsFromCache();
        new VisibilityReloader(2000, () => {
            this.loadResourceCounts();
        });
    }
    saveCountToStorage(url, count) {
        if (!window.localStorage) {
            return;
        }
        window.localStorage["left_menu_count-" + url] = count;
    }
    loadCountFromStorage(url) {
        if (!window.localStorage) {
            return "";
        }
        var pos = window.localStorage["left_menu_count-" + url];
        if (pos) {
            return pos;
        }
        return "";
    }
    setResourceCountsFromCache() {
        var items = document.querySelectorAll(".menu2_item_content");
        for (var i = 0; i < items.length; i++) {
            let item = items[i];
            let url = item.getAttribute("href");
            let count = this.loadCountFromStorage(url);
            if (count) {
                this.setResourceCount(item, count);
            }
        }
    }
    setResourceCounts(data) {
        var items = document.querySelectorAll(".menu2_item_content");
        for (var i = 0; i < items.length; i++) {
            let item = items[i];
            let url = item.getAttribute("href");
            let count = data[url];
            this.setResourceCount(item, count);
        }
    }
    setResourceCount(el, count) {
        let countEl = el.querySelector(".menu2_item_content_subname");
        if (count) {
            this.saveCountToStorage(el.getAttribute("href"), count);
            countEl.textContent = count;
        }
    }
    loadResourceCounts() {
        var request = new XMLHttpRequest();
        request.open("GET", "/admin/api/resource-counts", true);
        request.addEventListener("load", () => {
            if (request.status == 200) {
                var data = JSON.parse(request.response);
                this.setResourceCounts(data);
            }
            else {
                console.error("cant load resource counts");
            }
        });
        request.send();
    }
}
class RelationList {
    constructor(el) {
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
        request.open("POST", "/admin/api/relationlist", true);
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
        var count = 5;
        if (this.offset > 0) {
            count = 10;
        }
        request.send(JSON.stringify({
            SourceResource: this.sourceResource,
            TargetResource: this.targetResource,
            TargetField: this.targetField,
            IDValue: this.idValue,
            Offset: this.offset,
            Count: count,
        }));
    }
}
class NotificationCenter {
    constructor(el) {
        this.notifications = new Map();
        this.el = el;
        var data = el.getAttribute("data-notification-views");
        var notifications = [];
        if (data) {
            notifications = JSON.parse(data);
        }
        notifications.forEach((item) => {
            this.setData(item);
        });
        this.periodDataLoader();
    }
    periodDataLoader() {
        return __awaiter(this, void 0, void 0, function* () {
            for (;;) {
                if (!document.hidden)
                    this.loadData();
                yield sleep(1000);
            }
        });
    }
    loadData() {
        fetch("/admin/api/notifications")
            .then((response) => response.json())
            .then((data) => data.forEach((d) => {
            this.setData(d);
        }));
    }
    setData(data) {
        var notification;
        if (this.notifications.has(data.UUID)) {
            notification = this.notifications.get(data.UUID);
        }
        else {
            notification = new NotificationItem();
            this.notifications.set(data.UUID, notification);
            this.el.appendChild(notification.el);
        }
        notification.setData(data);
    }
    bindNotification(el) {
        el.querySelector(".notification_close").addEventListener("click", () => {
            el.classList.add("notification-closed");
        });
    }
    flashNotification(name, description, success, fail) {
        var style = "";
        if (success) {
            style = "success";
        }
        if (fail) {
            style = "fail";
        }
        this.setData({
            UUID: makeid(10),
            Name: name,
            Description: description,
            IsFlash: true,
            Style: style,
        });
    }
}
class NotificationItem {
    constructor() {
        this.el = document.createElement("div");
        this.el.classList.add("notification");
        this.el.innerHTML = `
      <div class="notification_close"></div>
      <div class="notification_left">
        <div class="notification_left_progress">
          <div class="notification_left_progress_human"></div>
          <progress class="notification_left_progressbar"></progress>
        </div>
      </div>
      <div class="notification_right">
          <div class="notification_prename"></div>
          <div class="notification_name"></div>
          <div class="notification_description"></div>
          <div class="notification_action" data-id="primary"></div>
          <div class="notification_action" data-id="secondary"></div>
      </div>
    `;
        this.actionElements = this.el.querySelectorAll(".notification_action");
        this.actionElements.forEach((el) => {
            el.addEventListener("click", (e) => {
                var target = e.currentTarget;
                this.sendAction(target.getAttribute("data-id"));
                return false;
            });
        });
        this.el.querySelector(".notification_left");
        this.el
            .querySelector(".notification_close")
            .addEventListener("click", (e) => {
            this.sendAction("delete");
            this.el.classList.add("notification-closed");
            e.stopPropagation();
            return false;
        });
        this.el.addEventListener("click", () => {
            var url = this.el.getAttribute("data-url");
            if (!url) {
                return;
            }
            window.location.href = url;
        });
    }
    sendAction(actionID) {
        fetch("/admin/api/notifications" +
            encodeParams({
                uuid: this.uuid,
                action: actionID,
            }), {
            method: "POST",
        }).then((e) => {
            if (!e.ok) {
                alert("error while deleting notification");
            }
        });
    }
    setAction(actionEl, action) {
        if (!action) {
            actionEl.classList.remove("notification_action-visible");
            return;
        }
        actionEl.classList.add("notification_action-visible");
        actionEl.textContent = action;
    }
    setData(data) {
        this.uuid = data.UUID;
        this.el.querySelector(".notification_prename").textContent = data.PreName;
        this.el.querySelector(".notification_name").textContent = data.Name;
        this.el.querySelector(".notification_description").textContent =
            data.Description;
        this.el
            .querySelector(".notification_description")
            .setAttribute("title", data.Description);
        var left = this.el.querySelector(".notification_left");
        left.classList.remove("notification_left-visible");
        if (data.Image) {
            left.classList.add("notification_left-visible");
            left.setAttribute("style", `background-image: url('${data.Image}');`);
        }
        var closeButton = this.el.querySelector(".notification_close");
        if (data.DisableCancel) {
            closeButton.classList.add("notification_close-disabled");
        }
        else {
            closeButton.classList.remove("notification_close-disabled");
        }
        this.setAction(this.actionElements[0], data.PrimaryAction);
        this.setAction(this.actionElements[1], data.SecondaryAction);
        this.el.classList.remove("notification-success");
        this.el.classList.remove("notification-fail");
        if (data.Style) {
            this.el.classList.add("notification-" + data.Style);
        }
        var progressEl = this.el.querySelector(".notification_left_progress");
        if (data.Progress) {
            left.classList.add("notification_left-visible");
            progressEl.classList.add("notification_left_progress-visible");
            this.el.querySelector(".notification_left_progress_human").textContent =
                data.Progress.Human;
            var progressBar = this.el.querySelector(".notification_left_progressbar");
            if (data.Progress.Percentage < 0) {
                delete progressBar.value;
            }
            else {
                progressBar.setAttribute("value", data.Progress.Percentage + "");
            }
        }
        else {
            progressEl.classList.remove("notification_left_progress-visible");
        }
        if (data.URL) {
            this.el.classList.add("notification-clickable");
            this.el.setAttribute("data-url", data.URL);
        }
        else {
            this.el.classList.remove("notification-clickable");
            this.el.setAttribute("data-url", "");
        }
        if (data.IsFlash) {
            window.setTimeout(() => {
                this.close();
            }, 5000);
        }
    }
    close() {
        this.el.classList.add("notification-closed");
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
                <img class="popup_header_icon hidden">
                <div class="popup_header_name"></div>
                <div class="popup_header_cancel"></div>
            </div>
            <div class="popup_content"></div>
            <div class="popup_footer"></div>
        </div>
        `;
        this.el.setAttribute("tabindex", "-1");
        this.el
            .querySelector(".popup_header_cancel")
            .addEventListener("click", this.cancel.bind(this));
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
        if (this.cancelAction) {
            this.cancelAction();
        }
        else {
            this.remove();
        }
    }
    remove() {
        this.el.remove();
    }
    setContent(el) {
        this.el.querySelector(".popup_content").innerHTML = "";
        this.el.querySelector(".popup_content").appendChild(el);
        this.el
            .querySelector(".popup_content")
            .classList.add("popup_content-visible");
    }
    setCancelable() {
        this.cancelable = true;
        this.el
            .querySelector(".popup_header_cancel")
            .classList.add("popup_header_cancel-visible");
    }
    setTitle(name) {
        this.el.querySelector(".popup_header_name").textContent = name;
    }
    setIcon(iconName) {
        if (!iconName) {
            return;
        }
        let iconEl = this.el.querySelector(".popup_header_icon");
        iconEl.classList.remove("hidden");
        iconEl.setAttribute("src", `/admin/api/icons?file=${iconName}&color=444444`);
    }
    addButton(name, handler, style) {
        this.el
            .querySelector(".popup_footer")
            .classList.add("popup_footer-visible");
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
        this.setIcon("glyphicons-basic-79-triangle-empty-alert.svg");
        this.addButton("OK", this.remove.bind(this), ButtonStyle.Accented).focus();
    }
}
class Confirm extends Popup {
    constructor(title, buttonName, handlerConfirm, handlerCancel, style) {
        super(title);
        this.setCancelable();
        if (!style) {
            style = ButtonStyle.Accented;
        }
        this.cancelAction = () => {
            this.remove();
            if (handlerCancel) {
                handlerCancel();
            }
        };
        this.addButton("Storno", () => {
            this.cancelAction();
        });
        var primaryText = buttonName;
        if (!primaryText) {
            primaryText = "OK";
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
        this.isShown = false;
        this.setCancelable();
        if (content)
            this.setContent(content);
        this.wide();
        this.cancelAction = this.hide.bind(this);
    }
    show() {
        this.present();
        this.focus();
        this.isShown = true;
    }
    hide() {
        this.unpresent();
        this.isShown = false;
    }
    setHiddenHandler(handler) {
        this.cancelAction = () => {
            handler();
            this.remove();
        };
    }
    setContent(content) {
        super.setContent(content);
    }
    setConfirmButtons(handler) {
        super.addButton("Storno", () => {
            super.unpresent();
        });
        super.addButton("Upravit", handler, ButtonStyle.Accented);
    }
}
class LoadingPopup extends Popup {
    constructor() {
        super("");
        var contentEl = document.createElement("div");
        contentEl.innerHTML = '<progress class="progress"></progress>';
        this.setContent(contentEl);
        this.present();
    }
    done() {
        this.remove();
    }
}
class PopupForm extends Popup {
    constructor(path, dataHandler) {
        super("⌛️");
        this.dataHandler = dataHandler;
        this.setCancelable();
        this.present();
        this.setIcon("glyphicons-basic-30-clipboard.svg");
        this.loadForm(path);
    }
    loadForm(path) {
        fetch(path)
            .then((response) => {
            if (response.ok) {
                return response.text();
            }
            else {
                this.unpresent();
                new Alert("Formulář nelze nahrát.");
            }
        })
            .then((textVal) => {
            this.wide();
            const parser = new DOMParser();
            const document = parser.parseFromString(textVal, "text/html");
            let formContainerEl = document.querySelector(".form_container");
            this.setContent(formContainerEl);
            new FormContainer(formContainerEl, this.okHandler.bind(this));
            this.setTitle(formContainerEl.getAttribute("data-form-name"));
            this.setIcon(formContainerEl.getAttribute("data-form-icon"));
        });
    }
    okHandler(data) {
        this.unpresent();
        this.dataHandler(data);
    }
}
function initGoogleMaps() {
    return __awaiter(this, void 0, void 0, function* () {
        var viewElements = document.querySelectorAll(".admin_item_view_place");
        var pickerElements = document.querySelectorAll(".map_picker");
        if (viewElements.length == 0 && pickerElements.length == 0) {
            return;
        }
        const { Map } = yield google.maps.importLibrary("maps");
        const { Autocomplete } = yield google.maps.importLibrary('places');
        const { AdvancedMarkerElement, PinElement } = yield google.maps.importLibrary("marker");
        viewElements.forEach((el) => {
            initGoogleMapView(el);
        });
        pickerElements.forEach((el) => {
            new GoogleMapEdit(el);
        });
    });
}
function initGoogleMapView(el) {
    var val = el.getAttribute("data-value");
    el.innerText = "";
    var coords = val.split(",");
    if (coords.length != 2) {
        el.classList.remove("admin_item_view_place");
        return;
    }
    const location = { lat: parseFloat(coords[0]), lng: parseFloat(coords[1]) };
    const map = new google.maps.Map(el, {
        zoom: 14,
        center: location,
        mapId: "3dab4a498a2dadb",
    });
    const marker = new google.maps.marker.AdvancedMarkerElement({
        position: location,
        map: map,
    });
}
class GoogleMapEdit {
    constructor(el) {
        this.el = el;
        this.statusEl = el.querySelector(".map_picker_description");
        var mapEl = el.querySelector(".map_picker_map");
        this.input = this.el.querySelector(".map_picker_value");
        this.deleteButton = el.querySelector(".map_picker_delete");
        this.searchContainer = el.querySelector(".map_picker_search");
        const location = { lng: 14.41854, lat: 50.073658 };
        let pac2 = new google.maps.places.PlaceAutocompleteElement();
        this.searchContainer.append(pac2);
        pac2.addEventListener('gmp-select', (_a) => __awaiter(this, [_a], void 0, function* ({ placePrediction }) {
            const place = placePrediction.toPlace();
            yield place.fetchFields({ fields: ['displayName', 'formattedAddress', 'location'] });
            this.setValue(place.location.lat(), place.location.lng());
            this.centreMap(place.location.lat(), place.location.lng());
        }));
        this.map = new google.maps.Map(mapEl, {
            zoom: 1,
            center: location,
            mapId: "3dab4a498a2dadb",
        });
        this.marker = new google.maps.marker.AdvancedMarkerElement({
            position: location,
            map: null,
            gmpDraggable: true,
        });
        this.marker.addListener("gmp-click", (e) => {
            this.deleteValue();
        });
        this.marker.addListener("drag", (e) => {
            this.setValue(e.latLng.lat(), e.latLng.lng());
        });
        this.map.addListener("click", (e) => {
            this.setValue(e.latLng.lat(), e.latLng.lng());
        });
        this.deleteButton.addEventListener("click", () => {
            this.deleteValue();
        });
        var inVals = this.input.value.split(",");
        if (inVals.length == 2) {
            let lat = parseFloat(inVals[0]);
            let lng = parseFloat(inVals[1]);
            this.setValue(lat, lng);
            this.centreMap(lat, lng);
        }
        else {
            this.deleteValue();
        }
    }
    centreMap(lat, lng) {
        let location = {
            lat: lat,
            lng: lng,
        };
        this.map.setCenter(location);
        this.map.setZoom(14);
    }
    setValue(lat, lng) {
        let location = {
            lat: lat,
            lng: lng,
        };
        this.marker.position = location;
        this.marker.map = this.map;
        this.input.value = lat + "," + lng;
        this.statusEl.textContent = "Latitude: " + lat + ", Longitude: " + lng;
        this.deleteButton.classList.remove("hidden");
    }
    deleteValue() {
        this.marker.map = null;
        this.input.value = "";
        this.statusEl.textContent = "Polohu vyberete kliknutím na mapu";
        this.deleteButton.classList.add("hidden");
    }
}
class QuickActions {
    constructor(el) {
        var buttons = el.querySelectorAll(".quick_actions_btn");
        for (var i = 0; i < buttons.length; i++) {
            let button = buttons[i];
            button.addEventListener("click", this.buttonClicked.bind(this));
        }
    }
    buttonClicked(e) {
        var btn = e.target;
        let actionURL = btn.getAttribute("data-url");
        new Confirm("Potvrdit akci", "", () => {
            let lp = new LoadingPopup();
            fetch(actionURL, {
                method: "POST",
            })
                .then((response) => {
                lp.done();
                if (response.ok) {
                    return response.text();
                }
                else {
                    throw response.text();
                }
            })
                .then((val) => {
                location.reload();
            })
                .catch((val) => {
                return val;
            })
                .then((val) => {
                if (val) {
                    new Alert(val);
                }
            });
        });
    }
}
function initDashboard() {
    var dashboardTables = document.querySelectorAll(".dashboard_table");
    dashboardTables.forEach((el) => {
        new DashboardTable(el);
    });
    var dashboardFigures = document.querySelectorAll(".dashboard_figure");
    dashboardFigures.forEach((el) => {
        new DashboardFigure(el);
    });
    var dashboardTimelines = document.querySelectorAll(".timeline");
    dashboardTimelines.forEach((el) => {
        new Timeline(el);
    });
}
class DashboardTable {
    constructor(el) {
        this.el = el;
        let reloadSeconds = parseInt(this.el.getAttribute("data-refresh-time-seconds"));
        new VisibilityReloader(reloadSeconds * 1000, this.loadTableData.bind(this));
    }
    loadTableData() {
        var request = new XMLHttpRequest();
        var params = {
            uuid: this.el.getAttribute("data-uuid"),
        };
        request.addEventListener("load", () => {
            if (request.status == 200) {
                this.el.innerHTML = request.response;
            }
            else {
                this.el.innerText = "Error while loading table";
            }
        });
        request.open("GET", "/admin/api/dashboard-table" + encodeParams(params), true);
        request.send();
    }
}
class DashboardFigure {
    constructor(el) {
        this.el = el;
        this.valueEl = el.querySelector(".dashboard_figure_value");
        this.descriptionEl = el.querySelector(".dashboard_figure_description");
        let reloadSeconds = parseInt(this.el.getAttribute("data-refresh-time-seconds"));
        new VisibilityReloader(reloadSeconds * 1000, this.loadFigureData.bind(this));
    }
    loadFigureData() {
        var request = new XMLHttpRequest();
        var params = {
            uuid: this.el.getAttribute("data-uuid"),
        };
        request.addEventListener("load", () => {
            this.el.classList.remove("dashboard_figure-loading");
            if (request.status == 200) {
                let data = JSON.parse(request.response);
                this.el.classList.remove("dashboard_figure-green", "dashboard_figure-red");
                this.valueEl.innerText = data["Value"];
                this.valueEl.setAttribute("title", data["Value"]);
                this.descriptionEl.innerText = data["Description"];
                this.descriptionEl.setAttribute("title", data["Description"]);
                if (data["IsRed"]) {
                    this.el.classList.add("dashboard_figure-red");
                }
                if (data["IsGreen"]) {
                    this.el.classList.add("dashboard_figure-green");
                }
            }
            else {
                this.valueEl.innerText = "Error while loading item.";
            }
        });
        request.open("GET", "/admin/api/dashboard-figure" + encodeParams(params), true);
        request.send();
    }
}
class Shortcuts {
    constructor(el) {
        this.el = el;
        this.shortcuts = [];
        this.el.addEventListener("keydown", (e) => {
            for (let shortcut of this.shortcuts) {
                if (shortcut.match(e)) {
                    shortcut.handler();
                    e.preventDefault();
                    e.stopPropagation();
                    return false;
                }
            }
        });
    }
    add(shortcut, description, handler) {
        this.shortcuts.push(new Shortcut(shortcut, description, handler));
    }
    addRootShortcuts() {
        let popup = new ContentPopup("Zkratky");
        this.add({
            Key: "?",
        }, "Zobrazit nápovědu", () => {
            if (document.activeElement !== document.body) {
                return;
            }
            let contentEl = document.createElement("div");
            for (let shortcut of this.shortcuts) {
                let shortcutEl = document.createElement("div");
                shortcutEl.innerText = shortcut.getDescription();
                contentEl.appendChild(shortcutEl);
            }
            if (popup.isShown) {
                popup.hide();
            }
            else {
                popup.setContent(contentEl);
                popup.show();
            }
        });
    }
}
class Shortcut {
    constructor(shortcut, description, handler) {
        this.shortcut = shortcut;
        this.handler = handler;
        this.description = description;
    }
    match(e) {
        if (e.key != this.shortcut.Key) {
            return false;
        }
        if (this.shortcut.Alt && !e.altKey) {
            return false;
        }
        if (this.shortcut.Shift && !e.shiftKey) {
            return false;
        }
        if (this.shortcut.Control && !e.ctrlKey && !e.metaKey) {
            return false;
        }
        return true;
    }
    getDescription() {
        let items = [];
        if (this.shortcut.Control) {
            items.push("Ctrl");
        }
        if (this.shortcut.Alt) {
            items.push("Alt");
        }
        if (this.shortcut.Shift) {
            items.push("Shift");
        }
        items.push(this.shortcut.Key);
        return items.join("+") + ": " + this.description;
    }
}
function cmenu(data) {
    Prago.cmenu.showWithData(data);
}
class CMenu {
    constructor() {
        for (let eventType of ["click", "visibilitychange", "blur"]) {
            document.addEventListener(eventType, (e) => {
                this.dismiss();
            });
        }
        document.addEventListener("keydown", (e) => {
            if (e.key == "Escape") {
                this.dismiss();
            }
        });
    }
    static rowsFromArray(inArr) {
        var rows = [];
        for (var j = 0; j < inArr.length; j++) {
            rows.push({
                Name: inArr[j][0],
                Value: inArr[j][1],
            });
        }
        return rows;
    }
    dismiss() {
        if (this.lastEl) {
            this.lastEl.remove();
            this.lastEl = null;
        }
        if (this.dismissHandler) {
            this.dismissHandler();
            this.dismissHandler = null;
        }
    }
    showWithData(data) {
        this.dismiss();
        let y = data.Event.clientY;
        let x = data.Event.clientX;
        let containerEl = document.createElement("div");
        containerEl.classList.add("cmenu_container");
        containerEl.addEventListener("contextmenu", (e) => {
            e.preventDefault();
        });
        let el = document.createElement("div");
        el.classList.add("cmenu");
        containerEl.appendChild(el);
        if (data.ImageURL) {
            let imageEl = document.createElement("img");
            imageEl.classList.add("cmenu_image");
            imageEl.setAttribute("src", data.ImageURL);
            el.appendChild(imageEl);
        }
        if (data.PreName) {
            let preNameEl = document.createElement("div");
            preNameEl.classList.add("cmenu_prename");
            preNameEl.innerText = data.PreName;
            preNameEl.setAttribute("title", data.PreName);
            el.appendChild(preNameEl);
        }
        if (data.Name) {
            let nameEl = document.createElement("div");
            nameEl.classList.add("cmenu_name");
            nameEl.innerText = data.Name;
            nameEl.setAttribute("title", data.Name);
            el.appendChild(nameEl);
        }
        if (data.Description) {
            let descEl = document.createElement("div");
            descEl.classList.add("cmenu_description");
            descEl.innerText = data.Description;
            descEl.setAttribute("title", data.Description);
            el.appendChild(descEl);
        }
        if (data.Rows) {
            let rowsEl = document.createElement("div");
            rowsEl.classList.add("cmenu_table_rows");
            for (let i = 0; i < data.Rows.length; i++) {
                let row = data.Rows[i];
                let rowEl = document.createElement("div");
                rowEl.classList.add("cmenu_table_row");
                let rowNameEl = document.createElement("div");
                rowNameEl.classList.add("cmenu_table_row_name");
                rowNameEl.innerText = row.Name;
                rowEl.appendChild(rowNameEl);
                let rowValueEl = document.createElement("div");
                rowValueEl.classList.add("cmenu_table_row_value");
                rowValueEl.innerText = row.Value;
                rowEl.appendChild(rowValueEl);
                rowsEl.appendChild(rowEl);
            }
            el.appendChild(rowsEl);
        }
        if (data.Commands) {
            let commandsEl = document.createElement("div");
            commandsEl.classList.add("cmenu_commands");
            for (let command of data.Commands) {
                let commandEl = document.createElement("div");
                commandEl.classList.add("cmenu_command");
                if (command.Style) {
                    commandEl.classList.add("cmenu_command-" + command.Style);
                }
                let commandNameEl = document.createElement("div");
                commandNameEl.classList.add("cmenu_command_name");
                commandNameEl.innerText = command.Name;
                commandEl.appendChild(commandNameEl);
                if (command.Icon) {
                    let commandNameIcon = document.createElement("img");
                    commandNameIcon.classList.add("cmenu_command_icon");
                    let color = "4077bf";
                    if (command.Style == "destroy") {
                        color = "cb2431";
                    }
                    commandNameIcon.setAttribute("src", "/admin/api/icons?file=" + command.Icon + "&color=" + color);
                    commandEl.appendChild(commandNameIcon);
                }
                commandEl.addEventListener("click", (e) => {
                    if (command.URL) {
                        if (e.shiftKey || e.metaKey || e.ctrlKey) {
                            var openedWindow = window.open(command.URL, "newwindow" + new Date() + Math.random());
                            openedWindow.focus();
                        }
                        else {
                            window.location.href = command.URL;
                        }
                    }
                    if (command.Handler) {
                        command.Handler();
                    }
                    this.dismiss();
                });
                commandsEl.appendChild(commandEl);
            }
            el.appendChild(commandsEl);
        }
        document.body.appendChild(containerEl);
        let elWidth = el.clientWidth;
        let elHeight = el.clientHeight;
        let viewportWidth = window.innerWidth;
        let viewportHeight = window.innerHeight;
        if (data.AlignByElement) {
            let targetEl = data.Event.currentTarget;
            let rect = targetEl.getBoundingClientRect();
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
        }
        else {
            if (x + elWidth > viewportWidth) {
                x = viewportWidth - elWidth;
            }
            if (y + elHeight > viewportHeight) {
                y = viewportHeight - elHeight;
            }
        }
        el.style.left = x + "px";
        el.style.top = y + "px";
        el.addEventListener("click", (e) => {
            e.stopPropagation();
        });
        this.lastEl = containerEl;
        this.dismissHandler = data.DismissHandler;
    }
    getOffset(el) {
        var _x = 0;
        var _y = 0;
        while (el && !isNaN(el.offsetLeft) && !isNaN(el.offsetTop)) {
            _x += el.offsetLeft - el.scrollLeft;
            _y += el.offsetTop - el.scrollTop;
            el = el.offsetParent;
        }
        return { top: _y, left: _x };
    }
}
class Timeline {
    constructor(el) {
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
        });
        this.el.querySelector(".timeline_toolbar_next").addEventListener("click", () => {
            this.changePosition(true);
        });
        this.changedType();
    }
    loadData() {
        this.setLoader();
        var dateStr;
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
        if (this.lastRequest) {
            this.lastRequest.abort();
        }
        this.lastRequest = request;
        var params = {
            _uuid: this.el.getAttribute("data-uuid"),
            _date: dateStr,
            _width: this.el.clientWidth,
        };
        request.addEventListener("load", () => {
            if (request.status == 200) {
                let data = JSON.parse(request.response);
                this.setData(data);
            }
            else {
                this.valuesEl.innerText = "Error while loading timeline";
                console.error("Error while loading timeline");
            }
            this.lastRequest = null;
        });
        request.open("GET", "/admin/api/timeline" + encodeParams(params), true);
        request.send();
    }
    setData(data) {
        this.valuesEl.innerText = "";
        for (var i = 0; i < data.Values.length; i++) {
            let val = data.Values[i];
            this.setValue(val);
        }
    }
    setLoader() {
        this.valuesEl.innerHTML = '<progress class="progress"></progress>';
    }
    setValue(data) {
        let valEl = document.createElement("div");
        valEl.innerHTML = `
            <div class="timeline_value_bars"></div>
            <div class="timeline_value_name" title="${data.Name}">
                <span class="timeline_value_name_inner">${data.Name}</span>
            </div>
        `;
        valEl.classList.add("timeline_value");
        let barsEl = valEl.querySelector(".timeline_value_bars");
        for (var i = 0; i < data.Bars.length; i++) {
            this.addBar(barsEl, data.Bars[i]);
        }
        if (data.IsCurrent) {
            valEl.classList.add("timeline_value-current");
        }
        this.valuesEl.appendChild(valEl);
        valEl.addEventListener("click", (e) => {
            var tableRows = [];
            for (let i = 0; i < data.Bars.length; i++) {
                let bar = data.Bars[i];
                tableRows.push({
                    Name: bar.KeyName,
                    Value: bar.ValueText,
                });
            }
            e.stopPropagation();
            cmenu({
                Event: e,
                Name: data.Name,
                Rows: tableRows,
            });
        });
    }
    addBar(el, barValue) {
        let barEl = document.createElement("div");
        barEl.innerHTML = `
            <div class="timeline_value_bar_inner" style="${barValue.StyleCSS}"></div>
        `;
        barEl.setAttribute("title", barValue.ValueText);
        barEl.classList.add("timeline_value_bar");
        el.appendChild(barEl);
        let labelEl = document.createElement("div");
        labelEl.classList.add("timeline_value_label");
        labelEl.innerText = barValue.ValueText;
        labelEl.setAttribute("style", barValue.LabelStyleCSS);
        barEl.appendChild(labelEl);
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
    changePosition(next) {
        let typ = this.typeSelect.value;
        if (typ == "day") {
            let date = new Date(this.datepicker.value);
            if (isNaN(date.getTime())) {
                console.error("Invalid date format. Please use YYYY-MM-DD.");
                return;
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
                month += 1;
                if (month > 12) {
                    month = 1;
                    year += 1;
                }
            }
            else {
                month -= 1;
                if (month < 1) {
                    month = 12;
                    year -= 1;
                }
            }
            let format = year + "-";
            if (month < 10) {
                format += "0";
            }
            format += month + "";
            this.monthpicker.value = format;
        }
        if (typ == "year") {
            let year = parseInt(this.yearpicker.value);
            if (next) {
                year += 1;
            }
            else {
                year -= 1;
            }
            this.yearpicker.value = year + "";
        }
        this.loadData();
    }
}
class Prago {
    static start() {
        document.addEventListener("DOMContentLoaded", Prago.init);
    }
    static init() {
        Prago.shortcuts = new Shortcuts(document.body);
        Prago.shortcuts.addRootShortcuts();
        Prago.cmenu = new CMenu();
        var listEl = document.querySelector(".list");
        if (listEl) {
            new List(listEl);
        }
        var formContainerElements = document.querySelectorAll(".form_container");
        formContainerElements.forEach((el) => {
            new FormContainer(el, (data) => {
                if (data.RedirectionLocation) {
                    window.location = data.RedirectionLocation;
                }
            });
        });
        var imageViews = document.querySelectorAll(".imageview");
        imageViews.forEach((el) => {
            new ImageView(el);
        });
        var menuEl = document.querySelector(".root_left");
        if (menuEl) {
            new Menu();
        }
        var relationListEls = document.querySelectorAll(".admin_relationlist");
        relationListEls.forEach((el) => {
            new RelationList(el);
        });
        Prago.notificationCenter = new NotificationCenter(document.querySelector(".notification_center"));
        var qa = document.querySelector(".quick_actions");
        if (qa) {
            new QuickActions(qa);
        }
        initDashboard();
        initGoogleMaps();
        let searchboxButton = document.querySelector(".searchbox_button");
        if (searchboxButton) {
            searchboxButton.addEventListener("click", (e) => {
                let input = document.querySelector(".searchbox_input");
                if (!input.value) {
                    input.focus();
                    e.stopPropagation();
                    e.preventDefault();
                }
            });
        }
    }
    static testPopupForm() {
        new PopupForm("/admin/packageview/new", (data) => {
            console.log("form data");
            console.log(data);
        });
    }
}
Prago.start();
class VisibilityReloader {
    constructor(reloadIntervalMilliseconds, handler) {
        this.lastRequestedTime = 0;
        window.setInterval(() => {
            if (document.visibilityState == "visible" &&
                Date.now() - this.lastRequestedTime >= reloadIntervalMilliseconds) {
                this.lastRequestedTime = Date.now();
                handler();
            }
        }, 100);
    }
}

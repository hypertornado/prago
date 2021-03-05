function bindStats() {
}
var Autoresize = (function () {
    function Autoresize(el) {
        this.el = el;
        this.el.addEventListener('input', this.resizeIt.bind(this));
        this.resizeIt();
    }
    Autoresize.prototype.resizeIt = function () {
        var height = this.el.scrollHeight + 2;
        this.el.style.height = height + 'px';
    };
    return Autoresize;
}());
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
var ImageView = (function () {
    function ImageView(el) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.el = el;
        var ids = el.getAttribute("data-images").split(",");
        this.addImages(ids);
    }
    ImageView.prototype.addImages = function (ids) {
        this.el.innerHTML = "";
        for (var i = 0; i < ids.length; i++) {
            if (ids[i] != "") {
                this.addImage(ids[i]);
            }
        }
    };
    ImageView.prototype.addImage = function (id) {
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("href", this.adminPrefix + "/file/uuid/" + id);
        container.setAttribute("style", "background-image: url('" + this.adminPrefix + "/_api/image/thumb/" + id + "');");
        var img = document.createElement("div");
        img.setAttribute("src", this.adminPrefix + "/_api/image/thumb/" + id);
        img.setAttribute("draggable", "false");
        var descriptionEl = document.createElement("div");
        descriptionEl.classList.add("admin_images_image_description");
        container.appendChild(descriptionEl);
        var request = new XMLHttpRequest();
        request.open("GET", this.adminPrefix + "/_api/imagedata/" + id);
        request.addEventListener("load", function (e) {
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
    };
    return ImageView;
}());
function bindImagePickers() {
    var els = document.querySelectorAll(".admin_images");
    for (var i = 0; i < els.length; i++) {
        new ImagePicker(els[i]);
    }
}
var ImagePicker = (function () {
    function ImagePicker(el) {
        var _this = this;
        this.el = el;
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.hiddenInput = el.querySelector(".admin_images_hidden");
        this.preview = el.querySelector(".admin_images_preview");
        this.fileInput = this.el.querySelector(".admin_images_fileinput input");
        this.progress = this.el.querySelector("progress");
        this.el.querySelector(".admin_images_loaded").classList.remove("hidden");
        this.hideProgress();
        var ids = this.hiddenInput.value.split(",");
        this.el.addEventListener("click", function (e) {
            if (e.altKey) {
                var ids = window.prompt("IDs of images", _this.hiddenInput.value);
                _this.hiddenInput.value = ids;
                e.preventDefault();
                return false;
            }
        });
        this.fileInput.addEventListener("dragenter", function (ev) {
            _this.fileInput.classList.add("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragleave", function (ev) {
            _this.fileInput.classList.remove("admin_images_fileinput-droparea");
        });
        this.fileInput.addEventListener("dragover", function (ev) {
            ev.preventDefault();
        });
        this.fileInput.addEventListener("drop", function (ev) {
            var text = ev.dataTransfer.getData('Text');
            return;
        });
        for (var i = 0; i < ids.length; i++) {
            var id = ids[i];
            if (id) {
                this.addImage(id);
            }
        }
        this.fileInput.addEventListener("change", function (e) {
            var files = _this.fileInput.files;
            var formData = new FormData();
            if (files.length == 0) {
                return;
            }
            for (var i = 0; i < files.length; i++) {
                formData.append("file", files[i]);
            }
            var request = new XMLHttpRequest();
            request.open("POST", _this.adminPrefix + "/_api/image/upload");
            request.addEventListener("load", function (e) {
                _this.hideProgress();
                if (request.status == 200) {
                    var data = JSON.parse(request.response);
                    for (var i = 0; i < data.length; i++) {
                        _this.addImage(data[i].UID);
                    }
                }
                else {
                    alert("Chyba při nahrávání souboru.");
                    console.error("Error while loading item.");
                }
            });
            _this.showProgress();
            request.send(formData);
        });
    }
    ImagePicker.prototype.updateHiddenData = function () {
        var ids = [];
        for (var i = 0; i < this.preview.children.length; i++) {
            var item = this.preview.children[i];
            var uuid = item.getAttribute("data-uuid");
            ids.push(uuid);
        }
        this.hiddenInput.value = ids.join(",");
    };
    ImagePicker.prototype.addImage = function (id) {
        var _this = this;
        var container = document.createElement("a");
        container.classList.add("admin_images_image");
        container.setAttribute("data-uuid", id);
        container.setAttribute("draggable", "true");
        container.setAttribute("target", "_blank");
        container.setAttribute("href", this.adminPrefix + "/file/uuid/" + id);
        container.setAttribute("style", "background-image: url('" + this.adminPrefix + "/_api/image/thumb/" + id + "');");
        var descriptionEl = document.createElement("div");
        descriptionEl.classList.add("admin_images_image_description");
        container.appendChild(descriptionEl);
        var request = new XMLHttpRequest();
        request.open("GET", this.adminPrefix + "/_api/imagedata/" + id);
        request.addEventListener("load", function (e) {
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
        container.addEventListener("dragstart", function (e) {
            _this.draggedElement = e.target;
        });
        container.addEventListener("drop", function (e) {
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
            var parent = _this.draggedElement.parentElement;
            for (var i = 0; i < parent.children.length; i++) {
                var child = parent.children[i];
                if (child == _this.draggedElement) {
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
            DOMinsertChildAtIndex(parent, _this.draggedElement, droppedIndex);
            _this.updateHiddenData();
            e.preventDefault();
            return false;
        });
        container.addEventListener("dragover", function (e) {
            e.preventDefault();
        });
        container.addEventListener("click", function (e) {
            var target = e.target;
            if (target.classList.contains("admin_images_image_delete")) {
                var parent = e.currentTarget.parentNode;
                parent.removeChild(e.currentTarget);
                _this.updateHiddenData();
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
    };
    ImagePicker.prototype.hideProgress = function () {
        this.progress.classList.add("hidden");
    };
    ImagePicker.prototype.showProgress = function () {
        this.progress.classList.remove("hidden");
    };
    return ImagePicker;
}());
var ListFilterRelations = (function () {
    function ListFilterRelations(el, value, list) {
        var _this = this;
        this.valueInput = el.querySelector(".filter_relations_hidden");
        this.input = el.querySelector(".filter_relations_search_input");
        this.search = el.querySelector(".filter_relations_search");
        this.suggestions = el.querySelector(".filter_relations_suggestions");
        this.preview = el.querySelector(".filter_relations_preview");
        this.previewName = el.querySelector(".filter_relations_preview_name");
        this.previewClose = el.querySelector(".filter_relations_preview_close");
        this.previewClose.addEventListener("click", this.closePreview.bind(this));
        this.preview.classList.add("hidden");
        var hiddenEl = el.querySelector("input");
        this.relatedResourceName = el.querySelector(".admin_table_filter_item-relations").getAttribute("data-related-resource");
        this.input.addEventListener("input", function () {
            _this.dirty = true;
            _this.lastChanged = Date.now();
            return false;
        });
        window.setInterval(function () {
            if (_this.dirty && Date.now() - _this.lastChanged > 100) {
                _this.loadSuggestions();
            }
        }, 30);
        if (this.valueInput.value) {
            this.loadPreview(this.valueInput.value);
        }
    }
    ListFilterRelations.prototype.loadPreview = function (value) {
        var _this = this;
        var request = new XMLHttpRequest();
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        request.open("GET", adminPrefix + "/_api/preview/" + this.relatedResourceName + "/" + value, true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                _this.renderPreview(JSON.parse(request.response));
            }
            else {
                console.error("not found");
            }
        });
        request.send();
    };
    ListFilterRelations.prototype.renderPreview = function (item) {
        this.valueInput.value = item.ID;
        this.preview.classList.remove("hidden");
        this.search.classList.add("hidden");
        this.previewName.textContent = item.Name;
        this.dispatchChange();
    };
    ListFilterRelations.prototype.dispatchChange = function () {
        var event = new Event('change');
        this.valueInput.dispatchEvent(event);
    };
    ListFilterRelations.prototype.closePreview = function () {
        this.valueInput.value = "";
        this.preview.classList.add("hidden");
        this.search.classList.remove("hidden");
        this.input.value = "";
        this.suggestions.innerHTML = "";
        this.suggestions.classList.add("filter_relations_suggestions-empty");
        this.dispatchChange();
        this.input.focus();
    };
    ListFilterRelations.prototype.loadSuggestions = function () {
        this.getSuggestions(this.input.value);
        this.dirty = false;
    };
    ListFilterRelations.prototype.getSuggestions = function (q) {
        var _this = this;
        var request = new XMLHttpRequest();
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        request.open("GET", adminPrefix + "/_api/search/" + this.relatedResourceName + "?q=" + encodeURIComponent(q), true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                _this.renderSuggestions(JSON.parse(request.response));
            }
            else {
                console.error("not found");
            }
        });
        request.send();
    };
    ListFilterRelations.prototype.renderSuggestions = function (data) {
        var _this = this;
        this.suggestions.innerHTML = "";
        this.suggestions.classList.add("filter_relations_suggestions-empty");
        var _loop_1 = function () {
            this_1.suggestions.classList.remove("filter_relations_suggestions-empty");
            var item = data[i];
            var el = this_1.renderSuggestion(item);
            this_1.suggestions.appendChild(el);
            var index = i;
            el.addEventListener("mousedown", function (e) {
                _this.renderPreview(item);
            });
        };
        var this_1 = this;
        for (var i = 0; i < data.length; i++) {
            _loop_1();
        }
    };
    ListFilterRelations.prototype.renderSuggestion = function (data) {
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
    };
    return ListFilterRelations;
}());
var ListFilterDate = (function () {
    function ListFilterDate(el, value) {
        this.hidden = el.querySelector(".admin_table_filter_item");
        this.from = el.querySelector(".admin_filter_layout_date_from");
        this.to = el.querySelector(".admin_filter_layout_date_to");
        this.from.addEventListener("input", this.changed.bind(this));
        this.from.addEventListener("change", this.changed.bind(this));
        this.to.addEventListener("input", this.changed.bind(this));
        this.to.addEventListener("change", this.changed.bind(this));
        this.setValue(value);
    }
    ListFilterDate.prototype.setValue = function (value) {
        if (!value) {
            return;
        }
        var splited = value.split(",");
        if (splited.length == 2) {
            this.from.value = splited[0];
            this.to.value = splited[1];
        }
        this.hidden.value = value;
    };
    ListFilterDate.prototype.changed = function () {
        var val = "";
        if (this.from.value || this.to.value) {
            val = this.from.value + "," + this.to.value;
        }
        this.hidden.value = val;
        var event = new Event('change');
        this.hidden.dispatchEvent(event);
    };
    return ListFilterDate;
}());
function bindLists() {
    var els = document.getElementsByClassName("admin_list");
    for (var i = 0; i < els.length; i++) {
        new List(els[i], document.querySelector(".admin_tablesettings_buttons"));
    }
}
var List = (function () {
    function List(el, openbutton) {
        var _this = this;
        this.el = el;
        this.settingsRow = this.el.querySelector(".admin_list_settingsrow");
        this.settingsRowColumn = this.el.querySelector(".admin_list_settingsrow_column");
        this.settingsEl = this.el.querySelector(".admin_tablesettings");
        this.settingsCheckbox = this.el.querySelector(".admin_list_showmore");
        this.settingsCheckbox.addEventListener("change", this.settingsCheckboxChange.bind(this));
        this.settingsCheckboxChange();
        this.exportButton = this.el.querySelector(".admin_exportbutton");
        var urlParams = new URLSearchParams(window.location.search);
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
        var visibleColumnsArr = visibleColumnsStr.split(",");
        var visibleColumnsMap = {};
        for (var i = 0; i < visibleColumnsArr.length; i++) {
            visibleColumnsMap[visibleColumnsArr[i]] = true;
        }
        this.itemsPerPage = parseInt(el.getAttribute("data-items-per-page"));
        this.paginationSelect = el.querySelector(".admin_tablesettings_pages");
        this.paginationSelect.addEventListener("change", this.load.bind(this));
        this.statsCheckbox = el.querySelector(".admin_tablesettings_stats");
        this.statsCheckbox.addEventListener("change", function () {
            _this.filterChanged();
        });
        this.statsCheckboxSelectCount = el.querySelector(".admin_tablesettings_stats_limit");
        this.statsCheckboxSelectCount.addEventListener("change", function () {
            _this.filterChanged();
        });
        this.statsContainer = el.querySelector(".admin_tablesettings_stats_container");
        this.bindOptions(visibleColumnsMap);
        this.bindOrder();
    }
    List.prototype.settingsCheckboxChange = function () {
        if (this.settingsCheckbox.checked) {
            this.settingsRow.classList.add("admin_list_settingsrow-visible");
        }
        else {
            this.settingsRow.classList.remove("admin_list_settingsrow-visible");
        }
    };
    List.prototype.load = function () {
        var _this = this;
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
        var filterData = this.getFilterData();
        for (var k in filterData) {
            params[k] = filterData[k];
        }
        this.colorActiveFilterItems();
        var selectedPages = parseInt(this.paginationSelect.value);
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
        this.exportButton.setAttribute("href", this.adminPrefix + "/" + this.typeName + encodeParams(params));
        params["_format"] = "json";
        encoded = encodeParams(params);
        request.open("GET", this.adminPrefix + "/" + this.typeName + encoded, true);
        request.addEventListener("load", function () {
            _this.tbody.innerHTML = "";
            if (request.status == 200) {
                var response = JSON.parse(request.response);
                _this.tbody.innerHTML = response.Content;
                var countStr = response.CountStr;
                _this.el.querySelector(".admin_table_count").textContent = countStr;
                _this.statsContainer.innerHTML = response.StatsStr;
                bindOrder();
                _this.bindPagination();
                _this.bindClick();
                _this.tbody.classList.remove("admin_table_loading");
            }
            else {
                console.error("error while loading list");
            }
            _this.progress.classList.add("admin_table_progress-inactive");
        });
        request.send(JSON.stringify({}));
    };
    List.prototype.bindOptions = function (visibleColumnsMap) {
        var _this = this;
        var columns = this.el.querySelectorAll(".admin_tablesettings_column");
        for (var i = 0; i < columns.length; i++) {
            var columnName = columns[i].getAttribute("data-column-name");
            if (visibleColumnsMap[columnName]) {
                columns[i].checked = true;
            }
            columns[i].addEventListener("change", function () {
                _this.changedOptions();
            });
        }
        this.changedOptions();
    };
    List.prototype.changedOptions = function () {
        var columns = this.getSelectedColumnsMap();
        var headers = this.el.querySelectorAll(".admin_list_orderitem");
        for (var i = 0; i < headers.length; i++) {
            var name = headers[i].getAttribute("data-name");
            if (columns[name]) {
                headers[i].classList.remove("hidden");
            }
            else {
                headers[i].classList.add("hidden");
            }
        }
        var filters = this.el.querySelectorAll(".admin_list_filteritem");
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
    };
    List.prototype.colorActiveFilterItems = function () {
        var itemsToColor = this.getFilterData();
        var filterItems = this.el.querySelectorAll(".admin_list_filteritem");
        for (var i = 0; i < filterItems.length; i++) {
            var item = filterItems[i];
            var name_1 = item.getAttribute("data-name");
            if (itemsToColor[name_1]) {
                item.classList.add("admin_list_filteritem-colored");
            }
            else {
                item.classList.remove("admin_list_filteritem-colored");
            }
        }
    };
    List.prototype.paginationChange = function (e) {
        var el = e.target;
        var page = parseInt(el.getAttribute("data-page"));
        this.page = page;
        this.load();
        e.preventDefault();
        return false;
    };
    List.prototype.bindPagination = function () {
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
    };
    List.prototype.bindClick = function () {
        var rows = this.el.querySelectorAll(".admin_table_row");
        for (var i = 0; i < rows.length; i++) {
            var row = rows[i];
            var id = row.getAttribute("data-id");
            row.addEventListener("click", function (e) {
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
            buttons.addEventListener("click", function (e) {
                var url = e.target.getAttribute("href");
                if (url != "") {
                    window.location.href = url;
                    e.preventDefault();
                    e.stopPropagation();
                    return false;
                }
            });
        }
    };
    List.prototype.bindOrder = function () {
        var _this = this;
        this.renderOrder();
        var headers = this.el.querySelectorAll(".admin_list_orderitem-canorder");
        for (var i = 0; i < headers.length; i++) {
            var header = headers[i];
            header.addEventListener("click", function (e) {
                var el = e.target;
                var name = el.getAttribute("data-name");
                if (name == _this.orderColumn) {
                    if (_this.orderDesc) {
                        _this.orderDesc = false;
                    }
                    else {
                        _this.orderDesc = true;
                    }
                }
                else {
                    _this.orderColumn = name;
                    _this.orderDesc = false;
                }
                _this.renderOrder();
                _this.load();
                e.preventDefault();
                return false;
            });
        }
    };
    List.prototype.renderOrder = function () {
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
    };
    List.prototype.getSelectedColumnsStr = function () {
        var ret = [];
        var checked = this.el.querySelectorAll(".admin_tablesettings_column:checked");
        for (var i = 0; i < checked.length; i++) {
            ret.push(checked[i].getAttribute("data-column-name"));
        }
        return ret.join(",");
    };
    List.prototype.getSelectedColumnsMap = function () {
        var columns = {};
        var checked = this.el.querySelectorAll(".admin_tablesettings_column:checked");
        for (var i = 0; i < checked.length; i++) {
            columns[checked[i].getAttribute("data-column-name")] = true;
        }
        return columns;
    };
    List.prototype.getFilterData = function () {
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
    };
    List.prototype.bindFilter = function (params) {
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
    };
    List.prototype.inputListener = function (e) {
        if (e.keyCode == 9 || e.keyCode == 16 || e.keyCode == 17 || e.keyCode == 18) {
            return;
        }
        this.filterChanged();
    };
    List.prototype.filterChanged = function () {
        this.colorActiveFilterItems();
        this.tbody.classList.add("admin_table_loading");
        this.page = 1;
        this.changed = true;
        this.changedTimestamp = Date.now();
        this.progress.classList.remove("admin_table_progress-inactive");
    };
    List.prototype.bindFilterRelation = function (el, value) {
        new ListFilterRelations(el, value, this);
    };
    List.prototype.bindFilterDate = function (el, value) {
        new ListFilterDate(el, value);
    };
    List.prototype.inputPeriodicListener = function () {
        var _this = this;
        setInterval(function () {
            if (_this.changed == true && Date.now() - _this.changedTimestamp > 500) {
                _this.changed = false;
                _this.load();
            }
        }, 200);
    };
    return List;
}());
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
            var ajaxPath = adminPrefix + "/_api/order/" + typ;
            var order = [];
            var rows = el.getElementsByClassName("admin_table_row");
            Array.prototype.forEach.call(rows, function (item, i) {
                order.push(parseInt(item.getAttribute("data-id")));
            });
            var request = new XMLHttpRequest();
            request.open("POST", ajaxPath, true);
            request.addEventListener("load", function () {
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
var MarkdownEditor = (function () {
    function MarkdownEditor(el) {
        var _this = this;
        this.el = el;
        this.textarea = el.querySelector(".textarea");
        this.preview = el.querySelector(".admin_markdown_preview");
        new Autoresize(this.textarea);
        var prefix = document.body.getAttribute("data-admin-prefix");
        var helpLink = el.querySelector(".admin_markdown_show_help");
        helpLink.setAttribute("href", prefix + "/_help/markdown");
        this.lastChanged = Date.now();
        this.changed = false;
        var showChange = el.querySelector(".admin_markdown_preview_show");
        showChange.addEventListener("change", function () {
            _this.preview.classList.toggle("hidden");
        });
        setInterval(function () {
            if (_this.changed && (Date.now() - _this.lastChanged > 500)) {
                _this.loadPreview();
            }
        }, 100);
        this.textarea.addEventListener("change", this.textareaChanged.bind(this));
        this.textarea.addEventListener("keyup", this.textareaChanged.bind(this));
        this.loadPreview();
        this.bindCommands();
        this.bindShortcuts();
    }
    MarkdownEditor.prototype.bindCommands = function () {
        var _this = this;
        var btns = this.el.querySelectorAll(".admin_markdown_command");
        for (var i = 0; i < btns.length; i++) {
            btns[i].addEventListener("mousedown", function (e) {
                var cmd = e.target.getAttribute("data-cmd");
                _this.executeCommand(cmd);
                e.preventDefault();
                return false;
            });
        }
    };
    MarkdownEditor.prototype.bindShortcuts = function () {
        var _this = this;
        this.textarea.addEventListener("keydown", function (e) {
            if (e.metaKey == false && e.ctrlKey == false) {
                return;
            }
            switch (e.keyCode) {
                case 66:
                    _this.executeCommand("b");
                    break;
                case 73:
                    _this.executeCommand("i");
                    break;
                case 75:
                    _this.executeCommand("h2");
                    break;
                case 85:
                    _this.executeCommand("a");
                    break;
            }
        });
    };
    MarkdownEditor.prototype.executeCommand = function (commandName) {
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
    };
    MarkdownEditor.prototype.setAroundMarkdown = function (before, after) {
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
    };
    MarkdownEditor.prototype.textareaChanged = function () {
        this.changed = true;
        this.lastChanged = Date.now();
    };
    MarkdownEditor.prototype.loadPreview = function () {
        var _this = this;
        this.changed = false;
        var request = new XMLHttpRequest();
        request.open("POST", document.body.getAttribute("data-admin-prefix") + "/_api/markdown", true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                _this.preview.innerHTML = JSON.parse(request.response);
            }
            else {
                console.error("Error while loading markdown preview.");
            }
        });
        request.send(this.textarea.value);
    };
    return MarkdownEditor;
}());
function bindTimestamps() {
    var elements = document.querySelectorAll(".admin_timestamp");
    Array.prototype.forEach.call(elements, function (el, i) {
        new Timestamp(el);
    });
}
var Timestamp = (function () {
    function Timestamp(el) {
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
    Timestamp.prototype.setTimestamp = function (v) {
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
    };
    Timestamp.prototype.initClock = function () {
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
    };
    Timestamp.prototype.saveValue = function () {
        var str = this.elTsDate.value + " " + this.elTsHour.value + ":" + this.elTsMinute.value;
        if (this.elTsDate.value == "") {
            str = "";
        }
        this.elTsInput.value = str;
    };
    return Timestamp;
}());
function bindRelations() {
    var elements = document.querySelectorAll(".admin_item_relation");
    Array.prototype.forEach.call(elements, function (el, i) {
        new RelationPicker(el);
    });
}
var RelationPicker = (function () {
    function RelationPicker(el) {
        var _this = this;
        this.selectedClass = "admin_item_relation_picker_suggestion-selected";
        this.input = el.getElementsByTagName("input")[0];
        this.previewContainer = el.querySelector(".admin_item_relation_preview");
        this.relationName = el.getAttribute("data-relation");
        this.progress = el.querySelector("progress");
        this.changeSection = el.querySelector(".admin_item_relation_change");
        this.changeButton = el.querySelector(".admin_item_relation_change_btn");
        this.changeButton.addEventListener("click", function () {
            _this.input.value = "0";
            _this.showSearch();
            _this.pickerInput.focus();
        });
        this.suggestionsEl = el.querySelector(".admin_item_relation_picker_suggestions_content");
        this.suggestions = [];
        this.picker = el.querySelector(".admin_item_relation_picker");
        this.pickerInput = this.picker.querySelector("input");
        this.pickerInput.addEventListener("input", function () {
            _this.getSuggestions(_this.pickerInput.value);
        });
        this.pickerInput.addEventListener("blur", function () {
            _this.suggestionsEl.classList.add("hidden");
        });
        this.pickerInput.addEventListener("focus", function () {
            _this.suggestionsEl.classList.remove("hidden");
            _this.getSuggestions(_this.pickerInput.value);
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
    RelationPicker.prototype.getData = function () {
        var _this = this;
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/preview/" + this.relationName + "/" + this.input.value, true);
        request.addEventListener("load", function () {
            _this.progress.classList.add("hidden");
            if (request.status == 200) {
                _this.showPreview(JSON.parse(request.response));
            }
            else {
                _this.showSearch();
            }
        });
        request.send();
    };
    RelationPicker.prototype.showPreview = function (data) {
        this.previewContainer.textContent = "";
        this.input.value = data.ID;
        var el = this.createPreview(data, true);
        this.previewContainer.appendChild(el);
        this.previewContainer.classList.remove("hidden");
        this.changeSection.classList.remove("hidden");
        this.picker.classList.add("hidden");
    };
    RelationPicker.prototype.showSearch = function () {
        this.previewContainer.classList.add("hidden");
        this.changeSection.classList.add("hidden");
        this.picker.classList.remove("hidden");
        this.suggestions = [];
        this.suggestionsEl.innerText = "";
        this.pickerInput.value = "";
    };
    RelationPicker.prototype.getSuggestions = function (q) {
        var _this = this;
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var request = new XMLHttpRequest();
        request.open("GET", adminPrefix + "/_api/search/" + this.relationName + "?q=" + encodeURIComponent(q), true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                if (q != _this.pickerInput.value) {
                    return;
                }
                var data = JSON.parse(request.response);
                _this.suggestions = data;
                _this.suggestionsEl.innerText = "";
                for (var i = 0; i < data.length; i++) {
                    var item = data[i];
                    var el = _this.createPreview(item, false);
                    el.classList.add("admin_item_relation_picker_suggestion");
                    el.setAttribute("data-position", i + "");
                    el.addEventListener("mousedown", _this.suggestionClick.bind(_this));
                    el.addEventListener("mouseenter", _this.suggestionSelect.bind(_this));
                    _this.suggestionsEl.appendChild(el);
                }
            }
            else {
                console.log("Error while searching");
            }
        });
        request.send();
    };
    RelationPicker.prototype.suggestionClick = function () {
        var selected = this.getSelected();
        if (selected >= 0) {
            this.showPreview(this.suggestions[selected]);
        }
    };
    RelationPicker.prototype.suggestionSelect = function (e) {
        var target = e.currentTarget;
        var position = parseInt(target.getAttribute("data-position"));
        this.select(position);
    };
    RelationPicker.prototype.getSelected = function () {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        return parseInt(selected.getAttribute("data-position"));
    };
    RelationPicker.prototype.unselect = function () {
        var selected = this.suggestionsEl.querySelector("." + this.selectedClass);
        if (!selected) {
            return -1;
        }
        selected.classList.remove(this.selectedClass);
        return parseInt(selected.getAttribute("data-position"));
    };
    RelationPicker.prototype.select = function (i) {
        this.unselect();
        this.suggestionsEl.querySelectorAll(".admin_preview")[i].classList.add(this.selectedClass);
    };
    RelationPicker.prototype.suggestionInput = function (e) {
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
    };
    RelationPicker.prototype.createPreview = function (data, anchor) {
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
    };
    return RelationPicker;
}());
function bindPlacesView() {
    var els = document.querySelectorAll(".admin_item_view_place");
    for (var i = 0; i < els.length; i++) {
        new PlacesView(els[i]);
    }
}
var PlacesView = (function () {
    function PlacesView(el) {
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
    return PlacesView;
}());
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
        searchBox.addListener('places_changed', function () {
            var places = searchBox.getPlaces();
            if (places.length > 0) {
                map.fitBounds(places[0].geometry.viewport);
                marker.setPosition({ lat: places[0].geometry.location.lat(), lng: places[0].geometry.location.lng() });
                marker.setVisible(true);
            }
        });
        searchInput.addEventListener("keydown", function (e) {
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
var Form = (function () {
    function Form(el) {
        var _this = this;
        this.dirty = false;
        el.addEventListener("submit", function () {
            _this.dirty = false;
        });
        var els = el.querySelectorAll(".form_watcher");
        for (var i = 0; i < els.length; i++) {
            var input = els[i];
            input.addEventListener("input", function () {
                _this.dirty = true;
            });
            input.addEventListener("change", function () {
                _this.dirty = true;
            });
        }
        window.addEventListener("beforeunload", function (e) {
            if (_this.dirty) {
                var confirmationMessage = "Chcete opustit stránku bez uložení změn?";
                e.returnValue = confirmationMessage;
                return confirmationMessage;
            }
        });
    }
    return Form;
}());
function bindDatePicker() {
    var dates = document.querySelectorAll(".form_input-date");
    for (var i = 0; i < dates.length; i++) {
        var dateEl = dates[i];
        new DatePicker(dateEl);
    }
}
var DatePicker = (function () {
    function DatePicker(el) {
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
            onSelect: function (date) {
                el.value = pd.toString();
            },
            toString: function (date) {
                var day = date.getDate();
                var dayStr = "" + day;
                if (day < 10) {
                    dayStr = "0" + dayStr;
                }
                var month = date.getMonth() + 1;
                var monthStr = "" + month;
                if (month < 10) {
                    monthStr = "0" + monthStr;
                }
                var year = date.getFullYear();
                var ret = year + "-" + monthStr + "-" + dayStr;
                return ret;
            }
        });
    }
    return DatePicker;
}());
function prettyDate(date) {
    var day = date.getDate();
    var month = date.getMonth() + 1;
    var year = date.getFullYear();
    return day + ". " + month + ". " + year;
}
function bindDropdowns() {
    var els = document.querySelectorAll(".admin_dropdown");
    for (var i = 0; i < els.length; i++) {
        new Dropdown(els[i]);
    }
}
var Dropdown = (function () {
    function Dropdown(el) {
        this.targetEl = el.querySelector(".admin_dropdown_target");
        this.contentEl = el.querySelector(".admin_dropdown_content");
        this.targetEl.addEventListener("mousedown", function (e) {
            if (document.activeElement == el) {
                el.blur();
                e.preventDefault();
                return false;
            }
        });
    }
    return Dropdown;
}());
function bindSearch() {
    var els = document.querySelectorAll(".admin_header_search");
    for (var i = 0; i < els.length; i++) {
        new SearchForm(els[i]);
    }
}
var SearchForm = (function () {
    function SearchForm(el) {
        var _this = this;
        this.searchForm = el;
        this.searchInput = el.querySelector(".admin_header_search_input");
        this.suggestionsEl = el.querySelector(".admin_header_search_suggestions");
        this.searchInput.value = document.body.getAttribute("data-search-query");
        this.searchInput.addEventListener("input", function () {
            _this.suggestions = [];
            _this.dirty = true;
            _this.lastChanged = Date.now();
            return false;
        });
        this.searchInput.addEventListener("blur", function () {
        });
        window.setInterval(function () {
            if (_this.dirty && Date.now() - _this.lastChanged > 100) {
                _this.loadSuggestions();
            }
        }, 30);
        this.searchInput.addEventListener("keydown", function (e) {
            if (!_this.suggestions || _this.suggestions.length == 0) {
                return;
            }
            switch (e.keyCode) {
                case 13:
                    var i = _this.getSelected();
                    if (i >= 0) {
                        var child = _this.suggestions[i];
                        if (child) {
                            window.location.href = child.getAttribute("href");
                        }
                        e.preventDefault();
                        return true;
                    }
                    return false;
                case 38:
                    var i = _this.getSelected();
                    if (i < 1) {
                        i = _this.suggestions.length - 1;
                    }
                    else {
                        i = i - 1;
                    }
                    _this.setSelected(i);
                    e.preventDefault();
                    return false;
                case 40:
                    var i = _this.getSelected();
                    if (i >= 0) {
                        i += 1;
                        i = i % _this.suggestions.length;
                    }
                    else {
                        i = 0;
                    }
                    _this.setSelected(i);
                    e.preventDefault();
                    return false;
            }
        });
    }
    SearchForm.prototype.loadSuggestions = function () {
        var _this = this;
        this.dirty = false;
        var suggestText = this.searchInput.value;
        var request = new XMLHttpRequest();
        var adminPrefix = document.body.getAttribute("data-admin-prefix");
        var url = adminPrefix + "/_search_suggest" + encodeParams({ "q": this.searchInput.value });
        request.open("GET", url);
        request.addEventListener("load", function () {
            if (suggestText != _this.searchInput.value) {
                return;
            }
            if (request.status == 200) {
                _this.addSuggestions(request.response);
            }
            else {
                _this.suggestionsEl.classList.add("hidden");
                console.error("Error while loading item.");
            }
        });
        request.send();
    };
    SearchForm.prototype.addSuggestions = function (content) {
        var _this = this;
        this.suggestionsEl.innerHTML = content;
        this.suggestionsEl.classList.remove("hidden");
        this.suggestions = this.suggestionsEl.querySelectorAll(".admin_search_suggestion");
        for (var i = 0; i < this.suggestions.length; i++) {
            var suggestion = this.suggestions[i];
            suggestion.addEventListener("touchend", function (e) {
                var el = e.currentTarget;
                window.location.href = el.getAttribute("href");
            });
            suggestion.addEventListener("click", function (e) {
                return false;
            });
            suggestion.addEventListener("mouseenter", function (e) {
                _this.deselect();
                var el = e.currentTarget;
                _this.setSelected(parseInt(el.getAttribute("data-position")));
            });
        }
    };
    SearchForm.prototype.deselect = function () {
        var el = this.suggestionsEl.querySelector(".admin_search_suggestion-selected");
        if (el) {
            el.classList.remove("admin_search_suggestion-selected");
        }
    };
    SearchForm.prototype.getSelected = function () {
        var el = this.suggestionsEl.querySelector(".admin_search_suggestion-selected");
        if (el) {
            return parseInt(el.getAttribute("data-position"));
        }
        return -1;
    };
    SearchForm.prototype.setSelected = function (position) {
        this.deselect();
        if (position >= 0) {
            var els = this.suggestionsEl.querySelectorAll(".admin_search_suggestion");
            els[position].classList.add("admin_search_suggestion-selected");
        }
    };
    return SearchForm;
}());
function bindEshopControl() {
    var el = document.querySelector(".eshop_control");
    if (el) {
        new EshopControl(el);
    }
}
var EshopControl = (function () {
    function EshopControl(el) {
        this.canvas = el.querySelector(".eshop_control_canvas");
        this.video = el.querySelector(".eshop_control_video");
        this.context = this.canvas.getContext('2d');
        this.initCamera();
        this.captureImage();
    }
    EshopControl.prototype.initCamera = function () {
        var _this = this;
        navigator.mediaDevices.getUserMedia({ video: true, audio: false })
            .then(function (stream) {
            _this.video.srcObject = stream;
            _this.video.play();
        });
    };
    EshopControl.prototype.captureImage = function () {
        var w = this.video.videoWidth;
        var h = this.video.videoHeight;
        this.canvas.width = w;
        this.canvas.height = h;
        this.context.drawImage(this.video, 0, 0, w, h);
        if (w > 0 && h > 0) {
            var imageData = this.context.getImageData(0, 0, w, h);
            try {
                var code = jsQR(imageData.data, w, h);
                if (code) {
                    console.log(code.data);
                }
            }
            catch (error) {
                console.log(error);
            }
        }
        requestAnimationFrame(this.captureImage.bind(this));
    };
    return EshopControl;
}());
function bindMainMenu() {
    var el = document.querySelector(".admin_layout_left");
    if (el) {
        new MainMenu(el);
    }
}
var MainMenu = (function () {
    function MainMenu(leftEl) {
        this.leftEl = leftEl;
        this.menuEl = document.querySelector(".admin_header_container_menu");
        this.menuEl.addEventListener("click", this.menuClick.bind(this));
        this.scrollTo(this.loadFromStorage());
        this.leftEl.addEventListener("scroll", this.scrollHandler.bind(this));
    }
    MainMenu.prototype.scrollHandler = function () {
        this.saveToStorage(this.leftEl.scrollTop);
    };
    MainMenu.prototype.saveToStorage = function (position) {
        window.localStorage["left_menu_position"] = position;
    };
    MainMenu.prototype.menuClick = function () {
        this.leftEl.classList.toggle("admin_layout_left-visible");
    };
    MainMenu.prototype.loadFromStorage = function () {
        var pos = window.localStorage["left_menu_position"];
        if (pos) {
            return parseInt(pos);
        }
        return 0;
    };
    MainMenu.prototype.scrollTo = function (position) {
        this.leftEl.scrollTo(0, position);
    };
    return MainMenu;
}());
function bindRelationList() {
    var els = document.getElementsByClassName("admin_relationlist");
    for (var i = 0; i < els.length; i++) {
        new RelationList(els[i]);
    }
}
var RelationList = (function () {
    function RelationList(el) {
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
    RelationList.prototype.load = function () {
        var _this = this;
        this.loadingEl.classList.remove("hidden");
        this.moreEl.classList.add("hidden");
        var request = new XMLHttpRequest();
        request.open("POST", this.adminPrefix + "/_api/relationlist", true);
        request.addEventListener("load", function () {
            _this.loadingEl.classList.add("hidden");
            if (request.status == 200) {
                _this.offset += 10;
                var parentEl = document.createElement("div");
                parentEl.innerHTML = request.response;
                var parentAr = [];
                for (var i = 0; i < parentEl.children.length; i++) {
                    parentAr.push(parentEl.children[i]);
                }
                for (var i = 0; i < parentAr.length; i++) {
                    _this.targetEl.appendChild(parentAr[i]);
                }
                if (_this.offset < _this.count) {
                    _this.moreEl.classList.remove("hidden");
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
            Count: 10
        }));
    };
    return RelationList;
}());
function bindTaskMonitor() {
    var el = document.querySelector(".taskmonitorcontainer");
    if (el) {
        new TaskMonitor(el);
    }
}
var TaskMonitor = (function () {
    function TaskMonitor(el) {
        this.el = el;
        window.setInterval(this.load.bind(this), 1000);
    }
    TaskMonitor.prototype.load = function () {
        var _this = this;
        var request = new XMLHttpRequest();
        request.open("GET", "/admin/_tasks/running", true);
        request.addEventListener("load", function () {
            _this.el.innerHTML = "";
            if (request.status == 200) {
                _this.el.innerHTML = request.response;
            }
            else {
                console.error("error while loading list");
            }
        });
        request.send();
    };
    return TaskMonitor;
}());
function bindNotifications() {
    new NotificationCenter(document.querySelector(".notification_center"));
}
var NotificationCenter = (function () {
    function NotificationCenter(el) {
        this.el = el;
        this.notifications = Array();
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.loadNotifications();
    }
    NotificationCenter.prototype.loadNotifications = function () {
        var _this = this;
        return;
        var request = new XMLHttpRequest();
        request.open("GET", this.adminPrefix + "/_api/notifications", true);
        request.addEventListener("load", function () {
            if (request.status == 200) {
                var notifications = JSON.parse(request.response);
                notifications.Views.forEach(function (item) {
                    var notification = _this.createNotification(item);
                    _this.notifications.push(notification);
                    _this.el.appendChild(notification.el);
                });
            }
            else {
                console.log("failed to load notifications");
            }
        });
        request.send();
    };
    NotificationCenter.prototype.createNotification = function (data) {
        return new NotificationItem(data);
    };
    return NotificationCenter;
}());
var NotificationItem = (function () {
    function NotificationItem(data) {
        this.adminPrefix = document.body.getAttribute("data-admin-prefix");
        this.createElement(data);
    }
    NotificationItem.prototype.createElement = function (data) {
        var ret;
        ret = document.createElement("div");
        ret.innerHTML = "\n            <div class=\"notification\">\n                <div class=\"notification_close\"></div>\n                <div class=\"notification_left\"></div>\n                <div class=\"notification_right\">\n                    <div class=\"notification_name\">" + e(data.Name) + "</div>\n                </div>\n            </div>        \n        ";
        this.el = ret.children[0];
        this.el.querySelector(".notification_close").addEventListener("click", this.closeNotification.bind(this));
    };
    NotificationItem.prototype.closeNotification = function () {
        this.el.classList.add("notification-closed");
        fetch(this.adminPrefix + "/_api/notification/" + this.uuid, { method: "DELETE" }).then(console.log).then(function (e) {
            console.log(e);
        });
    };
    return NotificationItem;
}());
document.addEventListener("DOMContentLoaded", function () {
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
    bindEshopControl();
    bindMainMenu();
    bindRelationList();
    bindTaskMonitor();
    bindNotifications();
});
function bindFlashMessages() {
    var messages = document.querySelectorAll(".flash_message");
    for (var i = 0; i < messages.length; i++) {
        var message = messages[i];
        message.addEventListener("click", function (e) {
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
    document.addEventListener("scroll", function (event) {
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
function DOMinsertChildAtIndex(
  parent: HTMLElement,
  child: HTMLElement,
  index: number
) {
  if (index >= parent.children.length) {
    parent.appendChild(child);
  } else {
    if (index < 0) {
      index = 0;
    }
    parent.insertBefore(child, parent.children[index]);
  }
}

function encodeParams(data: any) {
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

function e(str: String): String {
  return escapeHTML(str);
}

function escapeHTML(str: String): String {
  str = str.split("&").join("&amp;");
  str = str.split("<").join("&lt;");
  str = str.split(">").join("&gt;");
  str = str.split('"').join("&quot;");
  str = str.split("'").join("&#39;");
  //str = str.split("&").join("&amp;");
  //str = str.replaceAll("&", "&amp;");
  return str;
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

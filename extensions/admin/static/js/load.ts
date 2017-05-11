document.addEventListener("DOMContentLoaded", () => {
  //bindOrder();
  bindMarkdowns();
  bindTimestamps();
  bindRelations();
  bindImagePickers();
  //bindDelete();
  bindClickAndStay();
  bindLists();
});

function bindClickAndStay() {
  var els = document.getElementsByName("_submit_and_stay");
  var elsClicked = document.getElementsByName("_submit_and_stay_clicked");

  if (els.length == 1 && elsClicked.length == 1) {
    els[0].addEventListener("click", () => {
      (<HTMLInputElement>elsClicked[0]).value = "true";
    })
  }
}
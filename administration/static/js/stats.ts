function bindStats() {
  var elements = document.querySelectorAll(".admin_stats_pie");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    new PieChart(<HTMLDivElement>el);
  });

  var elements = document.querySelectorAll(".admin_stats_timeline");
  Array.prototype.forEach.call(elements, function(el: HTMLElement, i: number){
    new Timeline(<HTMLDivElement>el);
  });
}

class PieChart {

  constructor(el: HTMLDivElement) {
    var canvas = <HTMLCanvasElement>el.querySelector("canvas");
    var ctx = canvas.getContext('2d');

    var labelA = el.getAttribute("data-label-a");
    var labelB = el.getAttribute("data-label-b");

    var valueA = parseInt(el.getAttribute("data-value-a"));
    var valueB = parseInt(el.getAttribute("data-value-b"));

    var data = {
        datasets: [{
            data: [valueA, valueB],
            backgroundColor: ["#4078c0", "#eee"]
        }],
        labels: [
            labelA,
            labelB
        ]
    };

    var myChart = new Chart(ctx, {
      type: "pie",
      data: data,
      options: {
        "responsive": false
      }
    });
  }
}

class Timeline {

  constructor(el: HTMLDivElement) {
    return;
    var canvas = <HTMLCanvasElement>el.querySelector("canvas");
    var ctx = canvas.getContext('2d');

    var data = {
        labels: ["January", "February", "March"],
        datasets: [{
            //label: "My First dataset",
            backgroundColor: '#4078c0',
            data: [20, 30, 40]
        }]
    };
    var myChart = new Chart(ctx, {
      type: "bar",
      data: data
    });
  }

}
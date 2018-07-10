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
  adminPrefix: string;
  ctx: CanvasRenderingContext2D;

  constructor(el: HTMLDivElement) {
    this.adminPrefix = document.body.getAttribute("data-admin-prefix");
    var resource = el.getAttribute("data-resource");
    var field = el.getAttribute("data-field");

    var canvas = <HTMLCanvasElement>el.querySelector("canvas");
    this.ctx = canvas.getContext('2d');

    this.loadData(resource, field);
  }

  loadData(resource: string, field: string) {
    var data = {
      resource: resource,
      field: field
    }

    var request = new XMLHttpRequest();
    request.open("POST", this.adminPrefix + "/_api/stats", true);
    request.addEventListener("load", () => {
      if (request.status == 200) {
        var parsed = JSON.parse(request.responseText);
        this.createChart(parsed.labels, parsed.values);
      } else {
        console.error("error while loading list");
      }
    });
    request.send(JSON.stringify(data));
  }

  createChart(labels: any, values: any) {
    var data = {
        labels: labels,
        datasets: [{
            backgroundColor: '#4078c0',
            data: values
        }]
    };
    var myChart = new Chart(this.ctx, {
      type: "bar",
      data: data,
      options: {
        legend: {
          display: false
        },
        scales: {
                yAxes: [{
                    ticks: {
                        beginAtZero: true
                    }
                }]
            }
      }
    });
  }

}
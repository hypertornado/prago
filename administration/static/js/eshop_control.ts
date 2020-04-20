function bindEshopControl() {
    var el = document.querySelector(".eshop_control");
    if (el) {
        new EshopControl(<HTMLDivElement>el);
    }
}

//https://github.com/cozmo/jsQR
class EshopControl {

    canvas: HTMLCanvasElement;
    context: CanvasRenderingContext2D;
    video: HTMLVideoElement;

    constructor(el: HTMLDivElement) {
        this.canvas = el.querySelector(".eshop_control_canvas");
        this.video = el.querySelector(".eshop_control_video");
        this.context = this.canvas.getContext('2d');

        console.log(this.canvas);
        console.log(this.video);
        this.initCamera();
        this.captureImage();
    }

    initCamera() {
        console.log(navigator.mediaDevices);
        console.log(navigator.mediaDevices.getUserMedia({ video: true, audio: false }));
        navigator.mediaDevices.getUserMedia({ video: true, audio: false })
        .then(
            (stream: MediaStream) => {
                this.video.srcObject = stream;
                this.video.play();
            }
        )
    }

    captureImage() {

        /*this.canvas.width = this.width;
        this.canvas.height = this.height;

        var pixelRatio = window.devicePixelRatio;
        var canvasWidth = Math.floor(this.width / pixelRatio);
        var canvasHeight = Math.floor(this.height / pixelRatio);

        this.canvas.setAttribute("style", "width: " + canvasWidth + "px; height: " + canvasHeight + "px;");*/


        var w = this.video.videoWidth;
        var h = this.video.videoHeight;
        //console.log(this.video.videoWidth, this.video.videoHeight)

        this.canvas.width = w;
        this.canvas.height = h;

        this.context.drawImage(this.video ,0,0,w,h);

        if (w > 0 && h > 0) {
            var imageData = this.context.getImageData(0, 0, w, h);
            try {
                //console.log(imageData);
                //console.log(w, h);

                //@ts-ignore
                var code = jsQR(imageData.data, w, h);
                if (code) {
                    console.log(code.data);
                }
            } catch (error) {
                console.log(error);
            }
        }
        requestAnimationFrame(this.captureImage.bind(this));
    }

}
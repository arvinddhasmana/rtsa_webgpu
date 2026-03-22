const fs = require('fs');
const PNG = require('pngjs').PNG;

fs.createReadStream('e2e/snapshots/fusion-tracks-visible.png')
  .pipe(new PNG())
  .on('parsed', function() {
    let coloredPixels = 0;
    for (let y = 0; y < this.height; y++) {
      for (let x = 0; x < this.width; x++) {
        let idx = (this.width * y + x) << 2;
        let r = this.data[idx];
        let g = this.data[idx+1];
        let b = this.data[idx+2];
        let max = Math.max(r, g, b);
        let min = Math.min(r, g, b);
        if (max > 100 && max - min > 40) {
            coloredPixels++;
        }
      }
    }
    console.log("Colored pixels in e2e snapshot:", coloredPixels);
  });

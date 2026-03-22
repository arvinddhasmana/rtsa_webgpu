const PI = Math.PI;
function to_web_mercator(lon_rad, lat_rad) {
    let mx = lon_rad / (2.0 * PI) + 0.5;
    let my = 0.5 - Math.log(Math.tan(PI / 4.0 + lat_rad / 2.0)) / (2.0 * PI);
    return { mx, my };
}

console.log("Track:", to_web_mercator(55 * PI / 180, 25 * PI / 180));

function get_vp(centerLon, centerLat, scale, W, H) {
  const L = 1024 * scale;
  const cx = centerLon / 360.0 + 0.5;
  const sinLat = Math.sin(centerLat * Math.PI / 180.0);
  const cy = 0.5 - Math.log((1 + sinLat) / (1 - sinLat)) / (4 * Math.PI);
  const sx = 2 * L / W;
  const sy = -2 * L / H;
  const tx = -cx * sx;
  const ty = -cy * sy;
  return { cx, cy, sx, sy, tx, ty };
}
const W = 1920, H = 1080;
const vp = get_vp(54, 27, 16, W, H);
console.log("Viewport:", vp);

const t = to_web_mercator(55 * PI / 180, 25 * PI / 180);
const clip_x = vp.sx * t.mx + vp.tx;
const clip_y = vp.sy * t.my + vp.ty;
console.log("NDC:", { clip_x, clip_y });

const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ args: ['--enable-unsafe-webgpu', '--enable-features=WebGPU']});
  const page = await browser.newPage();
  await page.goto('http://localhost:5175/');
  await page.waitForTimeout(5000); // 5 seconds
  await page.screenshot({ path: 'test_current.png' });
  await browser.close();
})();

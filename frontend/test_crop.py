import asyncio
from playwright.async_api import async_playwright

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        page = await browser.new_page()
        await page.goto('http://localhost:5173/knowledge/documentation')
        await page.wait_for_selector('.ed-card')
        
        # Test 1: object-fit cover, left center
        await page.add_style_tag(content='''
            .ed-card-image-wrapper {
                width: 150px !important;
                padding: 0 !important;
                overflow: hidden !important;
            }
            .ed-card-image {
                width: 100% !important;
                height: 100% !important;
                max-height: none !important;
                object-fit: cover !important;
                object-position: left center !important;
            }
        ''')
        await page.screenshot(path='/Users/varun.kanade/.gemini/antigravity-ide/brain/a9084033-4be8-4e87-baa1-0002b38b3ff0/test_css_1.png')

        # Test 2: object-fit cover, 10% center
        await page.add_style_tag(content='''
            .ed-card-image {
                object-position: 10% center !important;
            }
        ''')
        await page.screenshot(path='/Users/varun.kanade/.gemini/antigravity-ide/brain/a9084033-4be8-4e87-baa1-0002b38b3ff0/test_css_2.png')
        
        # Test 3: scale image manually
        await page.add_style_tag(content='''
            .ed-card-image {
                width: 300% !important;
                height: 100% !important;
                object-fit: contain !important;
                object-position: left center !important;
                transform: translateX(-10%);
            }
        ''')
        await page.screenshot(path='/Users/varun.kanade/.gemini/antigravity-ide/brain/a9084033-4be8-4e87-baa1-0002b38b3ff0/test_css_3.png')

        await browser.close()

asyncio.run(main())

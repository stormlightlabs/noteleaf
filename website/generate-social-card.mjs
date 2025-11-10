import satori from 'satori';
import { html } from 'satori-html';
import sharp from 'sharp';
import { readFile, writeFile } from 'fs/promises';

// Color palette from internal/ui/palette.go
const colors = {
  bgBase: '#201F26',      // Pepper
  bgSecondary: '#2d2c35', // BBQ
  textPrimary: '#F1EFEF', // Salt
  textSecondary: '#BFBCC8', // Smoke
  primary: '#00A4FF',     // Malibu
  success: '#00FFB2',     // Julep
  accent: '#5CDFEA',      // Lichen
  warning: '#FF985A',     // Tang
};

async function generateSocialCard() {
  // Load bundled font from fontsource
  const fontData = await readFile('./node_modules/@fontsource/google-sans-code/files/google-sans-code-latin-400-normal.woff');

  const markup = html`
    <div
      style="
        display: flex;
        flex-direction: column;
        width: 100%;
        height: 100%;
        background: ${colors.bgBase};
        padding: 60px 80px;
        font-family: 'Google Sans Code';
      "
    >
      <div style="
        display: flex;
        flex-direction: column;
        flex: 1;
        justify-content: center;
      ">
        <div style="
          display: flex;
          align-items: center;
          margin-bottom: 40px;
        ">
          <span style="
            color: ${colors.success};
            font-size: 32px;
            margin-right: 16px;
          ">$</span>
          <span style="
            color: ${colors.textPrimary};
            font-size: 32px;
          ">noteleaf --help</span>
        </div>

        <div style="
          display: flex;
          flex-direction: column;
          background: ${colors.bgSecondary};
          border-left: 4px solid ${colors.primary};
          padding: 40px;
          border-radius: 8px;
        ">
          <div style="
            color: ${colors.primary};
            font-size: 72px;
            font-weight: bold;
            margin-bottom: 24px;
          ">noteleaf</div>

          <div style="
            color: ${colors.textSecondary};
            font-size: 36px;
            line-height: 1.5;
          ">
            A terminal-based productivity system for tasks, notes, and media tracking
          </div>
        </div>

        <div style="
          display: flex;
          margin-top: 40px;
          gap: 40px;
        ">
          <div style="display: flex; align-items: center;">
            <span style="color: ${colors.accent}; font-size: 24px; margin-right: 8px;">></span>
            <span style="color: ${colors.textPrimary}; font-size: 24px;">Tasks</span>
          </div>
          <div style="display: flex; align-items: center;">
            <span style="color: ${colors.accent}; font-size: 24px; margin-right: 8px;">></span>
            <span style="color: ${colors.textPrimary}; font-size: 24px;">Notes</span>
          </div>
          <div style="display: flex; align-items: center;">
            <span style="color: ${colors.accent}; font-size: 24px; margin-right: 8px;">></span>
            <span style="color: ${colors.textPrimary}; font-size: 24px;">Media</span>
          </div>
        </div>
      </div>

      <div style="
        display: flex;
        flex-direction: column;
        gap: 8px;
        margin-top: 40px;
        padding-top: 20px;
        border-top: 1px solid ${colors.bgSecondary};
      ">
        <div style="
          display: flex;
          justify-content: space-between;
          align-items: center;
        ">
          <div style="color: ${colors.textSecondary}; font-size: 20px;">
            stormlightlabs.github.io/noteleaf
          </div>
          <div style="
            color: ${colors.success};
            font-size: 20px;
          ">Open Source</div>
        </div>
        <div style="color: ${colors.accent}; font-size: 20px;">
          tangled.org/@desertthunder.dev/noteleaf
        </div>
      </div>
    </div>
  `;

  const svg = await satori(markup, {
    width: 1200,
    height: 630,
    fonts: [
      {
        name: 'Google Sans Code',
        data: fontData,
        weight: 400,
        style: 'normal',
      },
    ],
  });

  // Convert SVG to PNG using sharp
  const pngBuffer = await sharp(Buffer.from(svg))
    .png()
    .toBuffer();

  // Write the PNG file
  await writeFile('./static/img/social-card.png', pngBuffer);

  console.log('Social card generated successfully: static/img/social-card.png');
}

generateSocialCard().catch(console.error);

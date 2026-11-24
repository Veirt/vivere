# Asset prompts for watercolor archive fidelity

Target UI: `1920x1080` desktop character archive/detail page. Generate PNG assets on a clean white or warm ivory background so the background can be removed with Python. Do not include readable text in decorative assets. Keep everything soft, antique, low-contrast, and parchment-compatible.

Current implementation note: stable runtime assets now live in `web/public/archive-assets/`. They were derived from the existing `web/public/assets/*.png` files and cleaned with Python. Regenerate only if the current ornaments are not close enough.

## Status checklist

- [x] Detail page folio shell uses the archive watercolor system.
- [x] Character list page uses the same standardized folio width as detail.
- [x] Character list hero/search controls were simplified.
- [x] Character list sort controls are inside a collapsible refine panel.
- [x] Memories and voicelines only render in their dedicated tabs.
- [x] Existing generated watercolor assets were cleaned and wired into `web/public/archive-assets/`.
- [x] Divider and corner flourishes were created directly as SVG.
- [ ] Optional: regenerate the watercolor assets below if a closer 1:1 art match is needed.

## 1. Header watercolor banner wash

- **File**: `web/public/archive-assets/header-wash.png`
- **Size**: `1600x260`
- **Purpose**: faint watercolor clouds behind the character name/header area.
- **Prompt**:

```text
Soft antique watercolor wash for a fantasy illustrated archive UI header, clean white background, very pale warm beige with muted powder blue and dusty peach edges, asymmetrical cloudy stains, paper bleed texture, no hard outlines, no text, no characters, no objects, elegant Japanese game archive mood, low contrast, designed to sit on ivory parchment, wide horizontal composition with lots of quiet negative space.
```

## 2. Right-side botanical ornament

- **File**: `web/public/archive-assets/botanical-right.png`
- **Size**: `520x720`
- **Purpose**: delicate leaf/flower line art in right panels, like the reference.
- **Prompt**:

```text
Delicate botanical line art ornament on clean white background, thin sepia ink stems and small leaves, faint watercolor beige and pale lavender touches, elegant archive manuscript decoration, vertical composition leaning from lower right toward upper left, very light opacity feel, no text, no frame, refined fantasy character profile UI ornament.
```

## 3. Left-side botanical ornament

- **File**: `web/public/archive-assets/botanical-left.png`
- **Size**: `520x720`
- **Purpose**: optional sidebar/profile botanical wash.
- **Prompt**:

```text
Delicate botanical line art ornament on clean white background, thin sepia ink stems with small leaves and faint violet flowers, soft beige watercolor bleeding around the stem, vertical composition leaning from upper left toward lower right, antique illuminated manuscript mood, no text, no frame, low contrast for parchment UI decoration.
```

## 4. Open book illustration

- **File**: `web/public/archive-assets/open-book.png`
- **Size**: `760x450`
- **Purpose**: bottom-left decorative open book from the reference.
- **Prompt**:

```text
Small open antique book illustration on clean white background, soft watercolor and sepia ink, pages slightly curved, warm ivory paper, faint handwritten marks with no readable text, subtle powder blue and beige shadow wash around it, delicate fantasy archive manuscript style, viewed from slight top angle, isolated object, no characters.
```

## 5. Crest watermark

- **File**: `web/public/archive-assets/crest-watermark.png`
- **Size**: `560x560`
- **Purpose**: nearly invisible crest mark in hero panels.
- **Prompt**:

```text
Ornate fantasy shield crest watermark on clean white background, symmetrical heraldic emblem with fine sepia line art, elegant manuscript archive style, extremely low contrast pale beige ink, no readable text, no logo letters, centered composition, suitable as subtle UI watermark behind content.
```

## 6. Shield watermark

- **File**: `web/public/archive-assets/shield-watermark.png`
- **Size**: `560x560`
- **Purpose**: nearly invisible shield mark in combat panels.
- **Prompt**:

```text
Pale fantasy shield watermark on clean white background, fine sepia and ivory line art, subtle heraldic silhouette, manuscript archive style, extremely low contrast, no readable text, no letters, centered composition, suitable behind stat cards on a parchment UI.
```

## 7. Thin divider flourish

- **File**: `web/public/archive-assets/divider-flourish.svg`
- **Size**: scalable SVG, viewBox around `0 0 480 24`
- **Purpose**: hairline ornament under artwork and section titles.
- **Implementation**: created directly as SVG. Regenerate by editing the SVG paths, not with image generation.

## 8. Corner frame flourishes

- **File**: `web/public/archive-assets/corner-flourishes.svg`
- **Size**: scalable SVG, viewBox around `0 0 100 100`
- **Purpose**: subtle frame corners around portrait/card containers.
- **Implementation**: created directly as SVG. Regenerate by editing the SVG paths, not with image generation.

## 9. Generic skill icon placeholders, only if real skill icons are missing

- **File pattern**: `web/public/archive-assets/skill-placeholder-{type}.png`
- **Size**: `128x128`
- **Purpose**: use only when game skill icons fail to load.
- **Prompt**:

```text
Square fantasy game skill icon placeholder on clean white background, ornate small emblem on aged parchment, muted teal and sepia palette, painterly but readable at 48px, no text, no letters, no character face, softly framed rounded-square edge, cohesive with antique illustrated archive UI.
```

## Python cleanup note

Use Python/Pillow to convert white or ivory backgrounds into alpha before wiring PNGs into the UI. Keep generated source files in `web/public/assets/` if needed, and keep cleaned runtime files in `web/public/archive-assets/`.

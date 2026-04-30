#!/usr/bin/env python3
from pathlib import Path
from PIL import Image, ImageDraw, ImageFilter

ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / "cmd" / "proxy-node-studio-wails" / "app.ico"
FAVICON = ROOT / "cmd" / "proxy-node-studio-wails" / "web" / "favicon.ico"

SIZE = 256
img = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
draw = ImageDraw.Draw(img)

# Base rounded square matching the custom titlebar logo.
base_box = (14, 14, SIZE - 14, SIZE - 14)
draw.rounded_rectangle(base_box, radius=64, fill=(11, 22, 52, 255))

# Vertical background gradient.
gradient = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
g = ImageDraw.Draw(gradient)
for i in range(SIZE):
    blend = i / (SIZE - 1)
    color = (
        int(15 + 20 * blend),
        int(30 + 58 * blend),
        int(70 + 98 * blend),
        255,
    )
    g.line([(0, i), (SIZE, i)], fill=color)
mask = Image.new("L", (SIZE, SIZE), 0)
ImageDraw.Draw(mask).rounded_rectangle(base_box, radius=64, fill=255)
img = Image.composite(gradient, img, mask)
draw = ImageDraw.Draw(img)

# Inner glow panel.
inner_box = (28, 28, SIZE - 28, SIZE - 28)
inner = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
i = ImageDraw.Draw(inner)
i.rounded_rectangle(inner_box, radius=52, fill=(0, 0, 0, 0))
i.rounded_rectangle(inner_box, radius=52, fill=(0, 0, 0, 0), outline=(160, 220, 255, 50), width=2)
core_grad = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
cg = ImageDraw.Draw(core_grad)
for r in range(92, 12, -1):
    alpha = int(190 * (r / 92))
    cg.ellipse((SIZE/2-r, SIZE/2-r, SIZE/2+r, SIZE/2+r), fill=(34, 255, 165, alpha // 3))
for r in range(82, 8, -1):
    alpha = int(220 * (r / 82))
    cg.ellipse((SIZE/2-r, SIZE/2-r, SIZE/2+r, SIZE/2+r), fill=(0, 229, 255, alpha // 3))
for r in range(74, 6, -1):
    alpha = int(225 * (r / 74))
    cg.ellipse((SIZE/2-r, SIZE/2-r, SIZE/2+r, SIZE/2+r), fill=(155, 92, 255, alpha // 4))
core_grad = core_grad.filter(ImageFilter.GaussianBlur(8))
img.alpha_composite(core_grad)

# Orbital ring.
ring = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
r = ImageDraw.Draw(ring)
r.ellipse((58, 70, 198, 186), outline=(236, 251, 255, 176), width=12)
ring = ring.rotate(-24, resample=Image.Resampling.BICUBIC, center=(SIZE / 2, SIZE / 2))
img.alpha_composite(ring)

# Center light.
center_glow = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
cd = ImageDraw.Draw(center_glow)
for radius, alpha in [(30, 70), (22, 120), (14, 180)]:
    cd.ellipse((SIZE/2-radius, SIZE/2-radius, SIZE/2+radius, SIZE/2+radius), fill=(236, 251, 255, alpha))
center_glow = center_glow.filter(ImageFilter.GaussianBlur(4))
img.alpha_composite(center_glow)
draw = ImageDraw.Draw(img)
draw.ellipse((118, 118, 138, 138), fill=(236, 251, 255, 255))

# Outer stroke.
draw.rounded_rectangle(base_box, radius=64, outline=(135, 230, 255, 120), width=4)

OUT.parent.mkdir(parents=True, exist_ok=True)
FAVICON.parent.mkdir(parents=True, exist_ok=True)
img.save(OUT, format="ICO", sizes=[(256, 256), (128, 128), (64, 64), (48, 48), (32, 32), (16, 16)])
img.save(FAVICON, format="ICO", sizes=[(64, 64), (48, 48), (32, 32), (16, 16)])
print(OUT)
print(FAVICON)

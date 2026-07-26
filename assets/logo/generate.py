"""Mímir logo - over/under knot rendering, final museum-grade."""
import math
from PIL import Image, ImageDraw, ImageFont

W, H = 2400, 2400
BG = (6, 10, 20)
COPPER = (210, 130, 60)
COPPER_H = (240, 178, 100)
COPPER_D = (130, 75, 30)
TEAL = (50, 220, 200)
TEAL_D = (28, 125, 112)
WHITE = (240, 244, 248)
GRAY_L = (140, 150, 168)
GRAY = (100, 112, 130)
DARK = (18, 24, 38)

FONT_DIR = r"C:\Users\David\.config\opencode\skills\canvas-design\canvas-fonts"

def font(name, size):
    try:
        return ImageFont.truetype(f"{FONT_DIR}\\{name}", size)
    except:
        return ImageFont.load_default()

def lerp(a, b, t):
    return int(a + (b - a) * t)

def lerp_c(c1, c2, t):
    return tuple(lerp(a, b, t) for a, b in zip(c1, c2))

def smooth(pts, passes=4):
    """Chaikin subdivision for smooth curves."""
    for _ in range(passes):
        n = len(pts)
        out = []
        for i in range(n):
            p0, p1 = pts[i], pts[(i+1) % n]
            out.append((p0[0]*0.75+p1[0]*0.25, p0[1]*0.75+p1[1]*0.25))
            out.append((p0[0]*0.25+p1[0]*0.75, p0[1]*0.25+p1[1]*0.75))
        pts = out
    return pts

def strand_points(cx, cy, r, offset, N=100):
    """Generate smooth control points for one knot strand."""
    raw = []
    for i in range(N):
        a = offset + 2 * math.pi * i / N
        wave = math.sin(a * 4) * r * 0.22
        cr = r + wave
        raw.append((cx + cr * math.cos(a), cy + cr * math.sin(a)))
    return smooth(raw, 4)

def draw_strand(draw, pts, color, width):
    for i in range(len(pts) - 1):
        draw.line([pts[i], pts[i+1]], fill=color, width=width)

def find_crossings(pts1, pts2, threshold=25):
    """Find where two strands are close enough to cross."""
    crossings = []
    for i in range(0, len(pts1), 8):
        for j in range(0, len(pts2), 8):
            d = math.hypot(pts1[i][0]-pts2[j][0], pts1[i][1]-pts2[j][1])
            if d < threshold:
                crossings.append((i, j, pts1[i]))
    return crossings

def draw_norse_knot(draw, cx, cy, r):
    """Two-strand Norse knot with proper over/under crossings."""
    pts1 = strand_points(cx, cy, r, 0)
    pts2 = strand_points(cx, cy, r, math.pi)

    # Draw the full strand 2 first (background strand)
    draw_strand(draw, pts2, COPPER_D, 12)

    # Draw strand 1 in segments, with gaps where it goes "under" strand 2
    # At even crossings: strand 1 goes UNDER (gap in strand 1)
    # At odd crossings: strand 1 goes OVER (drawn on top)
    crossings = find_crossings(pts1, pts2, 30)

    if not crossings:
        # Fallback: just draw both
        draw_strand(draw, pts1, COPPER, 14)
    else:
        # Build segments of strand 1, alternating over/under
        crossing_indices = sorted(set(c[0] for c in crossings))
        # Group crossings into over/under pairs
        segments = []
        prev = 0
        for ci, idx in enumerate(crossing_indices):
            if ci % 2 == 0:
                # UNDER: draw from prev to idx (this segment is behind)
                segments.append(('under', prev, idx))
            else:
                # OVER: draw from prev to idx (this segment is on top)
                segments.append(('over', prev, idx))
            prev = idx
        segments.append(('under' if len(crossing_indices) % 2 == 0 else 'over', prev, len(pts1)-1))

        # Draw under segments first (darker), then over segments (brighter)
        for layer in ['under', 'over']:
            color = COPPER_D if layer == 'under' else COPPER
            w = 10 if layer == 'under' else 14
            for seg_type, s, e in segments:
                if seg_type == layer and e > s:
                    draw_strand(draw, pts1[s:e+1], color, w)

    # Highlight on strand 1 (top surface)
    draw_strand(draw, pts1, COPPER_H, 3)

    # Crossing markers
    for i in range(8):
        a = 2 * math.pi * i / 8
        wave = math.sin(a * 4) * r * 0.22
        cr = r + wave
        x = cx + cr * math.cos(a)
        y = cy + cr * math.sin(a)
        s = 14
        draw.polygon([(x, y-s), (x+s, y), (x, y+s), (x-s, y)], fill=COPPER_H)

def draw_neural_ring(draw, cx, cy, r, count=12):
    nodes = []
    for i in range(count):
        a = 2 * math.pi * i / count - math.pi / 2
        j = math.sin(a * 7 + i * 1.3) * r * 0.03
        nodes.append((cx + (r+j)*math.cos(a), cy + (r+j)*math.sin(a)))
    for i in range(count):
        draw.line([nodes[i], nodes[(i+1)%count]], fill=TEAL_D, width=2)
        if i % 3 != 0:
            draw.line([nodes[i], nodes[(i+2)%count]], fill=TEAL_D, width=1)
    for x, y in nodes:
        draw.ellipse([x-8, y-8, x+8, y+8], outline=TEAL, width=2)
        draw.ellipse([x-3, y-3, x+3, y+3], fill=TEAL)

def draw_memory_rings(draw, cx, cy):
    for r, w, c in [(220, 2, TEAL_D), (178, 2, COPPER_D), (140, 2, TEAL_D), (108, 1, COPPER_D)]:
        draw.ellipse([cx-r, cy-r, cx+r, cy+r], outline=c, width=w)

def draw_core(draw, cx, cy):
    for r in range(96, 0, -1):
        t = r / 96
        c = lerp_c(COPPER_D, COPPER, (t-0.55)/0.45) if t > 0.55 else lerp_c(COPPER, COPPER_H, t/0.55)
        draw.ellipse([cx-r, cy-r, cx+r, cy+r], fill=c)
    draw.ellipse([cx-36, cy-36, cx+36, cy+36], fill=TEAL)
    draw.ellipse([cx-17, cy-17, cx+17, cy+17], fill=BG)
    draw.ellipse([cx-6, cy-6, cx+6, cy+6], fill=TEAL)
    draw.ellipse([cx-7, cy-11, cx, cy-4], fill=(255, 255, 255))

def draw_spokes(draw, cx, cy, r_in, r_out, count=48):
    for i in range(count):
        a = 2 * math.pi * i / count
        draw.line([(cx+r_in*math.cos(a), cy+r_in*math.sin(a)),
                    (cx+r_out*math.cos(a), cy+r_out*math.sin(a))], fill=DARK, width=1)

def main():
    img = Image.new("RGBA", (W, H), (*BG, 255))
    draw = ImageDraw.Draw(img, "RGBA")
    cx, cy = W // 2, H // 2 - 140

    for r in range(900, 0, -3):
        t = 1 - r / 900
        draw.ellipse([cx-r, cy-r, cx+r, cy+r], fill=(lerp(6,14,t), lerp(10,12,t), lerp(20,28,t), 255))

    draw_spokes(draw, cx, cy, 300, 480, 48)
    draw_neural_ring(draw, cx, cy, 365, 12)
    draw_norse_knot(draw, cx, cy, 250)
    draw_memory_rings(draw, cx, cy)
    draw_core(draw, cx, cy)

    ty = cy + 395
    f_t = font("GeistMono-Bold.ttf", 175)
    bb = draw.textbbox((0,0), "MIMIR", font=f_t)
    draw.text((cx-(bb[2]-bb[0])//2, ty), "MIMIR", fill=WHITE, font=f_t)
    draw.line([(cx-150, ty+200), (cx+150, ty+200)], fill=COPPER, width=3)
    f_s = font("WorkSans-Regular.ttf", 54)
    bb2 = draw.textbbox((0,0), "the rememberer", font=f_s)
    draw.text((cx-(bb2[2]-bb2[0])//2, ty+220), "the rememberer", fill=GRAY_L, font=f_s)
    f_tag = font("DMMono-Regular.ttf", 24)
    bb3 = draw.textbbox((0,0), "SELF-LEARNING AI AGENT", font=f_tag)
    draw.text((cx-(bb3[2]-bb3[0])//2, ty+305), "SELF-LEARNING AI AGENT", fill=GRAY, font=f_tag)

    cl, m = 45, 90
    for x, y, dx, dy in [(m,m,1,1),(W-m,m,-1,1),(m,H-m,1,-1),(W-m,H-m,-1,-1)]:
        draw.line([(x,y),(x+cl*dx,y)], fill=COPPER_D, width=2)
        draw.line([(x,y),(x,y+cl*dy)], fill=COPPER_D, width=2)

    out = r"E:\BMAD\assets\logo\mimir-logo.png"
    img.convert("RGB").save(out, "PNG")
    print(f"Saved: {out}")

if __name__ == "__main__":
    main()

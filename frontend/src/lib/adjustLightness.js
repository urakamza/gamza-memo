export function adjustLightness(hex, amount, mode = 'default', saturation = 0) {
    const r = parseInt(hex.slice(1, 3), 16) / 255;
    const g = parseInt(hex.slice(3, 5), 16) / 255;
    const b = parseInt(hex.slice(5, 7), 16) / 255;

    const luminance = (0.299 * r + 0.587 * g + 0.114 * b);

    if (mode === 'clamp') {
        if (luminance <= 0.5) {
            // 원본 RGB를 그대로 반환
            const toInt = (x) => Math.round(x * 255);
            return `${toInt(r)}, ${toInt(g)}, ${toInt(b)}`;
        }
    } else if (mode === 'invert') {
        // 밝은 색 → 어둡게, 어두운 색 → 밝게
        amount = luminance > 0.5 ? -Math.abs(amount) : Math.abs(amount);
    }

    // RGB → HSL
    const max = Math.max(r, g, b), min = Math.min(r, g, b);
    let h, s, l = (max + min) / 2;

    if (max === min) {
        h = s = 0;
    } else {
        const d = max - min;
        s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
        switch (max) {
            case r: h = ((g - b) / d + (g < b ? 6 : 0)) / 6; break;
            case g: h = ((b - r) / d + 2) / 6; break;
            case b: h = ((r - g) / d + 4) / 6; break;
        }
    }

    l = Math.max(0, Math.min(1, l + amount));
    s = Math.max(0, Math.min(1, s + saturation));

    // HSL → RGB → HEX
    const hue2rgb = (p, q, t) => {
        if (t < 0) t += 1;
        if (t > 1) t -= 1;
        if (t < 1/6) return p + (q - p) * 6 * t;
        if (t < 1/2) return q;
        if (t < 2/3) return p + (q - p) * (2/3 - t) * 6;
        return p;
    };

    let r2, g2, b2;
    if (s === 0) {
        r2 = g2 = b2 = l;
    } else {
        const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
        const p = 2 * l - q;
        r2 = hue2rgb(p, q, h + 1/3);
        g2 = hue2rgb(p, q, h);
        b2 = hue2rgb(p, q, h - 1/3);
    }

    const toInt = (x) => Math.round(x * 255);
    return `${toInt(r2)}, ${toInt(g2)}, ${toInt(b2)}`;
}

export function getRgb(hex) {
    if (!hex) return;
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `${r},${g},${b}`;
}

export function getTextColor(hex, darkOrLight=false) {
    if (!hex || hex.length < 7) {
        return darkOrLight ? 'light' : '20,20,20';
    }
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
    if (darkOrLight) return luminance > 0.5 ? 'light' : 'dark';
    return luminance > 0.5 ? '20,20,20' : '240,240,240';
}
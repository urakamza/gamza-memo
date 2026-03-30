export function stretchToCss(stretch) {
    const map = {
        1: 'ultra-condensed',
        2: 'extra-condensed',
        3: 'condensed',
        4: 'semi-condensed',
        5: 'normal',
        6: 'semi-expanded',
        7: 'expanded',
        8: 'extra-expanded',
        9: 'ultra-expanded',
    };
    return map[stretch] ?? 'normal';
}
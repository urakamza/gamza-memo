export function applyTheme(theme) {
    const root = document.documentElement

    if (theme === 'system') {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
        root.setAttribute('data-theme', prefersDark ? 'dark' : 'light')
    } else {
        root.setAttribute('data-theme', theme)
    }
}

export function watchSystemTheme(currentTheme) {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    mq.addEventListener('change', () => {
        if (currentTheme() === 'system') {
            applyTheme('system')
        }
    })
}
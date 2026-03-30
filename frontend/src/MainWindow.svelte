<script>
    import { onMount } from 'svelte';
    import { NoteService } from '../bindings/changeme';
    import { Browser, Events, Window } from '@wailsio/runtime';
    import Icon from './Icon.svelte';
    import { applyTheme, watchSystemTheme } from './lib/theme';
    import { adjustLightness, getTextColor } from './lib/adjustLightness';
    import ConfirmDialog from './ConfirmDialog.svelte';
    import { stretchToCss } from './lib/stretchToCss';

    let startupEnabled = false;
    let notes = [];
    let systemFonts = [];
    let isDark = false;
    let confirmDelete = true;
    let showDeleteConfirm = false;
    let deleteTargetId = null;
    let searchQuery = '';

    $: fontFamilies = [...new Set(systemFonts.map(f => f.family))].sort((a, b) => a.localeCompare(b, 'ko'));
    $: fontWeights = systemFonts
        .filter(f => f.family === settings.fontFamily && !f.italic)
        .map(f => ({ weight: f.weight, name: f.name, stretch: f.stretch }))
        .sort((a, b) => {
            if (a.stretch === 5 && b.stretch !== 5) return -1;
            if (a.stretch !== 5 && b.stretch === 5) return 1;
            if (a.stretch === b.stretch) return a.weight - b.weight;
            return a.weight - b.weight || a.stretch - b.stretch;
        });

    $: filteredNotes = notes.filter(note => {
        if (!searchQuery.trim()) return true;
        const text = note.content.replace(/<[^>]*>/g, '').toLowerCase();
        return text.includes(searchQuery.toLowerCase());
    });

    let showSettings = false;
    let settings = { startupEnabled: false, fontFamily: '', fontSize: 14, fontWeight: 400, fontStretch: 5 };

    onMount(async () => {
        const result = await NoteService.GetNotes();
        notes = result ?? [];
        startupEnabled = await NoteService.GetStartupEnabled();
        settings = await NoteService.GetSettings();
        applyTheme(settings.theme);
        watchSystemTheme(() => settings.theme);
        isDark = settings.theme === 'dark' || (settings.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
        confirmDelete = settings.confirmDelete;

        const fonts = await NoteService.GetSystemFonts();
        systemFonts = fonts ?? [];

        // systemFonts 로드 후 현재 설정값 동기화
        if (settings.fontFamily && settings.fontWeight) {
            const weights = systemFonts
                .filter(f => f.family === settings.fontFamily && !f.italic)
                .map(f => ({ weight: f.weight, name: f.name, stretch: f.stretch }));
            
            console.log('저장된 fontWeight:', settings.fontWeight);
            console.log('저장된 fontStretch:', settings.fontStretch);
            console.log('weights 목록:', weights);
            console.log('fontStretch 타입:', typeof settings.fontStretch, settings.fontStretch);
            console.log('select value:', `${settings.fontWeight}|${settings.fontStretch ?? 5}`);
            const current = weights.find(f =>
                f.weight === Number(settings.fontWeight) && 
                f.stretch === Number(settings.fontStretch ?? 5)
            );
            console.log('current 찾음:', current);
        }

        Events.On('notes:updated', (e) => {
            notes = e.data ?? [];
        });
        Events.On('settings:updated', (e) => {
            applyTheme(e.data.theme);
            isDark = e.data.theme === 'dark' || (e.data.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
            confirmDelete = e.data.confirmDelete;
        });
    });

    async function updateSettings() {
        await NoteService.UpdateSettings(settings);
    }
</script>

<div class="container">
    <div class="header">
        {#if !showSettings}
        <button class="new-btn" on:click={() => NoteService.CreateNote()} title="새 메모"><Icon name="add" /></button>
        {:else}
        <button class="back-btn" on:click={() => showSettings = false} title="뒤로"><Icon name="back" /></button>
        {/if}
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <div class="drag-region" on:dblclick={async () => await Window.IsMaximised() ? Window.UnMaximise() : Window.Maximise()}>
            <h1>{showSettings ? '설정' : '스티커 메모'}</h1>
        </div>
        <div class="right">
            {#if !showSettings}<button class="set-btn" on:click={() => showSettings = true} title="설정"><Icon name="setting" /></button>{/if}
            <button class="close-btn" on:click={() => Window.Close()} title="닫기"><Icon name="close" /></button>
        </div>
    </div>

    <div class="settings-panel" class:show={showSettings}>
        <div class="settings-container">
            <div class="settings-content">
                <div class="setting-item">
                    <h2>일반</h2>
                    <h3>시작 시 자동 실행</h3>
                    <span class="checkbox">
                        <input
                            type="checkbox"
                            bind:checked={settings.startupEnabled}
                            on:change={async () => {
                                await NoteService.SetStartupEnabled(settings.startupEnabled);
                                await updateSettings();
                            }}
                        />
                        {settings.startupEnabled ? '켬' : '끔'}
                    </span>
                    <h3>삭제하기 전에 물어보기</h3>
                    <span class="checkbox">
                        <input
                            type="checkbox"
                            bind:checked={settings.confirmDelete}
                            on:change={async () => { await updateSettings(); }}
                        />
                        {settings.confirmDelete ? '켬' : '끔'}
                    </span>
                </div>

                <div class="setting-item">
                    <h2>테마</h2>
                    <div class="theme-radio">
                        <input type="radio" name="theme" id="t-light" value="light" bind:group={settings.theme} on:change={updateSettings} />
                        <label for="t-light">밝게</label>
                    </div>
                    <div class="theme-radio">
                        <input type="radio" name="theme" id="t-dark" value="dark" bind:group={settings.theme} on:change={updateSettings} />
                        <label for="t-dark">어둡게</label>
                    </div>
                    <div class="theme-radio">
                        <input type="radio" name="theme" id="t-system" value="system" bind:group={settings.theme} on:change={updateSettings} />
                        <label for="t-system">시스템 색상 모드 사용</label>
                    </div>
                </div>

                <div class="setting-item">
                    <h2>글꼴 설정</h2>
                    <h3>글꼴</h3>
                    <div class="font-selector">
                        <select bind:value={settings.fontFamily} on:change={() => {
                            const weights = systemFonts
                                .filter(f => f.family === settings.fontFamily && !f.italic)
                                .sort((a, b) => a.stretch - b.stretch || a.weight - b.weight);
                            if (weights.length > 0) {
                                const normal = weights.find(w => w.weight === 400 && w.stretch === 5)
                                    ?? weights.find(w => w.stretch === 5)
                                    ?? weights[0];
                                settings.fontWeight = normal.weight;
                                settings.fontStretch = normal.stretch;
                            }
                            updateSettings();
                        }}>
                            <option value="'Malgun Gothic', 'Apple SD Gothic Neo', 'Noto Sans KR', sans-serif">시스템 기본</option>
                            <option disabled>────────────────────</option>
                            {#each fontFamilies as family}
                                <option value={family} style="font-family: '{family}'">{family}</option>
                            {/each}
                        </select>
                        {#if fontWeights.length > 0}
                        <select
                            value="{settings.fontWeight}|{settings.fontStretch ?? 5}"
                            on:change={(e) => {
                                const [w, s] = e.target.value.split('|').map(Number);
                                settings.fontWeight = w;
                                settings.fontStretch = s;
                                updateSettings();
                            }}
                        >
                            {#each fontWeights as f}
                                <option value="{f.weight}|{f.stretch}">{f.name}</option>
                            {/each}
                        </select>
                        {/if}
                    </div>

                    <h3>글꼴 크기<span>(px)</span></h3>
                    <input type="number" min="6" max="32" bind:value={settings.fontSize} on:change={updateSettings} />
                    
                    <h3>자간<span>(em)</span></h3>
                    <input type="number" min="-0.3" max="0.3" step="0.01" bind:value={settings.letterSpacing} on:change={updateSettings} />
                    <h3>행간<span>(%)</span></h3>
                    <input type="number" min="50" max="600" step="10" bind:value={settings.lineHeight} on:change={updateSettings} />
                </div>

                <div class="setting-item">
                    <h2>이미지 최대 너비<span>(px)</span></h2>
                    <input type="number" bind:value={settings.imgWidth} on:change={updateSettings} step="10" min="0" max="9999" />
                    <p class="info">0 = 100%(기본값)</p>
                </div>

                <div class="setting-item">
                    <h2>프로그램 정보</h2>
                    <p class="verinfo">감자 메모 Ver 1.0.1</p>
                    <a href="https://github.com/urakamza" on:click|preventDefault={(e) => Browser.OpenURL(e.currentTarget.href)}>GitHub</a>
                    <a href="https://urakamza.kr" on:click|preventDefault={(e) => Browser.OpenURL(e.currentTarget.href)}>개발자 홈페이지</a>
                </div>
            </div>
        </div>
    </div>

    <div class="note-list-container">
        <div class="note-list">
            <div class="note-list-content">
                <div class="search-bar">
                    <Icon name="search" />
                    <input type="text" placeholder="메모 검색..." bind:value={searchQuery} />
                    {#if searchQuery}
                    <button class="clear-btn" on:click={() => searchQuery = ''}><Icon name="close" /></button>
                    {/if}
                </div>
                {#each filteredNotes as note}
                    <div
                        class="note-item"
                        role="button"
                        tabindex="0"
                        on:click={() => NoteService.OpenNoteWindow(note.id)}
                        on:keydown={(e) => { if(e.key === 'Enter') NoteService.OpenNoteWindow(note.id) }}
                    >
                        <div class="left-bar" style="background-color: {isDark ? `rgb(${adjustLightness(note.color, -0.2, 'clamp')})` : `rgb(${adjustLightness(note.color, 0.08, 'invert')})`}"></div>
                        <p class="note-preview">
                            {note.content ? note.content.replace(/<[^>]*>/g, '').slice(0, 100) || '빈 메모' : '빈 메모'}
                        </p>
                        <div class="note-btns">
                            <button class="delete-btn"
                                on:click|stopPropagation={async () => {
                                    if (confirmDelete) {
                                        deleteTargetId = note.id;
                                        showDeleteConfirm = true;
                                        return;
                                    }
                                    await NoteService.DeleteNote(note.id);
                                }}
                            >
                                <Icon name="bin" />
                            </button>
                            <button
                                class="note-status"
                                class:open={note.isOpen}
                                on:click|stopPropagation={async () => {
                                    if (note.isOpen) await NoteService.CloseNote(note.id);
                                    else await NoteService.OpenNoteWindow(note.id);
                                }}
                            >{note.isOpen ? '열림' : '닫힘'}</button>
                            <span class="note-status-text" class:open={note.isOpen}>{note.isOpen ? '열림' : '닫힘'}</span>
                        </div>
                    </div>
                {:else}
                    <div class="empty">메모가 없어요</div>
                {/each}
            </div>
        </div>
    </div>

    {#if showDeleteConfirm}
    <ConfirmDialog
        message="메모를 삭제할까요?"
        on:confirm={async () => {
            await NoteService.DeleteNote(deleteTargetId);
            showDeleteConfirm = false;
            deleteTargetId = null;
        }}
        on:cancel={() => {
            showDeleteConfirm = false;
            deleteTargetId = null;
        }}
    />
    {/if}
</div>

<style>
/* ===== 리셋 ===== */
* { margin: 0; padding: 0; box-sizing: border-box; user-select: none; }

:global(body) {
    font-family: 'Malgun Gothic', 'Apple SD Gothic Neo', 'Noto Sans KR', sans-serif;
    font-size: 16px;
    background-color: rgb(var(--bg));
    color: rgb(var(--text));
    transition: background-color 0.2s ease;
}

/* ===== 레이아웃 ===== */
.container {
    display: flex;
    flex-direction: column;
    align-items: center;
    height: 100vh;
}

/* ===== 헤더 ===== */
.header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    height: 41px;
    padding: 4px;
    background-color: rgb(var(--title-bar));
    border-bottom: 1px solid rgb(var(--border));
    color: rgb(var(--text));
    -webkit-app-region: drag;
    --wails-draggable: drag;
}

.drag-region {
    display: flex;
    align-items: center;
    padding-left: 8px;
    flex: 1;
    height: 100%;
    font-size: 18px;
    font-weight: bold;
    -webkit-app-region: drag;
    --wails-draggable: drag;
}

.header .right { display: flex; gap: 4px; height: 100%; }

.new-btn, .set-btn, .close-btn, .back-btn {
    width: 32px;
    height: 32px;
    border: none;
    border-radius: 6px;
    background-color: transparent;
    color: rgb(var(--text));
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background-color 0.2s;
    --wails-draggable: no-drag;
    -webkit-app-region: no-drag;
}
.new-btn, .back-btn { background: rgba(var(--text), 0.1); }
.new-btn:hover, .set-btn:hover, .close-btn:hover, .back-btn:hover { background: rgba(var(--text), 0.2); }
.new-btn:active, .set-btn:active, .close-btn:active, .back-btn:active { background: rgba(0,0,0,0.4); }

/* ===== 설정 패널 ===== */
.settings-panel {
    position: fixed;
    inset: 41px 0 0 0;
    z-index: 100;
    background-color: rgb(var(--bg));
    opacity: 0;
    transform: scale(0.9);
    pointer-events: none;
    transition: transform 0.2s ease, opacity 0.2s ease, background-color 0.2s ease;
    padding: 0 4px 16px;
}
.settings-panel.show {
    opacity: 1;
    transform: scale(1);
    pointer-events: auto;
}

.settings-container {
    display: flex;
    justify-content: center;
    overflow-y: auto;
    height: 100%;
    padding: 16px 8px 0;
}

.settings-content {
    width: 100%;
    max-width: 420px;
}
.settings-content::after {
    content: '';
    display: block;
    height: 32px;
}

/* ===== 설정 아이템 ===== */
.setting-item {
    display: flex;
    flex-direction: column;
    padding: 16px 0;
    border-bottom: 1px solid rgb(var(--border));
    color: rgb(var(--text));
}

h1 { font-size: 16px; font-weight: bold; line-height: 1; }
h2 { font-size: 21px; font-weight: bold; line-height: 1; margin-bottom: 16px; }
h2 span, h3 span { font-size: 14px; color: rgba(var(--text), 0.6); }
h3 { font-size: 18px; font-weight: bold; line-height: 1; margin: 8px 0 12px; color: rgba(var(--text), 0.8); }
h3:nth-child(n+3) { margin-top: 16px; }

p.info { font-size: 14px; color: rgba(var(--text), 0.4); margin-top: -8px; }
.setting-item a { color: rgb(var(--link)); }
.checkbox { display: flex; align-items: center; gap: 8px; }

.font-selector {
    display: grid;
    grid-template-columns: calc(60% - 2px) calc(40% - 2px);
    gap: 4px;
    margin-bottom: 12px;
}

.verinfo {
    margin-bottom: 16px;
}

/* ===== 폼 요소 ===== */
.setting-item select,
.setting-item input[type="number"] {
    height: 36px;
    padding: 6px;
    font-size: 16px;
    border: 1px solid rgb(var(--border));
    border-radius: 4px;
    background: rgb(var(--input));
    color: rgb(var(--text));
    outline: none;
    margin-bottom: 12px;
}
.font-selector select {
    height: 32px;
    padding: 2px;
    font-size: 12px;
    width: 100%;
    margin-bottom: 0;
}
.setting-item input[type="number"] { width: 100px; }

:global(html[data-theme="dark"]) .setting-item input[type="number"]::-webkit-inner-spin-button { filter: invert(1); }
:global(html[data-theme="dark"]) .setting-item input[type="number"]:hover,
:global(html[data-theme="dark"]) .setting-item input[type="number"]:active,
:global(html[data-theme="dark"]) .setting-item input[type="number"]:focus { background-color: #000; }

input[type="checkbox"] {
    appearance: none;
    position: relative;
    width: 48px;
    height: 24px;
    border: 2px solid rgb(var(--text));
    border-radius: 16px;
    background-color: rgb(var(--bg));
    transition: background-color 0.2s ease, border 0.2s ease;
}
input[type="checkbox"]::before {
    content: '';
    position: absolute;
    width: 16px;
    height: 16px;
    top: 2px;
    left: 2px;
    border-radius: 50%;
    background-color: rgb(var(--text));
    transition: left 0.2s ease;
}
input[type="checkbox"]:checked { border-color: transparent; background-color: rgb(var(--main-color)); }
input[type="checkbox"]:checked::before { left: 26px; background-color: #fff; }

.theme-radio { display: flex; align-items: center; padding-bottom: 8px; }
.theme-radio label { padding-left: 8px; }

input[type="radio"] {
    appearance: none;
    width: 18px;
    height: 18px;
    border: 2px solid rgba(var(--text), 0.4);
    border-radius: 50%;
    background-color: rgb(var(--bg));
    display: flex;
    align-items: center;
    justify-content: center;
}
input[type="radio"]:hover { border-color: rgba(var(--text), 0.6); }
input[type="radio"]:checked { border-color: rgb(var(--main-color)); }
input[type="radio"]:checked::before {
    content: '';
    display: block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: rgb(var(--text));
}

/* ===== 노트 목록 ===== */
.note-list-container {
    position: relative;
    width: 100%;
    height: calc(100% - 41px);
    padding: 0 4px 16px;
}

.note-list {
    display: flex;
    justify-content: center;
    overflow-y: auto;
    width: 100%;
    height: 100%;
}

.note-list-content {
    display: flex;
    flex-direction: column;
    width: 100%;
    max-width: 360px;
    gap: 8px;
    padding: 16px 8px 0;
}
.note-list-content::after {
    content: '';
    display: block;
    height: 12px;
    flex-shrink: 0;
}

/* ===== 검색바 ===== */
.search-bar {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 8px;
    margin-bottom: 8px;
    border-bottom: 1px solid rgb(var(--border));
    color: rgb(var(--text));
    flex-shrink: 0;
}
.search-bar input {
    flex: 1;
    height: 24px;
    border: none;
    background: transparent;
    outline: none;
    font-size: 13px;
    color: rgb(var(--text));
}
.search-bar input::placeholder { color: rgba(var(--text), 0.4); }

.clear-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: rgb(var(--text));
    cursor: pointer;
}
.clear-btn:hover { color: rgba(var(--text), 0.8); }

/* ===== 노트 아이템 ===== */
.note-item {
    position: relative;
    border-radius: 6px;
    border: 1px solid rgb(var(--border));
    overflow: hidden;
    flex-shrink: 0;
    cursor: pointer;
    transition: background-color 0.2s;
}
.note-item:hover { background: rgba(var(--text), 0.05); }
.note-item:active { background: rgba(var(--text), 0.1); }

.note-item .left-bar {
    position: absolute;
    top: 0;
    left: 0;
    width: 6px;
    height: 100%;
}

.note-preview {
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    overflow: hidden;
    white-space: normal;
    font-size: 14px;
    padding: 8px 10px 0 14px;
    color: rgb(var(--text));
}
:global(html[data-theme="dark"]) .note-preview { color: #fff; }

.note-btns {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    height: 24px;
    padding: 0 8px;
    background-color: rgba(var(--text), 0.06);
    margin-top: 10px;
}

/* ===== 삭제 버튼 ===== */
.delete-btn {
    width: 24px;
    height: 24px;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: rgb(var(--text));
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.1s, color 0.1s;
}
.note-item:hover .delete-btn { opacity: 1; }
.delete-btn:hover { color: rgb(var(--text)) !important; background: rgba(var(--delete), 1); }

/* ===== 노트 상태 스위치 ===== */
.note-status {
    width: 26px;
    height: 16px;
    border-radius: 16px;
    border: 1px solid rgb(var(--text));
    margin: 0 4px 0 8px;
    background: rgb(var(--bg));
    cursor: pointer;
    display: flex;
    align-items: center;
    overflow: hidden;
    text-indent: -9999px;
    font-size: 8px;
    transition: background-color 0.2s, border-color 0.2s;
}
.note-status::before {
    content: '';
    display: block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background-color: rgb(var(--text));
    margin-left: 2px;
    flex-shrink: 0;
    transition: margin-left 0.2s ease;
}
.note-status.open { background-color: rgb(var(--btn-active)); border-color: transparent; }
.note-status.open::before { background-color: #fff; margin-left: 12px; box-shadow: 0 0 2px 0.5px rgba(0,0,0,0.5); }

.note-status-text { font-size: 12px; line-height: 12px; color: rgba(var(--text), 0.4); }
.note-status-text.open { color: rgba(var(--text), 1); }

.empty { text-align: center; padding: 32px 0; color: rgb(var(--text)); }

/* ===== 스크롤바 ===== */
.note-list::-webkit-scrollbar,
.settings-container::-webkit-scrollbar { width: 4px; }

.note-list::-webkit-scrollbar-track,
.settings-container::-webkit-scrollbar-track { background: transparent; }

.note-list::-webkit-scrollbar-thumb,
.settings-container::-webkit-scrollbar-thumb { background: rgba(var(--text), 0.3); border-radius: 4px; }

.note-list::-webkit-scrollbar-thumb:hover,
.settings-container::-webkit-scrollbar-thumb:hover { background: rgba(var(--text), 0.7); }

:global(html[data-theme="dark"]) .note-list::-webkit-scrollbar-thumb,
:global(html[data-theme="dark"]) .settings-container::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.3); }

:global(html[data-theme="dark"]) .note-list::-webkit-scrollbar-thumb:hover,
:global(html[data-theme="dark"]) .settings-container::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,0.7); }
</style>
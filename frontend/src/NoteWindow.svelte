<script>
    import { onMount, tick } from "svelte";
    import { NoteService } from "../bindings/changeme";
    import { Events, Window } from "@wailsio/runtime";
    import { OpenNoteWindow } from "../bindings/changeme/noteservice";
    import { Editor } from "@tiptap/core";
    import StarterKit from "@tiptap/starter-kit";
    import Image from "@tiptap/extension-image";
    import Icon from "./Icon.svelte";
    import { applyTheme, watchSystemTheme } from "./lib/theme";
    import { adjustLightness, getRgb, getTextColor } from "./lib/adjustLightness";
    import ConfirmDialog from "./ConfirmDialog.svelte";
    import { stretchToCss } from "./lib/stretchToCss";

    const params = new URLSearchParams(window.location.search);
    const noteId = params.get('noteId');
    const PALETTE = ["#fff2c2", "#fee1ea", "#e5ffdb", "#e0f2ff", "#efdcfe", "#ededed", "#787878"];

    let editor;
    let editorEl;
    let note = null;
    let fileInput;
    let colorInput;
    let isDraggingInternal = false;
    let menuBtnEl;
    let editorFocused = false;
    let isDark = false;
    let confirmDelete = true;
    let showDeleteConfirm = false;

    let isComposing = false;
    let blurPending = false;

    let isBulletList = false;
    let isOrderedList = false;
    let isBold = false;
    let isItalic = false;
    let isUnderline = false;
    let isStrike = false;

    // 추가 메뉴
    let menuOpen = false;
    let menuEl;

    function updateActiveStates() {
        isBulletList = editor.isActive('bulletList');
        isOrderedList = editor.isActive('orderedList');
        isBold = editor.isActive('bold');
        isItalic = editor.isActive('italic');
        isStrike = editor.isActive('strike');
        isUnderline = editor.isActive('underline');
    }

    function applyColorVars(color) {
        document.documentElement.style.setProperty('--note-text', getTextColor(color));
        document.documentElement.style.setProperty('--note-color', getRgb(color));
        document.documentElement.style.setProperty('--note-color-dark', adjustLightness(color, -0.08));
        document.documentElement.style.setProperty('--note-color-darker', adjustLightness(color, -0.4, 'invert', -0.4));
        document.documentElement.style.setProperty('--note-color-scroll', adjustLightness(color, -0.2, 'invert', -0.45));
        document.documentElement.style.setProperty('--note-color-clamp', adjustLightness(color, -0.2, 'clamp'));
        document.documentElement.style.setProperty('--note-color-invert', adjustLightness(color, 0.08, 'invert'));
    }

    function setColor(color) {
        if (note.color === color) return;
        note.color = color;
        applyColorVars(color);
    }

    async function changeColor(color) {
        setColor(color);
        await NoteService.UpdateColor(note);
    }

    function toggleMenu() {
        menuOpen = !menuOpen;
    }

    function handleWindowMousedown(e) {
        if (!menuOpen) return;
        if (menuEl && !menuEl.contains(e.target) && menuBtnEl && !menuBtnEl.contains(e.target)) {
            menuOpen = false;
        }
    }

    function handleInput() {
        NoteService.UpdateNote(note);
    }

    async function handleImageFile(file) {
        if (!file || !file.type.startsWith('image/')) return;
        const buffer = await file.arrayBuffer();
        const bytes = Array.from(new Uint8Array(buffer));
        const path = await NoteService.SaveImage(noteId, bytes, file.name);
        const imgHtml = `<img src="${path}"><p></p>`;
        editor.chain().focus().insertContent(imgHtml).run();
    }

    async function handleImageUpload(e) {
        const files = Array.from(e.target.files);
        const paths = [];
        for (const file of files) {
            const buffer = await file.arrayBuffer();
            const bytes = Array.from(new Uint8Array(buffer));
            const path = await NoteService.SaveImage(noteId, bytes, file.name);
            paths.push(path);
        }
        const imgHtml = paths.map(path => `<img src="${path}">`).join('<p></p>');
        editor.chain().focus().insertContent(imgHtml).run();
    }
    

    async function syncFromDom(fromIME = false) {
        const proseMirror = editorEl.querySelector('.ProseMirror');
        if (!proseMirror) return;
        
        const clone = proseMirror.cloneNode(true);
        clone.querySelectorAll('br.ProseMirror-trailingBreak').forEach(br => br.remove());
        const domHtml = clone.innerHTML ?? '';
        
        if (!domHtml) return;

        if (!fromIME) {
            // 일반 blur: 공백 방지용 - 같으면 스킵
            if (domHtml === note.content) return;
        }
        
        note.content = domHtml;
        await NoteService.SaveNoteNow(note);
        const saved = await NoteService.GetNoteById(noteId);
        if (saved.content !== editor.getHTML()) {
            editor.commands.setContent(saved.content, false);
            note.content = saved.content;
        }
    }

    const InlineImage = Image.extend({
        inline: true,
        group: 'inline',
        selectable: false,
    });

    onMount(async () => {
        const prevent = (e) => {
            e.preventDefault();
            e.stopPropagation();
            if (e.dataTransfer) {
                e.dataTransfer.dropEffect = 'copy';
            }
        }
        window.addEventListener("dragover", prevent);
        window.addEventListener("drop", prevent);
        window.addEventListener("mousedown", handleWindowMousedown);

        const settings = await NoteService.GetSettings();
        applySettings(settings);
        applyTheme(settings.theme);
        isDark = settings.theme === 'dark' || (settings.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
        watchSystemTheme(() => settings.theme);
        confirmDelete = settings.confirmDelete;

        Events.On('settings:updated', (e) => {
            applySettings(e.data);
            applyTheme(e.data.theme);
            isDark = e.data.theme === 'dark' || (e.data.theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
            confirmDelete = e.data.confirmDelete;
        });
        Events.On('common:WindowLostFocus', () => {
            menuOpen = false;
        });

        note = await NoteService.GetNoteById(noteId);
        applyColorVars(note.color);

        await tick();
        editorEl.addEventListener('dragstart', () => { isDraggingInternal = true });
        editorEl.addEventListener('dragend', () => { isDraggingInternal = false });
        editorEl.addEventListener('compositionstart', () => { isComposing = true; });
        editorEl.addEventListener('compositionend', async () => {
            isComposing = false;
            if (!blurPending) return;
            blurPending = false;
            await syncFromDom(true);
        });
        editor = new Editor({
            element: editorEl,
            extensions: [StarterKit.configure({link: false}), InlineImage],
            content: note.content,
            editorProps: {
                attributes: {
                    spellcheck: 'false',
                },
                handleDoubleClick(view, pos, event) {
                    const target = event.target;
                    if (target.tagName === 'IMG') {
                        const src = target.getAttribute('src');
                        if (src) NoteService.OpenImageViewer(noteId, encodeURIComponent(src));
                        return true;
                    }
                    return false;
                },
                handlePaste(view, event) {
                    const items = event.clipboardData?.items;
                    if (!items) return false;
                    let images = 0;
                    for (const item of items) {
                        if (item.type.startsWith('image/')) {
                            handleImageFile(item.getAsFile());
                            images++;
                        }
                    }
                    if (images <= 0) return false;
                    return true;
                },
                handleDrop(view, event) {
                    if (isDraggingInternal) return false;
                    const types = Array.from(event.dataTransfer?.types || []);
                    if (!types.includes('Files')) return false;
                    const files = Array.from(event.dataTransfer?.files || []);
                    if (!files.length) return false;
                    event.preventDefault();
                    for (const file of files) {
                        if (file.type.startsWith('image/')) {
                            handleImageFile(file);
                        }
                    }
                    return true;
                }
            },
            onUpdate: ({ editor }) => {
                note.content = editor.getHTML();
                handleInput();
                updateActiveStates();
            },
            onSelectionUpdate() {
                updateActiveStates();
            },
            onFocus() { editorFocused = true; },
            onBlur() {
                editorFocused = false;
                if (isComposing) {
                    blurPending = true;
                    return;
                }
                setTimeout(() => syncFromDom(false), 50);
            },
        });

        return () => {
            window.removeEventListener("dragover", prevent);
            window.removeEventListener("drop", prevent);
            window.removeEventListener("mousedown", handleWindowMousedown);
            Events.Off('common:WindowLostFocus');
        }
    });

    function applySettings(settings) {
        const fontFamily = settings.fontFamily
            .split(',')
            .map(f => {
                f = f.trim();
                if (f.startsWith("'") || f.startsWith('"')) return f;
                return `'${f}'`;
            })
            .join(', ');
        document.documentElement.style.setProperty('--img-width', settings.imgWidth === 0 ? '100%' : `${settings.imgWidth}px`);
        document.documentElement.style.setProperty('--font-family', fontFamily);
        document.documentElement.style.setProperty('--font-weight', settings.fontWeight ?? 400);
        document.documentElement.style.setProperty('--font-style', settings.fontItalic ? 'italic' : 'normal');
        document.documentElement.style.setProperty('--font-size', settings.fontSize + 'px');
        document.documentElement.style.setProperty('--font-stretch', stretchToCss(settings.fontStretch ?? 5));
        document.documentElement.style.setProperty('--line-height', settings.lineHeight + '%');
        document.documentElement.style.setProperty('--letter-spacing', settings.letterSpacing + 'em');
    }
</script>

<!-- HTML -->
<div class="app" data-wails-dropzone="true">

    <!-- 타이틀바 -->
    <div class="titlebar">
        <div class="title-bg" 
            style="background: {note && !PALETTE.includes(note.color) && isDark ? 'rgb(var(--note-color))' : ''}" 
        />
        <!-- 왼쪽: 추가 + 메뉴 -->
        <div class="left-controls">
            <button class="ctrl-btn" title="새 메모" on:click={() => NoteService.CreateNote()}>
                <Icon name="add" />
            </button>
        </div>

        <!-- 드래그 영역 -->
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <div class="drag-region" on:dblclick={async () => await Window.IsMaximised() ? Window.UnMaximise() : Window.Maximise()}></div>

        <!-- 오른쪽: 핀 + 닫기 -->
        <div class="right-controls">
            <button class="ctrl-btn" title="메뉴" bind:this={menuBtnEl} on:click={toggleMenu} class:open={menuOpen}>
                <Icon name="more" />
            </button>
            <button class="ctrl-btn" class:pinned={note?.alwaysOnTop} title="항상 위"
                on:click={async () => {
                    note.alwaysOnTop = !note.alwaysOnTop;
                    await NoteService.UpdatePin(note);
                    Window.SetAlwaysOnTop(note.alwaysOnTop);
                }}>
                {#if !note?.alwaysOnTop}<Icon name="pin" />{:else}<Icon name="pinned" />{/if}
            </button>
            <button class="ctrl-btn close" title="닫기" on:click={() => Window.Close()}>
                <Icon name="close" />
            </button>
        </div>
    </div>

    <!-- 슬라이딩 메뉴 -->
    <div class="slide-menu" class:open={menuOpen} bind:this={menuEl}>
        <div class="menu-inner">
            <!-- 컬러 팔레트 -->
            <div class="color-palette">
                {#each PALETTE as color}
                    <button
                        class="color-dot"
                        class:active={note?.color === color}
                        style="background: {color}; border-color: rgb({adjustLightness(color, -0.2)})"
                        on:click={async () => {
                            await changeColor(color);
                            menuOpen = false;
                        }}
                    />
                {/each}
                <!-- 커스텀 색상 -->
                <button
                    class="color-dot custom-dot"
                    class:active={note && !PALETTE.includes(note.color)}
                    style="background: {note && !PALETTE.includes(note.color) ? note.color : 'transparent'}; border-color: {note && !PALETTE.includes(note.color) ? `rgb(${adjustLightness(note.color, -0.2)})`: ``};"
                    title="커스텀 색상"
                    on:click={() => colorInput.click()}
                />
                <input
                    type="color"
                    bind:this={colorInput}
                    value={note?.color ?? '#ffe590'}
                    style="display:none"
                    on:change={async (e) => {
                        await changeColor(e.target.value);
                        menuOpen = false;
                    }}
                />
            </div>

            <!-- 메뉴 버튼들 -->
            <div class="menu-actions">
                <button class="menu-btn" on:click={() => { NoteService.OpenMainWindow(); menuOpen = false; }}>
                    <Icon name="list" />
                    <span>메모 목록</span>
                </button>
                <button class="menu-btn danger"
                    on:click={async () => {
                        if (confirmDelete) {
                            showDeleteConfirm = true;
                            return;
                        }
                        await NoteService.DeleteNote(noteId);
                        Window.Close();
                    }}
                >
                    <Icon name="bin" />
                    <span>메모 삭제</span>
                </button>
            </div>
        </div>
    </div>

    <!-- 에디터 -->
    {#if note}
    <div class="editor-wrapper" class:focused={editorFocused}>
        <div class="editor" bind:this={editorEl}></div>
    </div>
    {/if}

    <!-- 하단 툴바 -->
    {#if editor}
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="toolbar" class:visible={editorFocused} on:mousedown|preventDefault>
        <div class="tool-bg" />
        <button class="tool-btn" class:active={isBold} on:click={() => editor.chain().focus().toggleBold().run()} title="굵게"><Icon name="bold" /></button>
        <button class="tool-btn italic" class:active={isItalic} on:click={() => editor.chain().focus().toggleItalic().run()} title="기울임꼴"><Icon name="italic" /></button>
        <button class="tool-btn" class:active={isUnderline} on:click={() => editor.chain().focus().toggleUnderline().run()} title="밑줄"><Icon name="under" /></button>
        <button class="tool-btn" class:active={isStrike} on:click={() => editor.chain().focus().toggleStrike().run()} title="취소선"><Icon name="strike" /></button>
        <button class="tool-btn" class:active={isBulletList} on:click={() => editor.chain().focus().toggleBulletList().run()} title="글머리 기호"><Icon name="ul" /></button>
        <button class="tool-btn" class:active={isOrderedList} on:click={() => editor.chain().focus().toggleOrderedList().run()} title="번호 목록"><Icon name="ol" /></button>
        <div class="separator"></div>
        <button class="tool-btn" on:click={() => fileInput.click()} title="이미지 추가"><Icon name="image" /></button>
        <input type="file" accept="image/*" multiple style="display:none" bind:this={fileInput} on:change={handleImageUpload}>
    </div>
    {/if}
    {#if showDeleteConfirm}
    <ConfirmDialog
        message="메모를 삭제할까요?"
        on:confirm={async () => {
            await NoteService.DeleteNote(noteId);
            Window.Close();
        }}
        on:cancel={() => showDeleteConfirm = false}
    />
    {/if}
</div>


<style>

.ctrl-btn, .color-dot, .left-controls, .right-controls {
    --wails-draggable: no-drag;
    -webkit-app-region: no-drag;
}

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

:global(body) {
    overflow: hidden;
    background-color: rgb(var(--bg));
    color: rgb(var(--note-text));
}

.app {
    display: flex;
    flex-direction: column;
    height: 100dvh;
    background: rgba(var(--note-color),1);
    transition: background-color 0.2s ease;
}

:global(html[data-theme="dark"]) .app {
    background-color: rgba(var(--note-color),0);
}

/* 타이틀바 */
.titlebar {
    z-index: 4;
    position: relative;
    display: flex;
    align-items: center;
    height: 40px;
    padding: 0 6px;
    gap: 4px;
    flex-shrink: 0;
}
.title-bg, .tool-bg {
    position: absolute;
    z-index: -1;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    transition: background-color 0.2s ease;
    background: rgb(var(--note-color-invert));
    
}
:global(html[data-theme="dark"]) .title-bg {
    background: rgb(var(--note-color-clamp));
}

.left-controls,
.right-controls {
    display: flex;
    gap: 2px;
    align-items: center;
}

.drag-region {
    flex: 1;
    height: 100%;
    -webkit-app-region: drag;
    --wails-draggable: drag;
}

.ctrl-btn {
    width: 24px;
    height: 24px;
    border: none;
    background: transparent;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    color: rgba(var(--note-color-darker), 1);
    transition: background-color 0.1s, color 0.1s;
}

.ctrl-btn:hover {
    background: rgba(var(--note-text),0.1);
    color: rgba(var(--note-text), 1);
}

.ctrl-btn.pinned {
    background: rgba(var(--note-text),0.1);
    color: rgba(var(--note-text), 1);
    border: 2px solid rgba(var(--note-text),0.3);
}

.ctrl-btn.open {
    background: rgba(var(--note-text), 0.15);
    color: rgba(var(--note-text),1);
}

/* 슬라이딩 메뉴 */
.slide-menu {
    z-index: 2;
    position: fixed;
    --wails-draggable: no-drag;
    -webkit-app-region: no-drag;
    overflow: hidden;
    max-height: 0px;
    transition: max-height 0.2s ease, box-shadow 0.2s ease;
    color: rgb(var(--text));
    flex-shrink: 0;
    top: 40px;
    right: 16px;
    border-radius: 0 0 8px 8px;
    box-shadow: 0px 2px 4px 0px rgba(0,0,0,0);
    user-select: none;
}

.slide-menu.open {
    max-height: 150px;
    box-shadow: 0px 2px 4px 0px rgba(0,0,0,0.6);
}

.menu-inner {
    padding: 10px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    background: rgb(var(--menu));
    border-left: 1px solid rgb(var(--menu-border));
    border-right: 1px solid rgb(var(--menu-border));
    border-bottom: 1px solid rgb(var(--menu-border));
    border-radius: 0 0 8px 8px;
}

/* 컬러 팔레트 */
.color-palette {
    margin-top: 4px;
    display: flex;
    gap: 8px;
    align-items: center;
}

.color-dot {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    border: 2px solid transparent;
    cursor: pointer;
    transition: transform 0.1s, border-color 0.1s;
    flex-shrink: 0;
    outline: 2px solid rgba(0,0,0,0.3);
    outline-offset: -2px;
    box-shadow: rgba(255,255,255,1) 0px 0px 4px 0px;
}

.color-dot:hover {
    transform: scale(1.2);
}

.color-dot.active::before ,.custom-dot.active::before {
    display: block;
    content: '';
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background-color: rgb(var(--note-text));
}

.custom-dot {
    border: 1px dashed rgba(var(--text),0.35);
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    outline: none;
    box-shadow: none;
}

.custom-dot:not(.active)::after {
    content: '+';
    font-size: 14px;
    font-weight: bold;
    color: rgba(var(--text),0.6);
    line-height: 1;
}

.custom-dot.active {
    border-width: 2px;
    border-style: solid;
    outline: 2px solid rgba(0,0,0,0.3);
    box-shadow: rgba(255,255,255,1) 0px 0px 4px 0px;
}

/* 메뉴 액션 버튼 */
.menu-actions {
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 100%;
}

.menu-btn {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 10px 10px;
    border: none;
    border-radius: 5px;
    background: rgba(var(--text),0.1);
    color: var(--text);
    cursor: pointer;
    font-size: 16px;
    line-height: 1;
    transition: background-color 0.1s;
}

.menu-btn:hover {
    background: rgba(var(--text),0.15);
}

.menu-btn.danger:hover {
    background: rgb(var(--delete));
    color: rgb(var(--text));
}
@media (max-height: 200px) {
    .menu-actions {
        flex-direction: row;
    }
    .menu-btn {
        padding: 8px;
        max-width: 100%;
        font-size: 12px;
        flex: 1;
    }
}
@media (max-width: 260px) {
    .slide-menu {
        width: 100%;
        left: 0;
    }
}
@media (max-width: 220px) {
    .color-palette { gap: 4px; }
}
/* 에디터 */
.editor-wrapper {
    flex: 1;
    height: 0;
    overflow-y: auto;
    -webkit-app-region: no-drag;
    --wails-draggable: no-drag;
    padding-top: 4px;
    margin: 0px 4px 8px 4px;
    transition: margin 0.2s ease;
}
.editor-wrapper.focused {
    margin-bottom: 0;
}
.editor-wrapper::-webkit-scrollbar {
    width: 4px;
}

.editor-wrapper::-webkit-scrollbar-track {
    background: transparent;
}
.editor-wrapper::-webkit-scrollbar-button:vertical:start,
.editor-wrapper::-webkit-scrollbar-button:vertical:end {
    background: transparent;
    height: 4px;
}

.editor-wrapper::-webkit-scrollbar-thumb {
    background: rgba(var(--note-color-scroll), 0.7);
    filter: opacity(0.2);
    border-radius: 4px;
}

.editor-wrapper::-webkit-scrollbar-thumb:hover {
    background: rgb(var(--note-text), 0.7);
}
:global(html[data-theme="dark"]) .editor-wrapper::-webkit-scrollbar-thumb {
    background: rgba(255,255,255,0.2);
}

:global(html[data-theme="dark"]) .editor-wrapper::-webkit-scrollbar-thumb:hover {
    background: rgba(255,255,255,0.7);
}
.editor {
    height: calc(100% - 32px);
}
:global(html[data-theme="dark"]) :global(.ProseMirror) {
    color: #ffffff;
}
:global(.ProseMirror) {
    font-family: var(--font-family);
    font-size: var(--font-size, 14px);
    font-weight: var(--font-weight, 400);
    font-style: var(--font-style, normal);
    font-stretch: var(--font-stretch, normal);
    line-height: var(--line-height, 160%);
    letter-spacing: var(--letter-spacing, 0);
    min-height: 100%;
    height: 100%;
    padding: 0 8px;
    outline: none;
    color: rgb(var(--note-text, "34,34,34"));
}

:global(.ProseMirror p) {
    margin-bottom: 0.15em;
}

:global(.ProseMirror ul, .ProseMirror ol) {
    margin-left: 1em;
    margin-bottom: 0.4em;
}

:global(.ProseMirror img) {
    width: 100%;
    max-width: var(--img-width);
    border-radius: 4px;
    display: inline-block;
}

:global(.ProseMirror img:hover) {
    outline: 2px solid rgba(var(--note-color-darker), 0.4);
}
:global(.ProseMirror img:active) {
    outline: 2px solid rgba(var(--note-color-darker), 0.7);
}

/* 하단 툴바 */
.toolbar {
    position:relative;
    isolation: isolate;
    display: flex;
    align-items: center;
    overflow: hidden;
    height: 0;
    padding: 0 6px;
    gap: 2px;
    transition: height 0.15s ease;
    flex-shrink: 0;
}
:global(html[data-theme="dark"]) .tool-bg {
    border-top: 1px solid rgb(var(--border));
    background: transparent;
}

.toolbar.visible {
    height: 40px;
}

.tool-btn {
    width: 24px;
    height: 24px;
    border: none;
    background: transparent;
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
    color: rgba(var(--note-text), 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.1s, color 0.1s;
}
.tool-btn:hover, .tool-btn.active {
    background: rgba(var(--text),0.1);
    color: rgba(var(--note-text), 1);
}

.separator {
    width: 1px;
    height: 16px;
    background: rgba(var(--note-text), 0.2);
    margin: 0 2px;
}

:global(html[data-theme="dark"]) .tool-btn {
    color: rgba(var(--text), 0.5);
}
:global(html[data-theme="dark"]) .tool-btn:hover,
:global(html[data-theme="dark"]) .tool-btn.active {
    color: rgba(var(--text), 1);
}
:global(html[data-theme="dark"]) .separator {
    background: rgba(var(--text), 0.2);
}

</style>
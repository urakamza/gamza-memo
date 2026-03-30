<script>
    import { Window } from '@wailsio/runtime';
    import { onMount } from 'svelte';
  import Icon from './Icon.svelte';

    export let src = '';

    const STEPS = [0.15, 0.333, 0.5, 0.666, 1, 1.5, 2, 3, 4, 6, 8, 12, 16];

    let scale = 1;
    let isDragging = false;
    let startX = 0, startY = 0;
    let translateX = 0, translateY = 0;
    let imgEl, containerEl;
    let imgWidth = 0, imgHeight = 0;

    function zoom(delta, cx, cy) {
        const prev = scale;
        scale = Math.min(16, Math.max(0.1, scale * (1 + delta)));
        const ratio = scale / prev;
        translateX = (cx - translateX) * (1 - ratio) + translateX;
        translateY = (cy - translateY) * (1 - ratio) + translateY;
    }

    function stepZoom(direction) {
        const list = direction > 0 ? STEPS : [...STEPS].reverse();
        const next = list.find(s => direction > 0 ? s > scale + 0.001 : s < scale - 0.001);
        const nextScale = next ?? (direction > 0 ? STEPS[STEPS.length - 1] : STEPS[0]);
        const ratio = nextScale / scale;
        scale = nextScale;
        translateX *= ratio;
        translateY *= ratio;
    }

    function handleWheel(e) {
        e.preventDefault();
        const rect = containerEl.getBoundingClientRect();
        const cx = e.clientX - rect.left - rect.width / 2;
        const cy = e.clientY - rect.top - rect.height / 2;
        zoom(e.deltaY > 0 ? -0.1 : 0.1, cx, cy);
    }

    function handleMousedown(e) {
        if (e.button !== 0) return;
        isDragging = true;
        startX = e.clientX - translateX;
        startY = e.clientY - translateY;
    }

    function handleMousemove(e) {
        if (!isDragging) return;
        translateX = e.clientX - startX;
        translateY = e.clientY - startY;
    }

    function handleMouseup() { isDragging = false; }

    function reset() {
        scale = 1;
        translateX = 0;
        translateY = 0;
    }

    onMount(() => {
        imgEl.onload = () => {
            imgWidth = imgEl.naturalWidth;
            imgHeight = imgEl.naturalHeight;
            scale = Math.min(
                containerEl.clientWidth / imgWidth,
                containerEl.clientHeight / imgHeight,
                1
            );
            translateX = 0;
            translateY = 0;
        };
    });
</script>

<div class="viewer">
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="toolbar">
        <button on:click={() => stepZoom(-1)} title="축소"><Icon name="minus" /></button>
        <span class="scale-label">{Math.round(scale * 100)}%</span>
        <button on:click={() => stepZoom(1)} title="확대"><Icon name="add" /></button>
        <button on:click={reset} title="원본 크기"><Icon name="orig" /></button>
        <div class="spacer" on:dblclick={async () => await Window.IsMaximised() ? Window.UnMaximise() : Window.Maximise()}></div>
        <button class="close-btn" on:click={() => Window.Close()} title="닫기"><Icon name="close" /></button>
    </div>

    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div
        class="container"
        bind:this={containerEl}
        on:wheel|preventDefault={handleWheel}
        on:mousedown={handleMousedown}
        on:mousemove={handleMousemove}
        on:mouseup={handleMouseup}
        on:mouseleave={handleMouseup}
        style="cursor: {isDragging ? 'grabbing' : 'grab'}"
    >
        <img
            bind:this={imgEl}
            src={decodeURIComponent(src)}
            alt="이미지"
            style="
                width: {imgWidth}px;
                height: {imgHeight}px;
                transform: translate({translateX}px, {translateY}px) scale({scale});
                transform-origin: center center;
            "
            draggable="false"
        />
    </div>
</div>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }

.viewer {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: #1a1a1a;
    color: #fff;
    user-select: none;
}

.toolbar {
    display: flex;
    align-items: center;
    gap: 4px;
    height: 40px;
    padding: 0 8px;
    background: #2a2a2a;
    border-bottom: 1px solid #444;
    flex-shrink: 0;
    -webkit-app-region: drag;
    --wails-draggable: drag;
}

.toolbar button {
    width: 28px;
    height: 28px;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: rgba(var(--dark-text), 0.4);
    font-size: 16px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.1s, color 0.1s;
    -webkit-app-region: no-drag;
    --wails-draggable: no-drag;
}

.toolbar button:hover {
    color: rgba(var(--dark-text), 1);
    background: rgba(255,255,255,0.1);
}

.scale-label {
    font-size: 13px;
    color: rgba(255,255,255,0.7);
    min-width: 48px;
    text-align: center;
    -webkit-app-region: no-drag;
    --wails-draggable: no-drag;
}

.spacer {
    flex: 1;
    height: 100%;
}

.close-btn:hover { background: rgba(220, 50, 50, 0.4) !important; }

.container {
    flex: 1;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 8px 8px 8px;
}

img {
    max-width: none;
    max-height: none;
    flex-shrink: 0;
    image-rendering: auto;
}
</style>
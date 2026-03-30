<script>
    import MainWindow from './MainWindow.svelte'
    import NoteWindow from './NoteWindow.svelte'
    import ImageViewer from './ImageViewer.svelte'

    const params = new URLSearchParams(window.location.search);
    const noteId = params.get('noteId');
    const isViewer = params.has('viewer');
    const viewerSrc = params.get('src');
</script>

{#if isViewer}
    <ImageViewer src={viewerSrc} />
{:else if noteId}
    <NoteWindow />
{:else}
    <MainWindow />
{/if}

<style>
:global(:root) {
    color-scheme: light dark;
    --light-text: 20, 20, 20;
    --dark-text: 240, 240, 240;
    --main-color: 44, 143, 235;
}

:global([data-theme='light']) {
    --bg: 255,255,255;
    --menu: 248, 248, 248;
    --menu-border: 221, 221, 221;
    --text: var(--light-text);
    --border: 177, 177, 177;
    --input: 255,255,255;
    --link: 38, 116, 206;
    --delete: 255, 152, 152;
    --title-bar: 243, 243, 243;
    --btn-active: 70, 211, 105;
}

:global([data-theme='dark']) {
    --bg: 48, 48, 48;
    --menu: 53, 53, 53;
    --menu-border: 101, 101, 101;
    --text: var(--dark-text);
    --border: 80, 80, 80;
    --input: 53,53,53;
    --link: 96, 170, 255;
    --delete: 151, 17, 17;
    --title-bar: 38, 38, 38;
    --btn-active: 7, 192, 53;
}
</style>
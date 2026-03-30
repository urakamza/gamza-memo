<script>
    import { createEventDispatcher } from 'svelte';
    
    export let message = '정말 삭제할까요?';
    export let confirmText = '삭제';
    export let cancelText = '취소';

    const dispatch = createEventDispatcher();
</script>

<div class="dialog-overlay">
    <div class="dialog">
        <p>{message}</p>
        <div class="dialog-actions">
            <button on:click={() => dispatch('cancel')}>{cancelText}</button>
            <button class="danger" on:click={() => dispatch('confirm')}>{confirmText}</button>
        </div>
    </div>
</div>

<style>
.dialog-overlay {
    font-family: 'Malgun Gothic', 'Apple SD Gothic Neo', 'Noto Sans KR', sans-serif;
    position: fixed;
    inset: 0;
    z-index: 999;
    background: rgba(0,0,0,0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    -webkit-app-region: no-drag;
    --wails-draggable: no-drag;
}

.dialog {
    background: rgb(var(--bg));
    border-radius: 8px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    font-size: 16px;
    color: rgb(var(--text));
    border: 1px solid rgba(var(--menu-border),1);
    box-shadow: 0 4px 20px rgba(0,0,0,0.2);
    min-width: 160px;
}

.dialog-actions {
    display: flex;
    gap: 6px;
    justify-content: flex-end;
}

.dialog-actions button {
    padding: 4px 12px;
    border-radius: 5px;
    border: none;
    cursor: pointer;
    font-size: 14px;
    background: rgba(var(--text),0.1);
    color: var(--text);
}

.dialog-actions button:hover {
    background: rgba(var(--text),0.2);
}

.dialog-actions button.danger {
    background: rgba(var(--delete),0.5);
    color: rgb(var(--text));
}

.dialog-actions button.danger:hover {
    background: rgba(var(--delete),0.7);
}
</style>
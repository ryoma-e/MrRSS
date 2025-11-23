<script setup>
import { store } from '../store.js';
import { computed, ref, onMounted, onBeforeUnmount } from 'vue';
import { BrowserOpenURL } from '../wailsjs/wailsjs/runtime/runtime.js';

const article = computed(() => store.articles.find(a => a.id === store.currentArticleId));
const showContent = ref(false); // Toggle between original webpage and RSS content

function close() {
    store.currentArticleId = null;
    showContent.value = false;
}

function toggleRead() {
    if (!article.value) return;
    const newState = !article.value.is_read;
    article.value.is_read = newState;
    fetch(`/api/articles/read?id=${article.value.id}&read=${newState}`, { method: 'POST' });
}

function toggleFavorite() {
    if (!article.value) return;
    const newState = !article.value.is_favorite;
    article.value.is_favorite = newState;
    fetch(`/api/articles/favorite?id=${article.value.id}`, { method: 'POST' });
}

function openOriginal() {
    if (article.value) BrowserOpenURL(article.value.url);
}

function toggleContentView() {
    showContent.value = !showContent.value;
}

// Listen for render content event from context menu
function handleRenderContent() {
    showContent.value = true;
}

onMounted(() => {
    window.addEventListener('render-article-content', handleRenderContent);
});

onBeforeUnmount(() => {
    window.removeEventListener('render-article-content', handleRenderContent);
});
</script>

<template>
    <main :class="['flex-1 bg-bg-primary flex flex-col h-full absolute w-full md:static md:w-auto z-30 transition-transform duration-300', article ? 'translate-x-0' : 'translate-x-full md:translate-x-0']">
        <div v-if="!article" class="hidden md:flex flex-col items-center justify-center h-full text-text-secondary text-center">
            <i class="ph ph-newspaper text-5xl mb-5 opacity-50"></i>
            <p>{{ store.i18n.t('selectArticle') }}</p>
        </div>

        <div v-else class="flex flex-col h-full bg-bg-primary">
            <div class="h-[50px] px-5 border-b border-border flex justify-between items-center bg-bg-primary shrink-0">
                <button @click="close" class="md:hidden flex items-center gap-2 text-text-secondary hover:text-text-primary">
                    <i class="ph ph-arrow-left"></i> {{ store.i18n.t('back') }}
                </button>
                <div class="flex gap-2 ml-auto">
                    <button @click="toggleContentView" class="action-btn" :title="showContent ? store.i18n.t('viewOriginal') : store.i18n.t('viewContent')">
                        <i :class="['ph', showContent ? 'ph-globe' : 'ph-article']"></i>
                    </button>
                    <button @click="toggleRead" class="action-btn" :title="article.is_read ? store.i18n.t('markAsUnread') : store.i18n.t('markAsRead')">
                        <i :class="['ph', article.is_read ? 'ph-envelope-open' : 'ph-envelope']"></i>
                    </button>
                    <button @click="toggleFavorite" :class="['action-btn', article.is_favorite ? 'text-yellow-400' : '']" :title="store.i18n.t('toggleFavorite')">
                        <i :class="['ph', article.is_favorite ? 'ph-star-fill' : 'ph-star']"></i>
                    </button>
                    <button @click="openOriginal" class="action-btn" :title="store.i18n.t('openInBrowser')">
                        <i class="ph ph-arrow-square-out"></i>
                    </button>
                </div>
            </div>
            
            <!-- Original webpage view -->
            <div v-if="!showContent" class="flex-1 bg-white w-full">
                <iframe :src="article.url" class="w-full h-full border-none" sandbox="allow-scripts allow-same-origin allow-popups"></iframe>
            </div>
            
            <!-- RSS content view -->
            <div v-else class="flex-1 overflow-y-auto bg-bg-primary p-6">
                <div class="max-w-3xl mx-auto">
                    <h1 class="text-3xl font-bold mb-4 text-text-primary">{{ article.title }}</h1>
                    <div class="text-sm text-text-secondary mb-6 flex items-center gap-4">
                        <span>{{ article.feed_title }}</span>
                        <span>•</span>
                        <span>{{ new Date(article.published_at).toLocaleDateString() }}</span>
                    </div>
                    <div v-if="article.content" class="prose prose-lg max-w-none text-text-primary" v-html="article.content"></div>
                    <div v-else class="text-center text-text-secondary py-8">
                        <i class="ph ph-article text-5xl mb-3 opacity-50"></i>
                        <p>{{ store.i18n.locale.value === 'zh' ? '此文章没有内容' : 'No content available for this article' }}</p>
                    </div>
                </div>
            </div>
        </div>
    </main>
</template>

<style scoped>
.action-btn {
    @apply text-xl cursor-pointer text-text-secondary p-1.5 rounded-md transition-colors hover:bg-bg-tertiary hover:text-text-primary;
}

/* Prose styling for article content */
.prose {
    color: var(--text-primary);
}
.prose :deep(h1), .prose :deep(h2), .prose :deep(h3), .prose :deep(h4), .prose :deep(h5), .prose :deep(h6) {
    color: var(--text-primary);
    font-weight: 600;
    margin-top: 1.5em;
    margin-bottom: 0.75em;
}
.prose :deep(p) {
    margin-bottom: 1em;
    line-height: 1.7;
}
.prose :deep(a) {
    color: var(--accent-color);
    text-decoration: underline;
}
.prose :deep(img) {
    max-width: 100%;
    height: auto;
    border-radius: 0.5rem;
    margin: 1.5em 0;
}
.prose :deep(pre) {
    background-color: var(--bg-secondary);
    padding: 1em;
    border-radius: 0.5rem;
    overflow-x: auto;
    margin: 1em 0;
}
.prose :deep(code) {
    background-color: var(--bg-secondary);
    padding: 0.2em 0.4em;
    border-radius: 0.25rem;
    font-size: 0.9em;
}
.prose :deep(blockquote) {
    border-left: 4px solid var(--accent-color);
    padding-left: 1em;
    margin: 1em 0;
    font-style: italic;
    color: var(--text-secondary);
}
.prose :deep(ul), .prose :deep(ol) {
    margin: 1em 0;
    padding-left: 2em;
}
.prose :deep(li) {
    margin-bottom: 0.5em;
}
</style>

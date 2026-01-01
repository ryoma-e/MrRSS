import { createApp } from 'vue';
import { createPinia } from 'pinia';
import PhosphorIcons from '@phosphor-icons/vue';
import i18n, { locale } from './i18n';
import './style.css';
import App from './App.vue';
import { useAppStore } from './stores/app';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(i18n);
app.use(PhosphorIcons);

// Initialize language setting before mounting
async function initializeApp() {
  try {
    const res = await fetch('/api/settings');
    const data = await res.json();
    if (data.language) {
      locale.value = data.language;
    }

    // Start FreshRSS status polling if enabled
    if (data.freshrss_enabled === 'true') {
      const appStore = useAppStore();
      await appStore.startFreshRSSStatusPolling();
    }
  } catch (e) {
    console.error('Error loading language setting:', e);
  }

  app.mount('#app');
}

// Initialize and mount
initializeApp();

import { onMounted, onUnmounted } from 'vue';

interface WindowState {
  x: number;
  y: number;
  width: number;
  height: number;
  maximized: boolean;
}

export function useWindowState() {
  let saveTimeout: NodeJS.Timeout | null = null;
  let isRestoringState = false;

  /**
   * Load and restore window state from database
   *
   * NOTE: Window state restoration is currently disabled because:
   * 1. MrRSS uses HTTP API instead of Wails bindings (-skipbindings flag)
   * 2. Wails runtime bindings are not available in this mode
   * 3. Window position/size is managed by Wails itself in main.go
   *
   * The window state is still saved for potential future use or
   * for reference, but not actively restored on startup.
   */
  async function restoreWindowState() {
    try {
      isRestoringState = true;

      const response = await fetch('/api/window/state');
      if (!response.ok) {
        console.warn('Failed to load window state');
        return;
      }

      const data = await response.json();
      console.log('Window state loaded (not restored):', data);

      // We don't actually restore the state because Wails bindings are not available
      // The window will use the default size/position defined in main.go
    } catch (error) {
      console.error('Error loading window state:', error);
    } finally {
      // Wait a bit before allowing saves
      setTimeout(() => {
        isRestoringState = false;
      }, 1000);
    }
  }

  /**
   * Save current window state to database
   *
   * NOTE: Window state saving is currently disabled because:
   * 1. MrRSS uses HTTP API instead of Wails bindings (-skipbindings flag)
   * 2. Wails runtime bindings (WindowGetPosition, WindowGetSize, etc.) are not available
   * 3. Window state is managed by Wails itself via runtime.WindowSetPosition/Size in main.go
   *
   * If window state persistence is needed in the future, it should be implemented
   * via backend Go code that can access the Wails runtime context.
   */
  async function saveWindowState() {
    // Don't save while we're restoring state
    if (isRestoringState) {
      return;
    }

    // Disabled: Cannot access Wails runtime bindings in -skipbindings mode
    // The code below would fail because WindowGetPosition, WindowGetSize, etc. are not available

    /*
    try {
      const state: WindowState = {
        x: window.screenX || 0,
        y: window.screenY || 0,
        width: window.innerWidth || 1024,
        height: window.innerHeight || 768,
        maximized: false,
      };

      await fetch('/api/window/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(state),
      });

      console.log('Window state saved:', state);
    } catch (error) {
      console.error('Error saving window state:', error);
    }
    */
  }

  /**
   * Debounced save to avoid excessive writes
   */
  function debouncedSave() {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
    }
    saveTimeout = setTimeout(saveWindowState, 500);
  }

  /**
   * Setup window event listeners
   */
  function setupListeners() {
    // Listen to window resize and move events
    // We use multiple approaches to catch window state changes:

    // 1. Browser resize event (fires when window size changes)
    const handleResize = () => {
      debouncedSave();
    };
    window.addEventListener('resize', handleResize);

    // 2. Visibility change (fires when window is minimized/maximized)
    const handleVisibilityChange = () => {
      if (!document.hidden) {
        debouncedSave();
      }
    };
    document.addEventListener('visibilitychange', handleVisibilityChange);

    // 3. Periodic check as fallback for position changes
    // (position changes don't trigger browser events)
    const checkInterval = setInterval(() => {
      debouncedSave();
    }, 2000); // Check every 2 seconds

    return () => {
      window.removeEventListener('resize', handleResize);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      clearInterval(checkInterval);
      if (saveTimeout) {
        clearTimeout(saveTimeout);
      }
    };
  }

  /**
   * Initialize window state management
   *
   * Currently minimal because window state persistence is disabled.
   * This function is kept for API compatibility and may be enhanced
   * in the future if window state management is re-implemented.
   */
  function init() {
    onMounted(async () => {
      // Just log that we're initialized
      console.log('Window state management initialized (persistence disabled)');

      // Load state for logging/reference only
      await restoreWindowState();
    });

    onUnmounted(() => {
      // Cleanup if needed
      if (saveTimeout) {
        clearTimeout(saveTimeout);
      }
    });
  }

  return {
    init,
    restoreWindowState,
    saveWindowState,
  };
}

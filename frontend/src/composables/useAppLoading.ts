import { ref } from 'vue'

/** Shared loading text for the app-level loading overlay.
 *  Set from the router guard to reflect the current auth step. */
export const appLoadingText = ref('')

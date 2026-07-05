import { useToast } from './useToast'

/**
 * Copy text to clipboard with fallback for non-HTTPS / older browsers.
 * Returns true on success, false on failure.
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch {
    // Fallback for non-HTTPS or older browsers
    try {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
      return true
    } catch {
      return false
    }
  }
}

/**
 * Convenience: copy + toast. Shows success or failure toast automatically.
 */
export async function copyWithToast(text: string, successMsg: string, failMsg: string): Promise<boolean> {
  const ok = await copyToClipboard(text)
  const toast = useToast()
  if (ok) {
    toast.success(successMsg)
  } else {
    toast.error(failMsg)
  }
  return ok
}

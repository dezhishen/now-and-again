import { ref, type Ref } from 'vue'

/** Global singleton for programmatic confirm dialogs. */
const modalRef: Ref<{ show: (msg: string) => Promise<boolean> } | null> = ref(null)

export function registerConfirmModal(modal: { show: (msg: string) => Promise<boolean> }) {
  modalRef.value = modal
}

/** Show a confirm dialog. Returns true if user confirmed, false if cancelled. */
export async function useConfirm(message: string): Promise<boolean> {
  if (!modalRef.value) {
    return window.confirm(message) // fallback if modal not registered
  }
  return modalRef.value.show(message)
}

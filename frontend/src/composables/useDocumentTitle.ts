import { watch, toRef, type MaybeRefOrGetter } from 'vue'

export function useDocumentTitle(title: MaybeRefOrGetter<string>) {
  watch(
    toRef(title),
    (t) => {
      if (typeof document !== 'undefined') {
        document.title = t ? `${t} - WhatsApp Reminder` : 'WhatsApp Reminder'
      }
    },
    { immediate: true }
  )
}

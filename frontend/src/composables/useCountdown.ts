import { ref, watch, toRef, toValue, onUnmounted, type MaybeRefOrGetter } from 'vue'

export function useCountdown(targetTime: MaybeRefOrGetter<string | Date | null | undefined>) {
  const formattedCountdown = ref('')
  const secondsLeft = ref(0)
  const isDue = ref(false)

  let timer: any = null

  const calculate = () => {
    const raw = toValue(targetTime)
    if (!raw) {
      formattedCountdown.value = ''
      secondsLeft.value = 0
      isDue.value = false
      return
    }

    const targetDate = typeof raw === 'string' ? new Date(raw) : raw
    const now = new Date()
    const diffMs = targetDate.getTime() - now.getTime()

    if (isNaN(diffMs)) {
      formattedCountdown.value = ''
      return
    }

    if (diffMs <= 0) {
      formattedCountdown.value = 'Waktu pengingat telah tiba'
      secondsLeft.value = 0
      isDue.value = true
      return
    }

    isDue.value = false
    const totalSecs = Math.floor(diffMs / 1000)
    secondsLeft.value = totalSecs

    const days = Math.floor(totalSecs / 86400)
    const hours = Math.floor((totalSecs % 86400) / 3600)
    const minutes = Math.floor((totalSecs % 3600) / 60)
    const seconds = totalSecs % 60

    if (days > 0) {
      formattedCountdown.value = `${days} hari ${hours} jam lagi`
    } else if (hours > 0) {
      formattedCountdown.value = `${hours} jam ${minutes} menit lagi`
    } else if (minutes > 0) {
      formattedCountdown.value = `${minutes} menit ${seconds} dtk lagi`
    } else {
      formattedCountdown.value = `${seconds} detik lagi`
    }
  }

  // Watch reactive target input
  watch(
    toRef(targetTime),
    () => {
      calculate()
    },
    { immediate: true }
  )

  timer = setInterval(calculate, 1000)

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })

  return {
    formattedCountdown,
    secondsLeft,
    isDue,
  }
}

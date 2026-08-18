<template>
  <div
    :class="
      twMerge(
        'latency-tag bg-base-100 border border-base-content/20 h-5 min-w-10 px-1.5 rounded-lg text-xs select-none shadow-xs font-mono font-bold inline-flex items-center justify-center',
      )
    "
    @mouseenter="handlerHistoryTip"
  >
    <span
      v-if="loading"
      class="loading loading-dots loading-xs text-base-content/70"
    ></span>
    <span
      v-else-if="latencyText"
      class="tabular-nums text-[11px] font-bold"
      :style="{ color: latencyColor }"
    >
      {{ latencyText }}
    </span>
    <BoltIcon
      v-else
      class="text-base-content/40 h-3 w-3"
    />
  </div>
</template>

<script setup lang="ts">
import { NOT_CONNECTED } from '@/constant'
import { useTooltip } from '@/helper/tooltip'
import { getHistoryByName, getLatencyByName } from '@/assembly/proxies'
import { BoltIcon } from '@heroicons/vue/24/outline'
import dayjs from 'dayjs'
import { twMerge } from 'tailwind-merge'
import { computed } from 'vue'

const props = defineProps<{
  name?: string
  loading?: boolean
  groupName?: string
}>()

const { showTip } = useTooltip()
const handlerHistoryTip = (e: Event) => {
  const history = getHistoryByName(props.name ?? '', props.groupName)

  if (!history.length) return

  const historyList = document.createElement('div')
  historyList.classList.add('flex', 'flex-col', 'gap-1')
  for (const item of history) {
    const itemDiv = document.createElement('div')
    const time = document.createElement('div')
    const latencyEl = document.createElement('div')

    time.textContent = dayjs(item.time).format('YYYY-MM-DD HH:mm:ss')
    latencyEl.textContent = item.delay + 'ms'
    if (item.delay < 200) latencyEl.style.color = '#22c55e'
    else if (item.delay < 500) latencyEl.style.color = '#eab308'
    else latencyEl.style.color = '#ef4444'

    itemDiv.classList.add('flex', 'items-center', 'gap-2')
    itemDiv.append(time, latencyEl)
    historyList.append(itemDiv)
  }

  showTip(e, historyList, {
    delay: [1000, 0],
    trigger: 'mouseenter',
    touch: false,
  })
}

const latency = computed(() => getLatencyByName(props.name ?? '', props.groupName))

const latencyText = computed(() => {
  if (props.loading) return ''
  const val = latency.value
  if (!val || val <= 0 || val === NOT_CONNECTED) return ''
  return `${val}`
})

const latencyColor = computed(() => {
  const val = latency.value
  if (!val || val <= 0 || val === NOT_CONNECTED) return '#9ca3af'
  if (val < 200) return '#22c55e'
  if (val < 500) return '#eab308'
  return '#ef4444'
})
</script>

<style scoped>
.latency-tag {
  transition: all 0.2s ease-out;
}
</style>

<template>
  <div class="base-container w-full p-4 select-none mb-3">
    <!-- Card Header -->
    <div class="flex items-center justify-between border-b border-base-content/10 pb-3 mb-3">
      <div class="flex items-center space-x-2">
        <span class="text-base-content/60 text-xs font-semibold tracking-wider uppercase">
          {{ $t('trafficAndRuleDistribution') }}
        </span>
      </div>
      <div class="text-[11px] font-mono text-base-content/50">
        {{ destData.length }} {{ $t('destinations') }}
      </div>
    </div>

    <!-- 2-Column Responsive Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- 1. Destination Traffic Distribution Pie -->
      <div class="p-3 zash-card-flat bg-base-200/50 flex flex-col justify-between space-y-2">
        <div class="flex items-center justify-between border-b border-base-content/10 pb-1.5">
          <span class="text-xs font-bold text-base-content flex items-center gap-1.5">
            <span class="w-2 h-2 rounded-full bg-success" />
            {{ $t('byDestination') }}
          </span>
          <span class="text-[10px] font-mono text-base-content/60 font-semibold">
            {{ formatBytes(totalDestBytes) }}
          </span>
        </div>

        <div class="h-44 w-full relative">
          <div ref="destChartRef" class="w-full h-full" />
          <div
            v-if="destData.length === 0"
            class="absolute inset-0 flex items-center justify-center text-base-content/40 text-xs italic"
          >
            {{ $t('noData') }}
          </div>
        </div>

        <!-- Top 3 Destination Items -->
        <div class="space-y-1.5 pt-1 text-[11px] font-mono">
          <div
            v-for="(item, idx) in destData.slice(0, 3)"
            :key="item.name"
            class="flex items-center justify-between"
          >
            <div class="flex items-center space-x-1.5 truncate max-w-[65%]">
              <span
                class="w-2 h-2 rounded-full flex-shrink-0"
                :style="{ backgroundColor: palette[(idx + 2) % palette.length] }"
              />
              <span class="truncate text-base-content/80 font-medium" :title="item.name">{{ item.name }}</span>
            </div>
            <span class="text-base-content font-bold tabular-nums">{{ formatBytes(item.value) }}</span>
          </div>
        </div>
      </div>

      <!-- 2. Rule Match / Hit Distribution Pie -->
      <div class="p-3 zash-card-flat bg-base-200/50 flex flex-col justify-between space-y-2">
        <div class="flex items-center justify-between border-b border-base-content/10 pb-1.5">
          <span class="text-xs font-bold text-base-content flex items-center gap-1.5">
            <span class="w-2 h-2 rounded-full bg-warning" />
            {{ $t('ruleHitDistribution') }}
          </span>
          <span class="text-[10px] font-mono text-base-content/60 font-semibold">
            {{ totalRuleHits }} hits
          </span>
        </div>

        <div class="h-44 w-full relative">
          <div ref="ruleChartRef" class="w-full h-full" />
          <div
            v-if="ruleData.length === 0"
            class="absolute inset-0 flex items-center justify-center text-base-content/40 text-xs italic"
          >
            {{ $t('noData') }}
          </div>
        </div>

        <!-- Top 3 Rule Items -->
        <div class="space-y-1.5 pt-1 text-[11px] font-mono">
          <div
            v-for="(item, idx) in ruleData.slice(0, 3)"
            :key="item.name"
            class="flex items-center justify-between"
          >
            <div class="flex items-center space-x-1.5 truncate max-w-[65%]">
              <span
                class="w-2 h-2 rounded-full flex-shrink-0"
                :style="{ backgroundColor: palette[(idx + 4) % palette.length] }"
              />
              <span class="truncate text-base-content/80 font-medium" :title="item.name">{{ item.name }}</span>
            </div>
            <span class="text-base-content font-bold tabular-nums">{{ item.value }} hits</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { Backend } from '@/utils/backend'
import { prettyBytesHelper as formatBytes } from '@/helper/utils'

interface PieItem {
  name: string
  value: number
}

const destChartRef = ref<HTMLDivElement | null>(null)
const ruleChartRef = ref<HTMLDivElement | null>(null)

let destChart: echarts.ECharts | null = null
let ruleChart: echarts.ECharts | null = null
let timer: number | null = null

const destData = ref<PieItem[]>([])
const ruleData = ref<PieItem[]>([])

const totalDestBytes = computed(() => destData.value.reduce((a, b) => a + b.value, 0))
const totalRuleHits = computed(() => ruleData.value.reduce((a, b) => a + b.value, 0))

const palette = [
  '#8583e9',
  '#6865c6',
  '#30d158',
  '#ff9f0a',
  '#64d2ff',
  '#ff453a',
  '#a855f7',
  '#ec4899',
  '#10b981',
  '#f59e0b',
]

async function fetchData() {
  try {
    const [analytics, rules] = await Promise.all([
      Backend.getBatchAnalytics('', 50),
      Backend.getRules(),
    ])

    if (analytics) {
      if (analytics.byDomain && analytics.byDomain.length > 0) {
        destData.value = analytics.byDomain.map((d) => ({
          name: d.name,
          value: d.totalBytes || d.uploadTotal + d.downloadTotal,
        }))
      }
    }

    if (rules && rules.length > 0) {
      ruleData.value = rules
        .filter((r) => r.hitCount > 0)
        .map((r) => ({
          name: r.payload || r.type,
          value: r.hitCount,
        }))

      if (ruleData.value.length === 0) {
        ruleData.value = rules.slice(0, 5).map((r, i) => ({
          name: r.payload || r.type,
          value: i === 0 ? 1 : 0,
        }))
      }
    }
  } catch {
    // ignore
  }
}

function buildOption(data: PieItem[], isBytes = true, colorOffset = 0): echarts.EChartsOption {
  const isDark = document.documentElement.classList.contains('dark') || document.body.getAttribute('data-theme') === 'dark'
  const colors = palette.slice(colorOffset).concat(palette.slice(0, colorOffset))

  return {
    animation: true,
    tooltip: {
      trigger: 'item',
      backgroundColor: isDark ? '#1c1c1e' : '#ffffff',
      borderColor: isDark ? '#2c2c2e' : '#e5e5ea',
      textStyle: {
        color: isDark ? '#f5f5f7' : '#1d1d1f',
        fontSize: 11,
        fontFamily: 'JetBrains Mono, monospace',
      },
      formatter: (params: any) => {
        const percent = params.percent || 0
        const formatted = isBytes ? formatBytes(params.value) : `${params.value} hits`
        return `<div class="font-sans font-bold text-xs mb-1">${params.name}</div>
          <div class="text-xs font-mono">
            Total: <strong>${formatted}</strong> (${percent}%)
          </div>`
      },
    },
    series: [
      {
        type: 'pie',
        radius: ['45%', '75%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 5,
          borderColor: isDark ? '#1c1c1e' : '#ffffff',
          borderWidth: 2,
        },
        label: {
          show: false,
        },
        emphasis: {
          scale: true,
          scaleSize: 6,
          label: {
            show: false,
          },
        },
        data: data.slice(0, 7).map((item, idx) => ({
          name: item.name,
          value: item.value,
          itemStyle: {
            color: colors[idx % colors.length],
          },
        })),
      },
    ],
  }
}

function updateCharts() {
  if (destChart && destChartRef.value) {
    destChart.setOption(buildOption(destData.value, true, 2))
  }
  if (ruleChart && ruleChartRef.value) {
    ruleChart.setOption(buildOption(ruleData.value, false, 4))
  }
}

function handleResize() {
  destChart?.resize()
  ruleChart?.resize()
}

onMounted(async () => {
  if (destChartRef.value) {
    destChart = echarts.init(destChartRef.value)
  }
  if (ruleChartRef.value) {
    ruleChart = echarts.init(ruleChartRef.value)
  }

  await fetchData()
  updateCharts()

  window.addEventListener('resize', handleResize)
  timer = window.setInterval(async () => {
    await fetchData()
    updateCharts()
  }, 2000)
})

onUnmounted(() => {
  clearInterval(timer!)
  window.removeEventListener('resize', handleResize)
  destChart?.dispose()
  ruleChart?.dispose()
})
</script>

<style scoped>
.zash-card-flat {
  border-radius: 0.75rem;
  border: 1px solid var(--color-base-content, rgba(0, 0, 0, 0.08));
}
</style>

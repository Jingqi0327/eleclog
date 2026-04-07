<template>
  <div class="room-detail-page surface-ground">
    <Card>
      <template #title>
        <div class="flex flex-column md:flex-row md:justify-content-between md:align-items-center gap-2">
          <div class="flex align-items-center gap-2">
            <Button icon="pi pi-arrow-left" text @click="$router.push('/')" />
            <span class="text-xl font-semibold">房间用电详情</span>
          </div>
          <span v-if="roomInfo" class="text-sm text-color-secondary">
            {{ roomInfo.building_code }}号楼 {{ roomInfo.floor_code }}层 {{ roomInfo.room_code }}室
          </span>
        </div>
      </template>

      <template #content>
        <div class="grid">
          <div class="col-12 md:col-4 flex">
            <div class="stat-card">
              <div class="stat-title">
                <i class="pi pi-building stat-icon"></i>
                房间名称
              </div>
              <div class="stat-value">{{ roomInfo?.name || '-' }}</div>
            </div>
          </div>
          <div class="col-12 md:col-4 flex">
            <div class="stat-card">
              <div class="stat-title">
                <i class="pi pi-wallet stat-icon"></i>
                当前余额
              </div>
              <div class="stat-value">{{ latestBalance !== null ? latestBalance.toFixed(2) : '-' }} 元</div>
            </div>
          </div>
          <div class="col-12 md:col-4 flex">
            <div class="stat-card">
              <div class="stat-title">
                <i class="pi pi-bolt stat-icon"></i>
                区间总用电
              </div>
              <div class="stat-value">{{ totalUsage.toFixed(2) }} 度</div>
            </div>
          </div>
        </div>

        <div class="filter-bar my-3">
          <!-- 左侧：快捷时间范围按钮 -->
          <div class="filter-left">
            <Button
              v-for="item in rangeOptions"
              :key="item.value"
              :label="item.label"
              size="small"
              :outlined="activeRange !== item.value"
              @click="setQuickRange(item.value)"
            />
          </div>
          <!-- 右侧：自定义日期范围 -->
          <div class="filter-right">
            <DatePicker
              v-model="customDates"
              selectionMode="range"
              placeholder="选择日期范围"
              dateFormat="yy-mm-dd"
              :maxDate="today"
              style="width: 240px"
            />
            <!-- 清除自定义日期的按钮 -->
            <Button
              icon="pi pi-times"
              size="small"
              severity="secondary"
              text
              :disabled="!customDates"
              @click="clearCustomDates"
            />
            <Button label="查询" icon="pi pi-search" size="small" @click="handleCustomSearch" />
          </div>
        </div>

        <div v-if="errorMsg" class="p-2 border-round bg-red-50 text-red-500 mb-3">
          {{ errorMsg }}
        </div>

        <div ref="chartRef" class="chart-wrap">
          <!-- 查询后无数据时显示提示，覆盖在图表区域上方 -->
          <div v-if="hasQueried && !loading && records.length === 0" class="no-data-overlay">
            暂无数据
          </div>
        </div>
      </template>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import { apiClient } from '@/client'
import * as echarts from 'echarts'
import dayjs, { Dayjs } from 'dayjs'
import Card from 'primevue/card'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'

interface ElectricityUsage {
  start_time: string
  end_time: string
  usage: number
  balance: number
}

interface RoomInfo {
  id: number
  name: string
  area_id: string
  building_code: string
  floor_code: string
  room_code: string
  created_at: string
}

// 对应后端 getElectricityBalanceResponse，balance 单位为「分」
interface LatestBalanceResponse {
  id: number
  room_id: number
  balance: number
  recorded_at: string
}
const route = useRoute()
const roomId = Number(route.params.id)

const records = ref<ElectricityUsage[]>([])
const roomInfo = ref<RoomInfo | null>(null)
const latestBalance = ref<number | null>(null) // 来自 /electricity-balances/latest，单位「元」
const activeRange = ref<string>('today')
const customDates = ref<Date[] | null>(null)
const chartRef = ref<HTMLElement | null>(null)
const loading = ref(false)
const errorMsg = ref('')
const hasQueried = ref(false) // 是否已发起过查询，用于区分「初始」和「查无数据」
const today = new Date() // 禁止选择未来日期

let myChart: echarts.ECharts | null = null


const totalUsage = computed(() => records.value.reduce((sum, row) => sum + row.usage, 0))

const rangeOptions = [
  { label: '今天', value: 'today' },
  { label: '昨天', value: 'yesterday' },
  { label: '本周', value: 'this_week' },
  { label: '上周', value: 'last_week' },
  { label: '本月', value: 'this_month' },
  { label: '上月', value: 'last_month' },
]

const resizeHandler = () => {
  myChart?.resize()
}

interface ChartPoint {
  label: string
  usage: number
  balance: number
  time: number
}

const round2 = (num: number) => Number(num.toFixed(2))

const formatBucketLabel = (timeMs: number, unit: 'hour' | 'day') => {
  return unit === 'day' ? dayjs(timeMs).format('MM-DD') : dayjs(timeMs).format('MM-DD HH:00')
}

const buildChartData = (rows: ElectricityUsage[]): ChartPoint[] => {
  if (rows.length === 0) {
    return []
  }

  const firstRow = rows[0]
  const lastRow = rows[rows.length - 1]
  if (!firstRow || !lastRow) {
    return []
  }

  const startMs = dayjs(firstRow.end_time).valueOf()
  const endMs = dayjs(lastRow.end_time).valueOf()
  const spanHours = Math.max((endMs - startMs) / (1000 * 60 * 60), 1)

  // 时间跨度大时直接按天聚合，否则按小时聚合。
  const baseUnit: 'hour' | 'day' = spanHours > 24 * 7 ? 'day' : 'hour'
  const bucketMap = new Map<number, ChartPoint>()

  for (const row of rows) {
    const t = dayjs(row.end_time)
    const bucketStart = baseUnit === 'day' ? t.startOf('day').valueOf() : t.startOf('hour').valueOf()
    const existed = bucketMap.get(bucketStart)

    if (!existed) {
      bucketMap.set(bucketStart, {
        label: formatBucketLabel(bucketStart, baseUnit),
        usage: row.usage,
        balance: row.balance,
        time: bucketStart,
      })
      continue
    }

    // 用电量是区间量，合并时累加；余额取该桶内最后一个点。
    existed.usage += row.usage
    existed.balance = row.balance
  }

  const merged = Array.from(bucketMap.values()).sort((a, b) => a.time - b.time)

  // 点数过大再做二次合并，避免图表过密。
  const maxPoints = 120
  if (merged.length <= maxPoints) {
    return merged.map((item) => ({ ...item, usage: round2(item.usage), balance: round2(item.balance) }))
  }

  const groupSize = Math.ceil(merged.length / maxPoints)
  const compact: ChartPoint[] = []

  for (let i = 0; i < merged.length; i += groupSize) {
    const group = merged.slice(i, i + groupSize)
    if (group.length === 0) {
      continue
    }
    const last = group[group.length - 1]
    if (!last) {
      continue
    }
    compact.push({
      label: last.label,
      usage: round2(group.reduce((sum, item) => sum + item.usage, 0)),
      balance: round2(last.balance),
      time: last.time,
    })
  }

  return compact
}

const initChart = () => {
  if (!chartRef.value) {
    return
  }
  myChart = echarts.init(chartRef.value)
  window.addEventListener('resize', resizeHandler)
}

const updateChart = () => {
  if (!myChart) {
    return
  }

  const chartData = buildChartData(records.value)

  // 计算余额的自适应纵轴范围：取区间内最小/最大值，向外延伸 30%，让曲线不贴边
  let balanceMin: number | undefined = undefined
  let balanceMax: number | undefined = undefined
  if (chartData.length > 0) {
    const balValues = chartData.map((p) => p.balance)
    const bMin = Math.min(...balValues)
    const bMax = Math.max(...balValues)
    const bRange = bMax - bMin || 10 // 若所有值相等则默认 10 的范围
    const bPad = Math.max(bRange * 0.3, 2) // 至少留 2 元的边距
    balanceMin = Math.floor(bMin - bPad)
    balanceMax = Math.ceil(bMax + bPad)
  }

  myChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['用电量(度)', '余额(元)'] },
    xAxis: {
      type: 'category',
      data: chartData.map((item) => item.label),
      axisLabel: { rotate: 25 },
    },
    yAxis: [
      { type: 'value', name: '用电量(度)', splitLine: { show: false } },
      { type: 'value', name: '余额(元)', splitLine: { show: false }, min: balanceMin, max: balanceMax },
    ],
    series: [
      {
        name: '用电量(度)',
        type: 'line',
        smooth: true,
        data: chartData.map((item) => item.usage),
      },
      {
        name: '余额(元)',
        type: 'line',
        yAxisIndex: 1,
        smooth: true,
        data: chartData.map((item) => item.balance),
      },
    ],
  })
}

const fetchRoomInfo = async () => {
  // 房间基本信息来自 GET /rooms/:id
  const response = await apiClient.get<RoomInfo>(`/rooms/${roomId}`)
  roomInfo.value = response.data
}

const fetchLatestBalance = async () => {
  // 当前最新余额来自 GET /electricity-balances/latest/:room_id
  // 后端返回的 balance 是 int64（分），除以 100 转为元，再保留两位小数
  const response = await apiClient.get<LatestBalanceResponse>(
    `/electricity-balances/latest/${roomId}`,
  )
  latestBalance.value = round2(response.data.balance / 100)
}

const loadUsageData = async (start: string, end: string) => {
  loading.value = true
  errorMsg.value = ''

  try {
    // 用电区间来自 GET /electricity-balances/hour-range/:room_id
    const response = await apiClient.get<ElectricityUsage[]>(
      `/electricity-balances/hour-range/${roomId}`,
      {
        params: {
          start_time: start,
          end_time: end,
        },
      },
    )

    // 后端返回后先标准化用电量精度，避免累计误差和展示抖动。
    records.value = response.data.map((item) => ({
      ...item,
      usage: round2(item.usage),
      balance: round2(item.balance),
    }))
    updateChart()
  } catch (error) {
    errorMsg.value = '加载用电明细失败，请检查后端服务或查询时间范围。'
    console.error(error)
  } finally {
    loading.value = false
    hasQueried.value = true // 无论成功还是失败，都标记为已查询过
  }
}

const setQuickRange = (type: string) => {
  activeRange.value = type

  let start: Dayjs = dayjs()
  let end: Dayjs = dayjs()

  if (type === 'today') {
    start = dayjs().startOf('day')
  } else if (type === 'yesterday') {
    start = dayjs().subtract(1, 'day').startOf('day')
    end = dayjs().subtract(1, 'day').endOf('day')
  } else if (type === 'this_week') {
    start = dayjs().startOf('week')
  } else if (type === 'last_week') {
    start = dayjs().subtract(1, 'week').startOf('week')
    end = dayjs().subtract(1, 'week').endOf('week')
  } else if (type === 'this_month') {
    start = dayjs().startOf('month')
  } else if (type === 'last_month') {
    start = dayjs().subtract(1, 'month').startOf('month')
    end = dayjs().subtract(1, 'month').endOf('month')
  }

  loadUsageData(start.toISOString(), end.toISOString())
}

const clearCustomDates = () => {
  customDates.value = null
  // 若当前是自定义状态，清除后恢复为无选中
  if (activeRange.value === 'custom') {
    activeRange.value = ''
  }
}

const handleCustomSearch = () => {
  if (!customDates.value || customDates.value.length !== 2 || !customDates.value[1]) {
    errorMsg.value = '请先选择完整的开始和结束时间。'
    return
  }

  activeRange.value = 'custom'
  // 自定义范围统一按整天处理：开始 00:00，结束 23:59:59。
  loadUsageData(
    dayjs(customDates.value[0]).startOf('day').toISOString(),
    dayjs(customDates.value[1]).endOf('day').toISOString(),
  )
}

onMounted(async () => {
  if (!Number.isInteger(roomId) || roomId < 1) {
    errorMsg.value = '房间 ID 无效。'
    return
  }

  initChart()

  try {
    await fetchRoomInfo()
    await fetchLatestBalance()
  } catch (error) {
    errorMsg.value = '加载房间信息失败，请检查房间是否存在。'
    console.error(error)
  }

  setQuickRange('today')
})

onUnmounted(() => {
  window.removeEventListener('resize', resizeHandler)
  myChart?.dispose()
  myChart = null
})
</script>

<style scoped>
/* 整页固定定位，覆盖视口，不再依赖 #app 的 padding/max-width */
.room-detail-page {
  position: fixed;
  inset: 0;
  padding: 1rem;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

/* flex 链：page → Card(root) → Card(body) → Card(content) → chart */
.room-detail-page :deep(.p-card) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.room-detail-page :deep(.p-card-body) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.room-detail-page :deep(.p-card-content) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  padding-top: 0;
}

.chart-wrap {
  position: relative; /* 让内部 no-data-overlay 绝对定位生效 */
  flex: 1;           /* 填满 p-card-content 剩余高度 */
  min-height: 0;     /* flex 子元素防止内容溢出的关键 */
}

.no-data-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--p-text-muted-color);
  font-size: 1rem;
  background: var(--p-content-background);
  border-radius: 6px;
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.filter-left {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}

.filter-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.stat-icon {
  margin-right: 0.35rem;
  font-size: 0.9rem;
  opacity: 0.75;
}

.stat-card {
  width: 100%;
  height: 100%;
  min-height: 110px;
  border: 1px solid var(--p-content-border-color);
  border-radius: 10px;
  padding: 1rem;
  background: var(--p-content-background);
  display: flex;
  flex-direction: column;
}

.stat-title {
  color: var(--p-text-muted-color);
  font-size: 0.875rem;
}

.stat-value {
  margin-top: 0.5rem;
  font-size: 1.25rem;
  font-weight: 600;
  line-height: 1.35;
  min-height: 3.4rem;
}
</style>

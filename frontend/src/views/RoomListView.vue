<template>
  <div class="room-list-page">
    <Card>
      <template #title>
        <div class="flex align-items-center justify-content-between">
          <div class="flex align-items-center gap-2">
            <i class="pi pi-home text-blue-500"></i>
            <span>房间列表管理</span>
          </div>
          <Button
            v-if="isLoggedIn"
            label="房间管理"
            icon="pi pi-cog"
            size="small"
            @click="$router.push('/rooms/manage')"
          />
        </div>
      </template>

      <template #content>
        <DataTable
          :value="rooms"
          lazy
          paginator
          :rows="lazyParams.rows"
          :rowsPerPageOptions="[5, 10, 20]"
          :totalRecords="totalRecords"
          :loading="loading"
          @page="onPage($event)"
          dataKey="id"
          scrollable
          scrollHeight="calc(100vh - 220px)"
        >
          <Column field="id" header="ID" style="width: 15%"></Column>
          <Column field="name" header="房间名称"></Column>

          <Column header="操作" style="width: 25%">
            <template #body="slotProps">
              <Button
                label="查看详情"
                icon="pi pi-chart-line"
                text
                @click="$router.push(`/rooms/${slotProps.data.id}`)"
              />
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import axios from 'axios'
import DataTable, { type DataTablePageEvent } from 'primevue/datatable'
import Column from 'primevue/column'
import Button from 'primevue/button'
import Card from 'primevue/card'

const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

interface Room {
  id: number
  name: string
  area_id: string
  building_code: string
  floor_code: string
  room_code: string
  created_at: string
}

const rooms = ref<Room[]>([])
const loading = ref(false)
const totalRecords = ref(0)
const isLoggedIn = computed(() => !!localStorage.getItem('accesstoken'))

// 对应后端的 listRoomsRequest
const lazyParams = ref({
  page: 1, // 对应 page_id
  rows: 5, // 对应 page_size
})

const loadLazyData = async () => {
  loading.value = true
  try {
    const response = await axios.get(`${API_BASE}/rooms`, {
      params: {
        page_id: lazyParams.value.page,
        page_size: lazyParams.value.rows,
      },
    })

    // 映射后端返回的 resp 数组
    rooms.value = response.data.rooms
    totalRecords.value = response.data.total
  } catch (error) {
    console.error('加载房间列表失败', error)
  } finally {
    loading.value = false
  }
}

// 翻页事件（PrimeVue 的 page 是 0-based，后端 page_id 是 1-based，所以要 +1）
const onPage = (event: DataTablePageEvent) => {
  lazyParams.value.page = event.page + 1
  lazyParams.value.rows = event.rows
  loadLazyData()
}

onMounted(() => {
  loadLazyData()
})
</script>

<style scoped>
/* 居中容器，宽度大约占屏幹50%（最小 480px、最大 760px） */
.room-list-page {
  width: clamp(480px, 50vw, 760px);
  margin: 0 auto;
}
</style>
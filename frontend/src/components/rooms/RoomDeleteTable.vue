<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import axios from 'axios'
import { apiClient } from '@/client'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import DataTable, { type DataTablePageEvent } from 'primevue/datatable'
import Column from 'primevue/column'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import store from '@/store'

interface ManagedRoom {
  id: number
  name: string
  area_id: string
  building_code: string
  floor_code: string
  room_code: string
  created_at: string
}

const props = withDefaults(
  defineProps<{
    refreshToken?: number
  }>(),
  {
    refreshToken: 0,
  },
)

const router = useRouter()
const toast = useToast()

const managedRooms = ref<ManagedRoom[]>([])
const loadingList = ref(false)
const totalRecords = ref(0)
const deletingId = ref<number | null>(null)
const lazyParams = ref({
  page: 1,
  rows: 5,
})

const forceLogin = (reason: 'missing' | 'expired') => {
  localStorage.removeItem('accesstoken')
  store.clearUser()
  router.replace({ path: '/login', query: { redirect: '/rooms/manage', auth: reason } })
}

const ensureToken = () => {
  const token = localStorage.getItem('accesstoken')
  if (token) {
    return true
  }
  forceLogin('missing')
  return false
}

const handleAuthError = (error: unknown, fallbackSummary: string) => {
  if (axios.isAxiosError(error) && error.response?.status === 401) {
    forceLogin('expired')
    return true
  }

  toast.add({ severity: 'error', summary: fallbackSummary, life: 3000 })
  return false
}

const loadLazyData = async () => {
  if (!ensureToken()) return

  loadingList.value = true
  try {
    const response = await apiClient.get<{ rooms: ManagedRoom[]; total: number }>('/rooms', {
      params: {
        page_id: lazyParams.value.page,
        page_size: lazyParams.value.rows,
      },
    })

    managedRooms.value = response.data.rooms
    totalRecords.value = response.data.total
  } catch (error: unknown) {
    handleAuthError(error, '加载房间列表失败')
  } finally {
    loadingList.value = false
  }
}

const onPage = (event: DataTablePageEvent) => {
  lazyParams.value.page = event.page + 1
  lazyParams.value.rows = event.rows
  loadLazyData()
}

const handleDeleteRoom = async (room: ManagedRoom) => {
  if (!ensureToken()) return
  if (!window.confirm(`确认删除房间：${room.name}？`)) return

  deletingId.value = room.id
  try {
    await apiClient.delete(`/rooms/${room.id}`)

    toast.add({ severity: 'success', summary: '删除成功', detail: room.name, life: 3000 })

    if (managedRooms.value.length === 1 && lazyParams.value.page > 1) {
      lazyParams.value.page -= 1
    }
    await loadLazyData()
  } catch (error: unknown) {
    handleAuthError(error, '删除房间失败')
  } finally {
    deletingId.value = null
  }
}

watch(
  () => props.refreshToken,
  () => {
    loadLazyData()
  },
)

onMounted(() => {
  if (!ensureToken()) return
  loadLazyData()
})
</script>

<template>
  <Toast />
  <Card class="list-card">
    <template #title>
      房间列表
    </template>
    <template #content>
      <DataTable
        :value="managedRooms"
        lazy
        paginator
        :rows="lazyParams.rows"
        :rowsPerPageOptions="[5, 10, 20]"
        :totalRecords="totalRecords"
        :loading="loadingList"
        @page="onPage($event)"
        dataKey="id"
        scrollable
        scrollHeight="calc(100vh - 280px)"
      >
        <Column field="id" header="ID" style="width: 18%" />
        <Column field="name" header="房间名称" />
        <Column header="操作" style="width: 30%">
          <template #body="slotProps">
            <Button
              label="删除"
              icon="pi pi-trash"
              severity="danger"
              text
              :loading="deletingId === slotProps.data.id"
              @click="handleDeleteRoom(slotProps.data)"
            />
          </template>
        </Column>
      </DataTable>
    </template>
  </Card>
</template>

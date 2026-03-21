<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import axios from 'axios'
import { apiClient } from '@/client'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
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

interface RoomNotification {
  room_id: number
  threshold: number
  is_enabled: boolean
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
const toastGroup = 'room-notification'

const managedRooms = ref<ManagedRoom[]>([])
const notifications = ref<Map<number, RoomNotification>>(new Map())
const loadingRooms = ref(false)
const loadingNotifications = ref(false)

const showConfigDialog = ref(false)
const selectedRoom = ref<ManagedRoom | null>(null)
const notificationThreshold = ref(0)
const notificationEnabled = ref(true)
const submitting = ref(false)

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

  if (!axios.isAxiosError(error)) {
    toast.add({ group: toastGroup, severity: 'error', summary: fallbackSummary, life: 3000 })
  }
  return false
}

const loadManagedRooms = async () => {
  if (!ensureToken()) return

  loadingRooms.value = true
  try {
    const response = await apiClient.get<{ rooms: ManagedRoom[]; total: number }>('/rooms', {
      params: {
        page_id: 1,
        page_size: 50,
      },
    })

    managedRooms.value = response.data.rooms
  } catch (error: unknown) {
    handleAuthError(error, '加载房间列表失败')
  } finally {
    loadingRooms.value = false
  }
}

const loadNotifications = async () => {
  if (!ensureToken()) return

  loadingNotifications.value = true
  try {
    const response = await apiClient.get<{ notifications: RoomNotification[] }>(
      '/notifications',
      {
        params: {
          page_id: 1,
          page_size: 50,
        },
      },
    )

    const notificationMap = new Map<number, RoomNotification>()
    response.data.notifications.forEach((n) => {
      notificationMap.set(n.room_id, n)
    })
    notifications.value = notificationMap
  } catch (error: unknown) {
    handleAuthError(error, '加载通知配置失败')
  } finally {
    loadingNotifications.value = false
  }
}

const getNotificationForRoom = (roomId: number) => {
  return notifications.value.get(roomId)
}

const openConfigDialog = (room: ManagedRoom) => {
  selectedRoom.value = room
  const existingNotification = getNotificationForRoom(room.id)
  notificationThreshold.value = existingNotification?.threshold ?? 0
  notificationEnabled.value = existingNotification?.is_enabled ?? true
  showConfigDialog.value = true
}

const handleSaveNotification = async () => {
  if (!selectedRoom.value || notificationThreshold.value < 0) {
    return
  }

  if (!ensureToken()) return

  submitting.value = true

  const isUpdate = !!getNotificationForRoom(selectedRoom.value.id)

  try {
    if (isUpdate) {
      // 更新现有通知
      await apiClient.patch(`/notifications/${selectedRoom.value.id}`, {
        threshold: notificationThreshold.value,
        is_enabled: notificationEnabled.value,
      })
      toast.add({
        group: toastGroup,
        severity: 'success',
        summary: '更新成功',
        detail: `房间 ${selectedRoom.value.name} 已更新`,
        life: 2000,
      })
    } else {
      // 创建新通知
      await apiClient.post('/notifications', {
        room_id: selectedRoom.value.id,
        threshold: notificationThreshold.value,
      })
      toast.add({
        group: toastGroup,
        severity: 'success',
        summary: '添加成功',
        detail: `房间 ${selectedRoom.value.name} 已添加到监控`,
        life: 2000,
      })
    }

    showConfigDialog.value = false
    await loadNotifications()
  } catch (error: unknown) {
    if (handleAuthError(error, isUpdate ? '更新失败' : '添加监控失败')) {
      return
    }

    if (axios.isAxiosError(error)) {
      const msg = (error.response?.data as { error?: string })?.error ?? '操作失败'
      toast.add({ group: toastGroup, severity: 'error', summary: '操作失败', detail: msg, life: 3200 })
      return
    }

    toast.add({ group: toastGroup, severity: 'error', summary: '操作失败', life: 3200 })
  } finally {
    submitting.value = false
  }
}

const handleDeleteNotification = async (room: ManagedRoom) => {
  if (!ensureToken()) return
  if (!window.confirm(`确认删除房间 ${room.name} 的监控配置？`)) return

  submitting.value = true

  try {
    await apiClient.delete(`/notifications/${room.id}`)
    toast.add({
      group: toastGroup,
      severity: 'success',
      summary: '删除成功',
      detail: `房间 ${room.name} 已删除监控配置`,
      life: 2000,
    })
    await loadNotifications()
  } catch (error: unknown) {
    if (handleAuthError(error, '删除失败')) {
      return
    }

    if (axios.isAxiosError(error)) {
      const msg = (error.response?.data as { error?: string })?.error ?? '删除失败'
      toast.add({ group: toastGroup, severity: 'error', summary: '删除失败', detail: msg, life: 3200 })
      return
    }

    toast.add({ group: toastGroup, severity: 'error', summary: '删除失败', life: 3200 })
  } finally {
    submitting.value = false
  }
}

watch(
  () => props.refreshToken,
  () => {
    loadManagedRooms()
    loadNotifications()
  },
)

onMounted(() => {
  if (!ensureToken()) return
  loadManagedRooms()
  loadNotifications()
})
</script>

<template>
  <Toast :group="toastGroup" />
  <Card class="threshold-card">
    <template #title>
      <div class="flex align-items-center gap-2">
        <i class="pi pi-bell text-orange-500"></i>
        <span>通知阈值设置</span>
      </div>
    </template>

    <template #content>
      <DataTable
        :value="managedRooms"
        :loading="loadingRooms"
        stripedRows
        scrollable
        scrollHeight="24rem"
        responsive-layout="scroll"
        class="rooms-table"
      >
        <Column field="name" header="房间名称" />
        <Column header="监控阈值" class="status-column">
          <template #body="slotProps">
            <div class="threshold-status">
              <span v-if="getNotificationForRoom(slotProps.data.id) as any" class="threshold-badge">
                {{ getNotificationForRoom(slotProps.data.id)!.threshold }}
              </span>
              <span v-else class="threshold-badge disabled">未配置</span>
            </div>
          </template>
        </Column>
        <Column header="状态" class="status-column">
          <template #body="slotProps">
            <div v-if="getNotificationForRoom(slotProps.data.id) as any" class="status-badge">
              <span
                v-if="getNotificationForRoom(slotProps.data.id)!.is_enabled"
                class="status-enabled"
              >
                <i class="pi pi-check"></i>
                启用
              </span>
              <span v-else class="status-disabled">
                <i class="pi pi-times"></i>
                禁用
              </span>
            </div>
            <div v-else class="status-badge">
              <span class="status-none">
                <i class="pi pi-minus"></i>
                -
              </span>
            </div>
          </template>
        </Column>
        <Column header="操作" class="actions-column">
          <template #body="slotProps">
            <div class="action-buttons">
              <Button
                v-if="getNotificationForRoom(slotProps.data.id) as any"
                icon="pi pi-pencil"
                rounded
                text
                severity="secondary"
                @click="openConfigDialog(slotProps.data)"
                :disabled="submitting"
                title="编辑"
              />
              <Button
                icon="pi pi-plus"
                rounded
                text
                severity="success"
                @click="openConfigDialog(slotProps.data)"
                v-if="!(getNotificationForRoom(slotProps.data.id) as any)"
                :disabled="submitting"
                title="添加监控"
              />
              <Button
                v-if="getNotificationForRoom(slotProps.data.id) as any"
                icon="pi pi-trash"
                rounded
                text
                severity="danger"
                @click="handleDeleteNotification(slotProps.data)"
                :disabled="submitting"
                title="删除"
              />
            </div>
          </template>
        </Column>

        <template #empty>
          <div class="empty-message">
            <i class="pi pi-inbox"></i>
            <p>还没有添加任何房间</p>
          </div>
        </template>
      </DataTable>
    </template>
  </Card>

  <Dialog
    v-model:visible="showConfigDialog"
    header="设置监控阈值"
    modal
    :style="{ width: 'min(92vw, 360px)' }"
  >
    <div class="config-form">
      <div v-if="selectedRoom" class="field-block">
        <label>房间</label>
        <div class="p-2" style="background-color: var(--p-surface-50, #f3f4f6); border-radius: var(--p-border-radius)">
          <strong>{{ selectedRoom.name }}</strong>
        </div>
      </div>

      <div class="field-block">
        <label for="threshold">告警阈值 (度)</label>
        <InputNumber
          id="threshold"
          v-model="notificationThreshold"
          :min="0"
          :max="1000"
          class="w-full"
        />
        <small>当用电量达到此值时会收到通知</small>
      </div>

      <div v-if="selectedRoom && (getNotificationForRoom(selectedRoom.id) as any)" class="field-block">
        <label>启用状态</label>
        <Button
          :label="notificationEnabled ? '已启用' : '已禁用'"
          :icon="notificationEnabled ? 'pi pi-check-circle' : 'pi pi-ban'"
          :severity="notificationEnabled ? 'success' : 'secondary'"
          outlined
          @click="notificationEnabled = !notificationEnabled"
          :disabled="submitting"
        />
        <small>可在编辑时直接切换当前房间监控的启用状态</small>
      </div>
    </div>

    <template #footer>
      <Button label="取消" text @click="showConfigDialog = false" :disabled="submitting" />
      <Button
        label="保存"
        icon="pi pi-check"
        :loading="submitting"
        :disabled="submitting"
        @click="handleSaveNotification"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.threshold-card {
  height: 100%;
  width: 100%;
}

.rooms-table {
  font-size: 0.9rem;
}

.status-column,
.actions-column {
  width: 8rem;
  text-align: center;
}

.threshold-badge {
  display: inline-block;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 500;
  background-color: var(--p-blue-100, #dbeafe);
  color: var(--p-blue-700, #1e40af);
}

.threshold-badge.disabled {
  background-color: var(--p-gray-100, #f3f4f6);
  color: var(--p-gray-600, #4b5563);
}

.status-badge {
  display: inline-block;
}

.status-enabled {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 500;
  background-color: var(--p-green-100, #dcfce7);
  color: var(--p-green-700, #16a34a);
}

.status-disabled {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 500;
  background-color: var(--p-red-100, #fee2e2);
  color: var(--p-red-700, #b91c1c);
}

.status-none {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.35rem 0.65rem;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 500;
  background-color: var(--p-gray-100, #f3f4f6);
  color: var(--p-gray-600, #4b5563);
}

.action-buttons {
  display: flex;
  gap: 0.35rem;
  justify-content: center;
}

.empty-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--p-text-color-secondary);
  text-align: center;
}

.empty-message i {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
  opacity: 0.5;
}

.config-form {
  display: flex;
  flex-direction: column;
  gap: 0.95rem;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.field-block label {
  font-size: 0.92rem;
  font-weight: 500;
}

.field-block small {
  font-size: 0.8rem;
  color: var(--p-text-color-secondary);
}
</style>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import axios from 'axios'
import { apiClient } from '@/client'
import Card from 'primevue/card'
import Select from 'primevue/select'
import Button from 'primevue/button'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import { useRouter } from 'vue-router'

interface AreaItem { id: string; areaName: string }
interface BuildingItem { buildingCode: string; buildingName: string }
interface FloorItem { floorCode: string; floorName: string }
interface RoomItem { roomCode: string; roomName: string }

const props = withDefaults(
  defineProps<{
    redirectPath?: string
  }>(),
  {
    redirectPath: '/rooms/manage',
  },
)

const emit = defineEmits<{
  (e: 'added', name: string): void
}>()

const areas = ref<AreaItem[]>([])
const buildings = ref<BuildingItem[]>([])
const floors = ref<FloorItem[]>([])
const rooms = ref<RoomItem[]>([])

const selectedArea = ref<AreaItem | null>(null)
const selectedBuilding = ref<BuildingItem | null>(null)
const selectedFloor = ref<FloorItem | null>(null)
const selectedRoom = ref<RoomItem | null>(null)

const loadingAreas = ref(false)
const loadingBuildings = ref(false)
const loadingFloors = ref(false)
const loadingRooms = ref(false)
const submitting = ref(false)

const toast = useToast()
const router = useRouter()

const isReady = computed(
  () => !!(selectedArea.value && selectedBuilding.value && selectedFloor.value && selectedRoom.value),
)

const forceLogin = (reason: 'missing' | 'expired') => {
  localStorage.removeItem('accesstoken')
  router.replace({ path: '/login', query: { redirect: props.redirectPath, auth: reason } })
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

const fetchAreas = async () => {
  if (!ensureToken()) return

  loadingAreas.value = true
  try {
    const { data } = await apiClient.get('/proxy/areas')
    areas.value = (data.rows ?? []) as AreaItem[]
  } catch (error: unknown) {
    handleAuthError(error, '获取校区失败')
  } finally {
    loadingAreas.value = false
  }
}

const fetchBuildings = async (areaId: string) => {
  if (!ensureToken()) return

  loadingBuildings.value = true
  try {
    const { data } = await apiClient.get('/proxy/buildings', {
      params: { areaId },
    })
    buildings.value = (data.rows ?? []) as BuildingItem[]
  } catch (error: unknown) {
    handleAuthError(error, '获取楼栋失败')
  } finally {
    loadingBuildings.value = false
  }
}

const fetchFloors = async (areaId: string, buildingCode: string) => {
  if (!ensureToken()) return

  loadingFloors.value = true
  try {
    const { data } = await apiClient.get('/proxy/floors', {
      params: { areaId, buildingCode },
    })
    floors.value = (data.rows ?? []) as FloorItem[]
  } catch (error: unknown) {
    handleAuthError(error, '获取楼层失败')
  } finally {
    loadingFloors.value = false
  }
}

const fetchRooms = async (areaId: string, buildingCode: string, floorCode: string) => {
  if (!ensureToken()) return

  loadingRooms.value = true
  try {
    const { data } = await apiClient.get('/proxy/rooms', {
      params: { areaId, buildingCode, floorCode },
    })
    rooms.value = (data.rows ?? []) as RoomItem[]
  } catch (error: unknown) {
    handleAuthError(error, '获取房间失败')
  } finally {
    loadingRooms.value = false
  }
}

watch(selectedArea, (area) => {
  selectedBuilding.value = null
  selectedFloor.value = null
  selectedRoom.value = null
  buildings.value = []
  floors.value = []
  rooms.value = []
  if (area) fetchBuildings(area.id)
})

watch(selectedBuilding, (building) => {
  selectedFloor.value = null
  selectedRoom.value = null
  floors.value = []
  rooms.value = []
  if (building && selectedArea.value) {
    fetchFloors(selectedArea.value.id, building.buildingCode)
  }
})

watch(selectedFloor, (floor) => {
  selectedRoom.value = null
  rooms.value = []
  if (floor && selectedArea.value && selectedBuilding.value) {
    fetchRooms(selectedArea.value.id, selectedBuilding.value.buildingCode, floor.floorCode)
  }
})

const handleAddRoom = async () => {
  if (!isReady.value || !ensureToken()) return

  submitting.value = true
  try {
    const surplusResp = await apiClient.get('/proxy/room-surplus', {
      params: {
        areaId: selectedArea.value!.id,
        buildingCode: selectedBuilding.value!.buildingCode,
        floorCode: selectedFloor.value!.floorCode,
        roomCode: selectedRoom.value!.roomCode,
      },
    })

    const displayName: string =
      surplusResp.data?.data?.displayRoomName ?? selectedRoom.value!.roomName

    await apiClient.post('/rooms', {
      name: displayName,
      area_id: selectedArea.value!.id,
      building_code: selectedBuilding.value!.buildingCode,
      floor_code: selectedFloor.value!.floorCode,
      room_code: selectedRoom.value!.roomCode,
    })

    toast.add({ severity: 'success', summary: '添加成功', detail: displayName, life: 3000 })
    selectedArea.value = null
    emit('added', displayName)
  } catch (error: unknown) {
    if (handleAuthError(error, '添加房间失败')) {
      return
    }

    if (axios.isAxiosError(error)) {
      const msg = (error.response?.data as { error?: string })?.error ?? '请求失败'
      toast.add({ severity: 'error', summary: '添加失败', detail: msg, life: 4000 })
      return
    }

    toast.add({ severity: 'error', summary: '添加失败', life: 4000 })
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  if (!ensureToken()) return
  fetchAreas()
})
</script>

<template>
  <Toast />
  <Card class="form-card">
    <template #title>
      <div class="flex align-items-center gap-2">
        <i class="pi pi-plus-circle text-blue-500"></i>
        <span>添加监控房间</span>
      </div>
    </template>

    <template #content>
      <div class="flex flex-column row-gap-4">
        <div class="flex flex-column row-gap-1">
          <label class="form-label">校区</label>
          <Select
            v-model="selectedArea"
            :options="areas"
            optionLabel="areaName"
            placeholder="请选择校区"
            :loading="loadingAreas"
            class="w-full"
          />
        </div>

        <div class="flex flex-column row-gap-1">
          <label class="form-label">楼栋</label>
          <Select
            v-model="selectedBuilding"
            :options="buildings"
            optionLabel="buildingName"
            placeholder="请先选择校区"
            :disabled="!selectedArea"
            :loading="loadingBuildings"
            class="w-full"
          />
        </div>

        <div class="flex flex-column row-gap-1">
          <label class="form-label">楼层</label>
          <Select
            v-model="selectedFloor"
            :options="floors"
            optionLabel="floorName"
            placeholder="请先选择楼栋"
            :disabled="!selectedBuilding"
            :loading="loadingFloors"
            class="w-full"
          />
        </div>

        <div class="flex flex-column row-gap-1">
          <label class="form-label">房间</label>
          <Select
            v-model="selectedRoom"
            :options="rooms"
            optionLabel="roomName"
            placeholder="请先选择楼层"
            :disabled="!selectedFloor"
            :loading="loadingRooms"
            class="w-full"
          />
        </div>

        <Button
          label="添加房间"
          icon="pi pi-check"
          :disabled="!isReady || submitting"
          :loading="submitting"
          @click="handleAddRoom"
          class="mt-2"
        />
      </div>
    </template>
  </Card>
</template>

<style scoped>
.form-card {
  height: 100%;
  width: 100%;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--p-text-color);
}
</style>

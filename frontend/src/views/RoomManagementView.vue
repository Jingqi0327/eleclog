<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import store from '@/store'
import RoomAddForm from '@/components/rooms/RoomAddForm.vue'
import RoomDeleteTable from '@/components/rooms/RoomDeleteTable.vue'

const router = useRouter()
const refreshToken = ref(0)

const handleRoomAdded = () => {
  refreshToken.value += 1
}

const handleLogout = async () => {
  localStorage.removeItem('accesstoken')
  store.clearUser()
  await router.replace({ path: '/login', query: { auth: 'logout' } })
}
</script>

<template>
  <div class="room-management-page">
    <div class="page-header">
      <div class="title-group">
        <i class="pi pi-home text-blue-500"></i>
        <h2>房间管理</h2>
      </div>
      <Button
        label="退出登录"
        icon="pi pi-sign-out"
        severity="secondary"
        outlined
        @click="handleLogout"
      />
    </div>

    <div class="management-grid">
      <RoomDeleteTable :refreshToken="refreshToken" />
      <RoomAddForm redirectPath="/rooms/manage" @added="handleRoomAdded" />
    </div>
  </div>
</template>

<style scoped>
.room-management-page {
  min-height: 100vh;
  padding: 1rem;
  background: var(--p-surface-ground, #f8f9fa);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.title-group h2 {
  margin: 0;
  font-size: 1.2rem;
}

.management-grid {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 1rem;
  align-items: start;
}

@media (max-width: 1024px) {
  .management-grid {
    grid-template-columns: 1fr;
  }
}
</style>

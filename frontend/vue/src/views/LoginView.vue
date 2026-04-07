<script setup lang="ts">
import LoginUser from '@/components/LoginUser.vue'
import Toast from 'primevue/toast'
import { onMounted } from 'vue'
import { useToast } from 'primevue/usetoast'
import { useRoute, useRouter } from 'vue-router'

const toast = useToast()
const route = useRoute()
const router = useRouter()

onMounted(async () => {
  const authReason = route.query.auth
  if (authReason === 'expired') {
    toast.add({ severity: 'warn', summary: '登录已过期，请重新登录', life: 2500 })
  } else if (authReason === 'missing') {
    toast.add({ severity: 'warn', summary: '未登录，请先登录', life: 2500 })
  } else if (authReason === 'logout') {
    toast.add({ severity: 'success', summary: '已退出登录', life: 2500 })
  } else {
    return
  }

  const rest = { ...route.query }
  delete rest.auth
  await router.replace({ path: route.path, query: rest })
})
</script>

<template>
  <div class="login-page">
    <Toast />
    <div class="login-card">
      <h1 class="green">User Login</h1>
      <LoginUser  />
    </div>
  </div>
</template>

<style scoped>
h1 {
  font-weight: 500;
  font-size: 2.6rem;
  margin-bottom: 1.5rem;
}

/* 占满视口、垂直水平居中 */
.login-page {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--p-surface-ground, #f8f9fa);
}

/* 居中卡片，宽度限制在屏幵95平分戔50% */
.login-card {
  width: min(95vw, 440px);
  padding: 2.5rem 2rem;
  border-radius: 12px;
  background: var(--p-content-background);
  border: 1px solid var(--p-content-border-color);
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}
</style>

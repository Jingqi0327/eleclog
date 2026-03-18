<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { apiClient } from '@/client'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Menu from 'primevue/menu'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import store from '@/store'
import type { User } from '@/types/user'

interface UpdateUserApiResponse {
  user: {
    username: string
    full_name: string
    email: string
  }
}

const router = useRouter()
const toast = useToast()

const menuRef = ref()
const showEditDialog = ref(false)
const submitting = ref(false)

const fullName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')

const currentUser = computed(() => store.state.user)
const username = computed(() => currentUser.value?.username ?? '')
const displayName = computed(() => currentUser.value?.full_name || currentUser.value?.username || '用户')

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const isEmailValid = computed(() => emailPattern.test(email.value.trim()))
const hasPasswordInput = computed(() => password.value.length > 0 || confirmPassword.value.length > 0)
const isPasswordLengthValid = computed(() => password.value.length === 0 || password.value.length >= 6)
const isPasswordMatch = computed(() => password.value === confirmPassword.value)

const canSubmit = computed(() => {
  if (!fullName.value.trim() || !isEmailValid.value) {
    return false
  }

  if (!hasPasswordInput.value) {
    return true
  }

  return isPasswordLengthValid.value && isPasswordMatch.value && password.value.length >= 6
})

const forceLogin = (reason: 'missing' | 'expired') => {
  localStorage.removeItem('accesstoken')
  store.clearUser()
  router.replace({ path: '/login', query: { redirect: '/rooms/manage', auth: reason } })
}

const openEditDialog = () => {
  if (!currentUser.value) {
    toast.add({ severity: 'warn', summary: '用户信息丢失，请重新登录', life: 2600 })
    forceLogin('missing')
    return
  }

  fullName.value = currentUser.value.full_name
  email.value = currentUser.value.email
  password.value = ''
  confirmPassword.value = ''
  showEditDialog.value = true
}

const menuItems = ref([
  {
    label: '修改信息',
    icon: 'pi pi-user-edit',
    command: () => openEditDialog(),
  },
])

const toggleMenu = (event: Event) => {
  menuRef.value?.toggle(event)
}

const submitUpdate = async () => {
  if (!canSubmit.value || !username.value) {
    return
  }

  const token = localStorage.getItem('accesstoken')
  if (!token) {
    forceLogin('missing')
    return
  }

  submitting.value = true

  try {
    const payload: {
      username: string
      full_name: string
      email: string
      password?: string
    } = {
      username: username.value,
      full_name: fullName.value.trim(),
      email: email.value.trim(),
    }

    if (password.value) {
      payload.password = password.value
    }

    const { data } = await apiClient.put<UpdateUserApiResponse>('/users', payload)

    const updatedUser: User = {
      username: data.user.username,
      full_name: data.user.full_name,
      email: data.user.email,
    }

    store.setUser(updatedUser, token)
    showEditDialog.value = false
    toast.add({ severity: 'success', summary: '用户信息已更新', life: 2600 })
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      forceLogin('expired')
      return
    }

    if (axios.isAxiosError(error)) {
      const msg = (error.response?.data as { error?: string })?.error ?? '更新失败'
      toast.add({ severity: 'error', summary: '修改失败', detail: msg, life: 3200 })
      return
    }

    toast.add({ severity: 'error', summary: '修改失败', life: 3200 })
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Toast />

  <div class="user-profile-menu">
    <Button
      type="button"
      severity="secondary"
      outlined
      @click="toggleMenu"
      aria-haspopup="true"
      aria-controls="user_menu"
      class="user-trigger"
    >
      <span class="user-trigger-content">
        <i class="pi pi-user"></i>
        <span class="user-name">{{ displayName }}</span>
        <i class="pi pi-angle-down"></i>
      </span>
    </Button>
    <Menu id="user_menu" ref="menuRef" :model="menuItems" :popup="true" />
  </div>

  <Dialog
    v-model:visible="showEditDialog"
    header="修改用户信息"
    modal
    :style="{ width: 'min(92vw, 460px)' }"
  >
    <div class="edit-form">
      <div class="field-block">
        <label for="fullname">Fullname</label>
        <InputText id="fullname" v-model="fullName" class="w-full" />
      </div>

      <div class="field-block">
        <label for="email">Email</label>
        <InputText id="email" v-model="email" type="email" class="w-full" />
        <small v-if="email && !isEmailValid" class="error-text">请输入正确的邮箱格式</small>
      </div>

      <div class="field-block">
        <label for="password">Password</label>
        <Password
          id="password"
          v-model="password"
          toggleMask
          :feedback="false"
          inputClass="w-full"
          class="w-full"
        />
        <small v-if="password && !isPasswordLengthValid" class="error-text">
          密码长度至少 6 位
        </small>
      </div>

      <div class="field-block">
        <label for="confirmPassword">Confirm Password</label>
        <Password
          id="confirmPassword"
          v-model="confirmPassword"
          toggleMask
          :feedback="false"
          inputClass="w-full"
          class="w-full"
        />
        <small v-if="hasPasswordInput && !isPasswordMatch" class="error-text">
          两次输入的密码不一致
        </small>
      </div>
    </div>

    <template #footer>
      <Button label="取消" text @click="showEditDialog = false" :disabled="submitting" />
      <Button
        label="保存"
        icon="pi pi-check"
        :loading="submitting"
        :disabled="!canSubmit || submitting"
        @click="submitUpdate"
      />
    </template>
  </Dialog>
</template>

<style scoped>
.user-profile-menu {
  display: flex;
  align-items: center;
}

.user-trigger {
  max-width: 230px;
  min-width: 140px;
}

.user-trigger-content {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.user-name {
  max-width: 130px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
}

.edit-form {
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

.error-text {
  color: var(--p-red-500, #ef4444);
}
</style>

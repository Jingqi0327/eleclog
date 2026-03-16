<script setup lang="ts">
//ts指定了脚本的语言为TypeScript
import InputGroup from 'primevue/inputgroup'
import InputGroupAddon from 'primevue/inputgroupaddon'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import Button from 'primevue/button'
import { computed, ref } from 'vue'
import axios from 'axios'
import type { User } from '@/types/user'
import store from '@/store'
import { useToast } from 'primevue/usetoast'
import { useRoute, useRouter } from 'vue-router'

interface LoginResponse {
  user: User
  access_token: string
}

const username = ref<string>('')
const password = ref<string>('')
// 计算属性，判断登录按钮是否应该被禁用
const isLoginDisabled = computed(() => !username.value || !password.value)

const errorMessage = ref<string>('')
const toast = useToast()
const router = useRouter()
const route = useRoute()

const handleLogin = async () => {
  try {
    const response = await axios.post<LoginResponse>('http://localhost:8080/users/login', {
      username: username.value,
      password: password.value,
    })

    store.setUser(response.data.user, response.data.access_token)
    localStorage.setItem('accesstoken', response.data.access_token)
    toast.add({
      severity: 'success',
      summary: '登录成功',
      detail: `欢迎回来，${response.data.user.full_name}`,
      life: 3000,
    })
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/rooms/manage'
    await router.replace(redirect)
  } catch (error: unknown) {
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      errorMessage.value = (error.response.data as { message?: string })?.message ?? 'User not found'
    } else {
      errorMessage.value = 'An error occurred. Please try again later'
    }
    toast.add({
      severity: 'error',
      summary: 'Login failed',
      detail: errorMessage.value,
      life: 3000,
    })
  }
}
</script>

<template>
  <div class="flex flex-column row-gap-5">
    <InputGroup>
      <InputGroupAddon>
        <i class="pi pi-user"></i>
      </InputGroupAddon>
      <FloatLabel>
        <InputText id="username" v-model="username" />
        <label for="username">Username</label>
      </FloatLabel>
    </InputGroup>

    <InputGroup>
      <InputGroupAddon>
        <i class="pi pi-lock"></i>
      </InputGroupAddon>
      <FloatLabel>
        <InputText id="password" type="password" v-model="password" />
        <label for="password">Password</label>
      </FloatLabel>
    </InputGroup>
    <Button label="Login" :disabled="isLoginDisabled" @click="handleLogin" />
  </div>
</template>

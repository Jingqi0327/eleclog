import { reactive, readonly } from 'vue'
import type { AuthState } from './types/auth_state'
import type { User } from './types/user'

// reactive: 让对象变成“活的”。一旦对象里的属性变了，所有用到它的 UI 界面都会自动刷新。
// <AuthState>: 这是一个“泛型”，告诉 Vue 这个对象必须符合我们刚才定义的结构。
const state = reactive<AuthState>({
  user: null,
  accessToken: null
})

// 这个方法用来设置用户信息
function setUser(user: User, accessToken: string) {
  state.user = user
  state.accessToken = accessToken
}

// 这个方法用来清除用户信息
function clearUser() {
  state.user = null
  state.accessToken = null
}

// 导出一个对象，这个对象包含了我们定义的状态和操作状态的方法
export default {
  state: readonly(state),// 只读包装
  setUser,
  clearUser
}

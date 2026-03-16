import type { User } from './user.ts'

export interface AuthState {
  user: User | null
  accessToken: string | null
}

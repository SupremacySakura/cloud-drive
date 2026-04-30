import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore(
  'user',
  () => {
    const token = ref('')
    const refreshToken = ref('')
    const setToken = (newToken: string) => {
      token.value = newToken
    }
    const setRefreshToken = (newRefreshToken: string) => {
      refreshToken.value = newRefreshToken
    }

    const logout = () => {
      token.value = ''
      refreshToken.value = ''
    }

    const isLoggedIn = computed(() => !!token.value)

    return {
      token,
      refreshToken,
      isLoggedIn,
      setToken,
      setRefreshToken,
      logout,
    }
  },
  {
    persist: true,
  },
)

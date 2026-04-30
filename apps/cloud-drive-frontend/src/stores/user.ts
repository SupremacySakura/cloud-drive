import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore(
  'user',
  () => {
    const token = ref('')
    const setToken = (newToken: string) => {
      token.value = newToken
    }

    const logout = () => {
      token.value = ''
    }

    const isLoggedIn = computed(() => !!token.value)

    return {
      token,
      isLoggedIn,
      setToken,
      logout,
    }
  },
  {
    persist: true,
  },
)

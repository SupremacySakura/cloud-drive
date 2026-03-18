import { defineStore } from "pinia";
import { ref } from "vue";

export const useUserStore = defineStore("user", () => {
    const token = ref('')
    const setToken = (newToken: string) => {
        token.value = newToken
    }

    const logout = () => {
        token.value = ''
    }
    return {
        token,
        setToken,
        logout
    }
}, {
    persist: true
});
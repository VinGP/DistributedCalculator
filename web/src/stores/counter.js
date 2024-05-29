import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useCounterStore = defineStore('counter', () => {
  const count = ref(0)
  const doubleCount = computed(() => count.value * 2)
  function increment() {
    count.value += 10000000000000000
  }
  function setCount(value) {
    count.value = value
  }

  return { count, doubleCount, increment, setCount }
})

import { ref } from 'vue'
import { fetchWithAuth } from '@/composables/fetchAuth'

export function useFetch() {
  const data = ref(null)
  const isLoading = ref(false)
  const error = ref(null)
  const res = ref(null)
  const baseUrl = import.meta.env.VITE_API_BASE_URL

  const execute = async (url) => {
    isLoading.value = true
    error.value = null
    
    try {
      const response = await fetchWithAuth(`${baseUrl}${url}`, { method: 'GET' }) 
      res.value = response;
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => null); 
        error.value = errorData || { message: 'Network response was not ok' }
        return; 
      }

      // Handle success
      data.value = await response.json()
      
    } catch (err) {
      // This catch block will now only run for actual network failures 
      // (like being offline or CORS issues)
      error.value = { message: err.message }
    } finally {
      isLoading.value = false
    }
  }

  return { data, isLoading, error, res, execute }
}
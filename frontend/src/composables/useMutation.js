// src/composables/useMutation.js
import { ref } from 'vue'
import { fetchWithAuth } from '@/composables/fetchAuth'

export function useMutation() {
  const data = ref(null)
  const isLoading = ref(false)
  const error = ref(null)
  const res = ref(null)
  const baseUrl = import.meta.env.VITE_API_BASE_URL

  // ⚡ Added payloadType parameter (defaults to 'json')
  const execute = async (url, payload, method = 'POST', payloadType = 'json') => {
    isLoading.value = true
    error.value = null

    try {
      // ⚡ Only stringify if the type is json, otherwise pass the raw FormData object
      const requestBody = payloadType === 'json' ? JSON.stringify(payload) : payload;

      const response = await fetchWithAuth(`${baseUrl}${url}`, {
        method: method,
        body: requestBody
      })
      res.value = response;

      if (!response.ok) {
        const errorData = await response.json().catch(() => "Error")
        error.value = errorData
        return false 
      }

      // Handle successful response
      const responseText = await response.text()
      data.value = responseText ? JSON.parse(responseText) : null
      return true

    } catch (err) {
      error.value = { message: err.message }
      return false
    } finally {
      isLoading.value = false
    }
  }

  return { data, isLoading, error, res, execute }
}
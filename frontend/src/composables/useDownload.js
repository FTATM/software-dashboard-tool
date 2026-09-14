// src/composables/useDownload.js
import { ref } from 'vue'
import { fetchWithAuth } from '@/composables/fetchAuth'

export function useDownload() {
  const isDownloading = ref(false)
  const error = ref(null)
  const baseUrl = import.meta.env.VITE_API_BASE_URL || ''

  const executeDownload = async (url, filename) => {
    isDownloading.value = true
    error.value = null

    try {
      const response = await fetchWithAuth(`${baseUrl}${url}`, {
        method: 'GET'
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => null)
        error.value = errorData || { message: 'Failed to download file' }
        return false
      }

      // ⚡ Convert securely to a binary Blob and trigger browser download
      const blob = await response.blob()
      const downloadUrl = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = downloadUrl
      link.download = filename
      
      document.body.appendChild(link)
      link.click()
      
      // Clean up memory
      document.body.removeChild(link)
      window.URL.revokeObjectURL(downloadUrl)

      return true

    } catch (err) {
      error.value = { message: err.message }
      return false
    } finally {
      isDownloading.value = false
    }
  }

  return { isDownloading, error, executeDownload }
}
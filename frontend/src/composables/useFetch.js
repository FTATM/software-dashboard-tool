import { ref } from 'vue'
import { fetchWithAuth } from '@/composables/fetchAuth'
import { i18n } from '@/i18n'; // Wrapped in curly braces

export function useFetch() {
  const data = ref(null);
  const isLoading = ref(false);
  const error = ref(null);
  const res = ref(null);
  const baseUrl = import.meta.env.VITE_API_BASE_URL;

  const execute = async (url) => {
    isLoading.value = true;
    error.value = null;
    
    try {
      const response = await fetchWithAuth(`${baseUrl}${url}`, { method: 'GET' });
      res.value = response;
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => null); 

        // Handle 401 Unauthorized using i18n.global.t
        if (response.status === 401) {
          error.value = errorData?.message 
            ? errorData 
            : { message: i18n.global.t('login.sessionExpired') };
          return;
        }

        error.value = errorData || { message: 'Network response was not ok' };
        return; 
      }

      data.value = await response.json();
      
    } catch (err) {
      error.value = { message: err.message };
    } finally {
      isLoading.value = false;
    }
  }

  return { data, isLoading, error, res, execute };
}
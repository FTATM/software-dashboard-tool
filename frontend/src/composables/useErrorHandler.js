import { useI18n } from 'vue-i18n';

export function useErrorHandler() {
  const { t } = useI18n();

  /**
   * Handles API errors, translates them, and returns the message string.
   * 
   * @param {Object|String} error - The error object ref or string (e.g., userLoginError)
   * @param {String} fallbackKey - The i18n key for the default error (e.g., 'common.messages.loadError')
   * @param {Object} params - Dynamic parameters for the translation (e.g., { item: 'User' })
   */
  const handleError = (error, fallbackKey, params = {}) => {
    const backendMsg = error?.value?.message || error?.message || error;
    const displayItem = params.item || '';

    switch (backendMsg) {
      case 't_dup':
        return t('common.messages.duplicateError', { item: displayItem });

      case 't_no_access':
        return t('common.messages.noAccessError');

      case 't_delete_in_used':
        return t('common.messages.deleteFailedInUsed', { item: displayItem });

      case 't_invalid_body':
        return t('common.messages.invalidBody');

      case 't_not_active':
        return t('common.messages.inActive');

      default:
        // Returns the raw backend string if present, otherwise the translated fallback key
        return backendMsg || t(fallbackKey, params);
    }
  };

  return { handleError };
}
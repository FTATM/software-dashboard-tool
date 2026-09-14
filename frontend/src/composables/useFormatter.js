import { useI18n } from 'vue-i18n';

export function useFormatter() {
  const { locale } = useI18n();

  const formatTime = (ts) => {
    if (!ts) return '-';
    const date = new Date(ts);
    
    const currentLocale = locale.value === 'th' ? 'th-TH' : 'en-GB';
    
    return date.toLocaleString(currentLocale, {
      calendar: 'gregory',
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false
    });
  };

  return {
    formatTime
  };
}
import { defineStore } from 'pinia';
import { ref, watch } from 'vue';

export const useThemeStore = defineStore('theme', () => {
  // สร้างตัวแปรเก็บสถานะ Theme (ค่าเริ่มต้นเป็น false = light theme)
  const isDarkTheme = ref(false);

  // ฟังก์ชันสำหรับสลับ Theme
  const toggleTheme = () => {
    isDarkTheme.value = !isDarkTheme.value;
  };

  // ใช้ watch เพื่อคอยดูการเปลี่ยนแปลง เมื่อค่าเปลี่ยนให้อัปเดต HTML ทันที
  watch(isDarkTheme, (isDark) => {
    const themeName = isDark ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', themeName);
  }, { immediate: true }); 
  // immediate: true ทำให้ระบบอัปเดตแท็ก HTML ทันทีเมื่อเปิดแอปพลิเคชัน 
  // (ดึงค่าที่ persist ไว้มาใช้ต่อได้เลย)

  return {
    isDarkTheme,
    toggleTheme
  };
}, { persist: true });
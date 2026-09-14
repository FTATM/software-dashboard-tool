import { defineStore } from 'pinia';
import { ref } from 'vue';

export const usePermissionStore = defineStore('permissions', () => {
  // Initialize a global reactive Map[cite: 2]
  const userPermissions = ref(new Map());

  // Renamed the parameter to 'permissionsObj' to be clearer about what it expects[cite: 2]
  const setPermissions = (permissionsObj) => {
    const newPermissionsMap = new Map();
    
    // SAFEGUARD: If the object is null or undefined, stop here.[cite: 2]
    if (!permissionsObj) {
      userPermissions.value = newPermissionsMap;
      return;
    }
    
    // Loop through the backend JSON and convert each array into a Set[cite: 2]
    for (const [menuName, actionsArray] of Object.entries(permissionsObj)) {
      if (Array.isArray(actionsArray)) {
        newPermissionsMap.set(menuName, new Set(actionsArray));
      }
    }
    
    userPermissions.value = newPermissionsMap;
  };

  const hasPermission = (menuName, actionName) => {
    const actionsSet = userPermissions.value.get(menuName);
    
    if (!actionsSet) return false;
    
    return actionsSet.has(actionName);
  };

  return {
    userPermissions,
    setPermissions,
    hasPermission
  };
});
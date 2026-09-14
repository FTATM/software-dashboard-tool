import { defineStore } from 'pinia';
import { ref } from 'vue';

/**
 * Define the shape of your User object.
 * The bracket syntax like [userName] means the property is optional.
 * 
 * @typedef {Object} UserData
 * @property {number} [id]
 * @property {string} [fullName]
 * @property {string} [firstName]
 * @property {string} [lastName]
 */

export const useUserStore = defineStore('user', () => {

  // Apply the type to the ref. 
  // We use import('vue').Ref to tell the IDE it is a reactive Vue reference.
  /** @type {import('vue').Ref<UserData>} */
  const user = ref({});

  // You can also type the parameter of your action
  /**
   * Action to update the user data
   * @param {UserData | null} userData 
   */
  const setUser = (userData) => {
    user.value = userData || {};
  };


  return {
    user,
    setUser,
  };
}, { persist: true });
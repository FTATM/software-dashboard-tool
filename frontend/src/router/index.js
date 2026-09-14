import { createRouter, createWebHistory } from 'vue-router';
import LoginView from '@/views/LoginView.vue';
import DashboardView from '@/views/DashboardView.vue';
import LogReportView from '@/views/LogReportView.vue';
import UserView from '@/views/management/UserView.vue';
import DeviceView from '@/views/management/DeviceView.vue';
import DeviceGroupView from '@/views/management/DeviceGroupView.vue';
import RoleView from '@/views/management/RoleView.vue';
import SchedulerView from '@/views/SchedulerView.vue';
import CanvasView from '@/views/canvasManagement/CanvasView.vue';
import CanvasDesignView from '@/views/canvasManagement/CanvasDesignView.vue';
import CanvasAccessView from '@/views/canvasManagement/CanvasAccessView.vue';
import NotifUserView from '@/views/notification/NotifUserView.vue';
import NotifDeviceRuleView from '@/views/notification/NotifDeviceRuleView.vue';
import { useFetch } from '@/composables/useFetch';
import { toast } from 'vue3-toastify';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { useUserStore } from '@/stores/useUserStore';

// const { data: userPermissionData, error: userPermissionError, execute: userPermissionApi } = useFetch();


const router = createRouter({
  history: createWebHistory(),
  routes: [
    // redirect any unknown routes to dashboard (or login)
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
    {
      path: '/login',
      component: LoginView,
      name: 'login',
      meta: { hideLayout: true }
    },
    {
      path: '/dashboard',
      component: DashboardView,
      name: 'dashboard',
      meta: { requiresAuth: true }
    },
    {
      path: '/logReport',
      name: 'logReport',
      component: LogReportView,
      meta: { requiresAuth: true },
    },
    {
      path: '/canvasManagement',
      name: 'canvasManagement',
      meta: { requiresAuth: true },
      children: [
        {
          path: 'canvas',
          name: 'canvas',
          component: CanvasView,
        },
        {
          path: 'design',
          name: 'canvasDesign',
          component: CanvasDesignView,
        },
        {
          path: 'access',
          name: 'canvasAccess',
          component: CanvasAccessView,
        },
      ]
    },
    {
      path: '/management',
      name: 'management',
      meta: { requiresAuth: true },
      children: [
        {
          path: 'user',
          name: 'user',
          component: UserView,
        },
        {
          path: 'role',
          name: 'role',
          component: RoleView,
        },
        {
          path: 'device',
          name: 'device',
          component: DeviceView,
        },
        {
          path: 'deviceGroup',
          name: 'deviceGroup',
          component: DeviceGroupView,
        },
      ]
    },
    {
      path: '/scheduler',
      name: 'scheduler',
      component: SchedulerView,
      meta: { requiresAuth: true }
    },
    {
      path: '/notification',
      name: 'notification',
      meta: { requiresAuth: true },
      children: [
        {
          path: 'user',
          name: 'notifUser',
          component: NotifUserView,
        },
        {
          path: 'devicerule',
          name: 'notifDevicerule',
          component: NotifDeviceRuleView,
        }
      ]
    }
  ]
});

// GLOBAL ROUTE GUARD
router.beforeEach(async (to, from) => {
  const userStore = useUserStore(); //[cite: 1]
  const isLoggedIn = !!userStore.user?.id; //[cite: 1]

  // 1. Guest Guard: Prevent logged-in users from seeing the login page
  if (to.name === 'login' && isLoggedIn) { //[cite: 1]
    return { name: 'dashboard' }; //[cite: 1]
  }

  // 2. Auth Guard: Check if the route requires authentication
  if (to.meta.requiresAuth) { //[cite: 1]
    if (!isLoggedIn) { //[cite: 1]
      return { name: 'login' }; //[cite: 1]
    }

    const permissionStore = usePermissionStore(); //[cite: 1]

    // Destructure 'res' from useFetch to inspect status codes
    const {
      data: userPermissionData, //[cite: 1, 3]
      error: userPermissionError, //[cite: 1, 3]
      res: userPermissionRes, //
      execute: userPermissionApi //[cite: 1, 3]
    } = useFetch(); //[cite: 1, 3]

    await userPermissionApi('/user/permission'); //[cite: 1]

    // Case 1: Permissions retrieved successfully
    if (!userPermissionError.value && userPermissionData.value) { //[cite: 1]
      permissionStore.setPermissions(userPermissionData.value.data); //[cite: 1]
      return true; //[cite: 1]
    }

    // Case 2: Server Down / Network Offline (No response or 5xx)
    const isServerDown = !userPermissionRes.value || userPermissionRes.value.status >= 500;

    if (isServerDown) {
      // Clear permissions so hasPermission(...) returns false -> displays <NoAccess />
      permissionStore.setPermissions([]);

      // Stay on the same page. Do NOT touch userStore and do NOT redirect to login.
      return true;
    }

    // Case 3: Token Expired / Unauthorized (401, 403)
    // (If fetchWithAuth has not already redirected via window.location.href)
    toast.error(userPermissionError.value?.message || "Session expired. Please login again."); //[cite: 1]
    userStore.setUser({}); //
    localStorage.removeItem('user');
    permissionStore.setPermissions([]);

    return { name: 'login' }; //[cite: 1]
  }

  // 3. Public Routes
  return true; //[cite: 1]
});

export default router;
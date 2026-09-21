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
import LineRegisterView from '@/views/line/LineRegisterView.vue';
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
    },
    {
      path: '/liff/register',
      name: 'lineRegister',
      component: LineRegisterView,
      meta: {
        hideLayout: true,  // Bypasses MainLayout sidebar and navbar
        isLiffOnly: true   // Custom flag to guard against desktop web browsers
      }
    }
  ]
});

// GLOBAL ROUTE GUARD
router.beforeEach(async (to, from) => {
  const userStore = useUserStore();
  const isLoggedIn = !!userStore.user?.id;

  // 1. Guest Guard
  if (to.name === 'login' && isLoggedIn) {
    return { name: 'dashboard' };
  }

  // 2. Auth Guard
  if (to.meta.requiresAuth) {
    if (!isLoggedIn) {
      return { name: 'login' };
    }

    const permissionStore = usePermissionStore();
    const {
      data: userPermissionData,
      error: userPermissionError,
      res: userPermissionRes,
      execute: userPermissionApi
    } = useFetch();

    // Fetches on every route change; backend serves from cache
    await userPermissionApi('/user/permission');

    // Case 1: Permissions retrieved successfully
    if (!userPermissionError.value && userPermissionData.value?.data) {
      permissionStore.setPermissions(userPermissionData.value.data);
      return true;
    }

    // Case 2: Backend Offline or 5xx error -> preserve user session
    const isServerDown = !userPermissionRes.value || userPermissionRes.value.status >= 500;
    if (isServerDown) {
      permissionStore.setPermissions([]);
      return true;
    }

    // Case 3: Unauthorized (401) -> cookie invalid/missing
    if (userPermissionRes.value?.status === 401) {
      toast.error(userPermissionError.value?.message || "Session expired. Please login again.");
      userStore.setUser({});
      permissionStore.setPermissions([]);
      return { name: 'login' };
    }

    return true;
  }

  return true;
});

export default router;
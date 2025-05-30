import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/components/layout/Layout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/Login.vue'),
      meta: {
        title: '登录',
        hidden: true
      }
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/auth/Register.vue'),
      meta: {
        title: '注册',
        hidden: true
      }
    },
    {
      path: '/404',
      name: 'NotFound',
      component: () => import('@/views/auth/NotFound.vue'),
      meta: {
        title: '404',
        hidden: true
      }
    },
    {
      path: '/',
      component: Layout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/Index.vue'),
          meta: {
            title: '首页',
            icon: 'House'
          }
        }
      ]
    },
    {
      path: '/task',
      component: Layout,
      meta: {
        title: '任务管理',
        icon: 'List'
      },
      children: [
        {
          path: 'list',
          name: 'TaskList',
          component: () => import('@/views/task/List.vue'),
          meta: {
            title: '任务列表',
            icon: 'Document'
          }
        },
        {
          path: 'create',
          name: 'TaskCreate',
          component: () => import('@/views/task/Edit.vue'),
          meta: {
            title: '创建任务',
            icon: 'Plus',
            hidden: true
          }
        },
        {
          path: 'edit/:id',
          name: 'TaskEdit',
          component: () => import('@/views/task/Edit.vue'),
          meta: {
            title: '编辑任务',
            icon: 'Edit',
            hidden: true
          }
        }
      ]
    },
    {
      path: '/record',
      component: Layout,
      meta: {
        title: '执行记录',
        icon: 'Histogram'
      },
      children: [
        {
          path: 'list',
          name: 'RecordList',
          component: () => import('@/views/record/List.vue'),
          meta: {
            title: '记录列表',
            icon: 'Tickets'
          }
        },
        {
          path: 'detail/:id',
          name: 'RecordDetail',
          component: () => import('@/views/record/Detail.vue'),
          meta: {
            title: '记录详情',
            icon: 'InfoFilled',
            hidden: true
          }
        }
      ]
    },
    {
      path: '/department',
      component: Layout,
      meta: {
        title: '部门管理',
        icon: 'OfficeBuilding'
      },
      children: [
        {
          path: 'list',
          name: 'DepartmentList',
          component: () => import('@/views/department/List.vue'),
          meta: {
            title: '部门列表',
            icon: 'Sell'
          }
        }
      ]
    },
    {
      path: '/user',
      component: Layout,
      meta: {
        title: '用户管理',
        icon: 'User'
      },
      children: [
        {
          path: 'list',
          name: 'UserList',
          component: () => import('@/views/user/List.vue'),
          meta: {
            title: '用户列表',
            icon: 'Avatar'
          }
        }
      ]
    },
    {
      path: '/role',
      component: Layout,
      meta: {
        title: '角色管理',
        icon: 'Lock'
      },
      children: [
        {
          path: 'list',
          name: 'RoleList',
          component: () => import('@/views/role/List.vue'),
          meta: {
            title: '角色列表',
            icon: 'Key'
          }
        }
      ]
    },
    // Path not found - 404
    {
      path: '/:pathMatch(.*)*',
      redirect: '/404',
      meta: {
        hidden: true
      }
    }
  ]
})

// Global navigation guard to check authentication
router.beforeEach((to, _, next) => {
  // Set page title
  if (to.meta.title) {
    document.title = `${to.meta.title} - DistributedJob`
  }
  // Check if needs auth
  const token = localStorage.getItem('token')
  // 登录和注册页面不需要认证
  if (to.path === '/login' || to.path === '/register') {
    if (token) {
      next({ path: '/' })
    } else {
      next()
    }
  } else {
    if (token) {
      next()
    } else {
      next({ path: '/login', query: { redirect: to.fullPath } })
    }
  }
})

export default router
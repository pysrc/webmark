import { createRouter, createWebHistory } from 'vue-router'
import Login from '@/views/Login.vue'
import System from '@/views/System.vue'
import GroupDetail from '@/views/GroupDetail.vue'
import DocumentDetail from '@/views/DocumentDetail.vue'
import DocumentEdit from '@/views/DocumentEdit.vue'
import DocumentNew from '@/views/DocumentNew.vue'


const routes = [
  { path: '/', redirect: '/login' },
  { path: '/login', name: 'login', component: Login },
  { path: '/system', name: 'system', component: System },
  { path: '/group/:groupname', component: GroupDetail },
  { path: '/document/:groupname/:mdname', component: DocumentDetail },
  { path: '/edit/:groupname/:mdname', component: DocumentEdit },
  { path: '/newdocument/:groupname/:mdname', component: DocumentNew}
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

/* 全局前置守卫：防止没登录就进后台 */
router.beforeEach((to, from, next) => {
  const isLoggedIn = localStorage.getItem('isLoggedIn') === 'true'
  if (to.meta.requiresAuth && !isLoggedIn) {
    next('/login')
  } else if(to.name === 'login' && isLoggedIn) {
    next("/system")
  } else {
    next()
  }
})

export default router

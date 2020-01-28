import Vue from 'vue';
import Router from 'vue-router';

import HomePage from '../home/HomePage'
import LoginPage from '../login/LoginPage'
import RegisterPage from '../register/RegisterPage'
import UserProfilePage from '../userProfile/userProfile'
import VotePage from '../userProfile/participatePoll'
import CreateVotePage from '../userProfile/createPoll'
import ViewResult from '../userProfile/viewResult'


Vue.use(Router);

export const router = new Router({
  mode: 'history',
  routes: [
    // { path: '/', component: HomePage },
    { path: '/login', component: LoginPage },
    { path: '/register', component: RegisterPage },
    { path: '/users', component: UserProfilePage},
    { path: '/users/:id', component: UserProfilePage},
    { name: 'vote', path: '/vote', component: VotePage, props: true },
    { path: '/create', component: CreateVotePage},
    { name: 'result', path: '/result', component: ViewResult, props: true},

    // otherwise redirect to home
    { path: '*', redirect: '/login' }
  ]
});

router.beforeEach((to, from, next) => {
  // redirect to login page if not logged in and trying to access a restricted page
  const publicPages = ['/login', '/register', '/users', '/result'];
  const authRequired = !publicPages.includes(to.path);
  const loggedIn = localStorage.getItem('user');

  if (authRequired && !loggedIn) {
    return next('/login');
  }

  next();
})
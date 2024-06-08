import {createRouter, createWebHistory} from 'vue-router'
import HomePage from "@/pages/HomePage.vue";
import TestPage from "@/pages/TestPage.vue";
import PathNotFound from "@/pages/PathNotFound.vue";

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            name: 'home',
            component: HomePage
        },
        {
            path: '/test',
            name: 'test',
            component: TestPage
        },
        {
            path: '/:pathMatch(.*)*',
            name: "not found",
            component: PathNotFound
        }
    ]
})

export default router

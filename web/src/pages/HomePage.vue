<script setup>
//
//import TestItem from "@/components/TestItem.vue";
//import {useCounterStore} from "@/stores/counter.js";
//import {onMounted, ref} from "vue";
//import {getGroupUrl} from "@/constants/index.js";
//import axios from "axios";
// import ScheduleItem from "@/components/ScheduleItem.vue";
// import Department from "@/components/Departments.vue";
//
//
//const useSchedule = () => {
//  const schedule = ref(null)
//  const getSchedule = () => axios.get(getGroupUrl("344")).then((response) => {
//    schedule.value = response.data
//    console.log(schedule)
//  })
//  return {schedule, getSchedule}
//}
//
//
//const {schedule, getSchedule} = useSchedule()
//
//// const schedule = ref(null)
//// const getSchedule = () => axios.get(getGroupUrl("344")).then((response) => {
////   schedule.value = response.data
////   console.log(schedule)
//// })
//// getSchedule()
//
//
//
//console.log(schedule.value)
//
//const counter = useCounterStore()
//
//const test = ref(0)
//
//test.value = counter.count
//
//counter.$subscribe(
//    (mutation, state) => {
//      test.value = state.count
//    }
//)
//
//onMounted(() => {
//  getSchedule()
//})
//
//
//const input = (event) => {
//  console.log(event)
//  // test.value =
//  counter.setCount(Number(event.target.value))
//}

import {onMounted, ref} from "vue";
import Expresions from "@/components/Expresions.vue";
import axios from "axios";
import {API_URL} from "@/constants.js";

const searchVal = ref("")

const sendExpression = async () => {
  try {
    const response = await axios.post(`${API_URL}/api/v1/calculate`, {"expression": searchVal.value});
    const resp = response.data;
    console.log(resp)
  } catch (error) {
    console.error('Ошибка при загрузке schedule:', error);
  }
}

const inputEnter = async (event) => {
  searchVal.value = event.target.value
  if (event.key === "Enter") {
    console.log(searchVal.value)

    await sendExpression()

    searchVal.value = ""
    await getExpressions()
  }
}

const expressions = ref({})

const getExpressions = async () => {
  try {
    const response = await axios.get(`${API_URL}/api/v1/expressions`);
    expressions.value = response.data;
    console.log(expressions)
  } catch (error) {
    console.error('Ошибка при загрузке schedule:', error);
  }
};
// setTimeout(async () => {await getExpressions()}, 1)
// setTimeout(()=>console.log("timeoiy"), 1)

onMounted(async () => {
  console.log("onMounted")

  await getExpressions()
  // setInterval(async () => {await getExpressions()}, 1000)


})


</script>

<template>
  <body>
  <header><h1>Distributed arithmetic expression calculator</h1></header>
  <main>
    <div class="content">
      <div class="container">
        <input :value="searchVal" placeholder="Введите выражение" @keydown="inputEnter($event)">
      </div>
    </div>
    <!--    <div class="content">-->
    <!--      <button @click="getExpressions"> Обновить</button>-->
    <!--    </div>-->
    <div class="content">
      <Expresions :expressions="expressions"></Expresions>
    </div>
  </main>
<!--  <footer>Footer</footer>-->
  </body>
</template>

<style lang="scss" scoped>

header {
  padding: 10px;
  background: #4d9efe;
}

//
//.wrapper {
//  display: flex;
//  justify-content: center;
//  align-items: center;
//}

input {
  height: 40px;
  width: 300px;
  border-color: #4d9efe;
  padding: 5px;
  font-size: 20px;
}

.container {
  display: flex;
  margin: 0 auto;
  margin: 10px;
}

.title {
  color: blueviolet;
  text-align: center;
}
</style>